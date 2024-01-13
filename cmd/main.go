package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type BackendServers struct {
	backendUrls []string
	mu          sync.Mutex
	urlIndex    int
}

func (b *BackendServers) add(backendServer string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.backendUrls = append(b.backendUrls, backendServer)
}

func (b *BackendServers) remove(backenServer string) {
	contains, index := b.contains(backenServer)
	if contains {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.backendUrls = append(b.backendUrls[:index], b.backendUrls[index+1:]...)
	}
}

func (b *BackendServers) contains(backendServer string) (bool, int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for index, server := range b.backendUrls {
		if server == backendServer {
			return true, index
		}
	}
	return false, 0
}

func (b *BackendServers) nextRoundRobinServer() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.backendUrls) < 1 {
		return "", fmt.Errorf("no backend server available")
	}

	b.urlIndex += 1
	if b.urlIndex >= len(b.backendUrls) {
		b.urlIndex = 0
		return b.backendUrls[len(b.backendUrls)-1], nil
	} else {
		return b.backendUrls[b.urlIndex-1], nil
	}
}

func NewBackendServers() BackendServers {
	return BackendServers{
		backendUrls: []string{"localhost:8081", "localhost:8082"},
		urlIndex:    0,
	}
}

func forwardMessage(data []byte, backendServerUrl string) ([]byte, error) {
	connection, err := net.Dial("tcp", backendServerUrl)
	if err != nil {
		return nil, fmt.Errorf("could not connect to backend server")
	}
	defer connection.Close()

	_, err = connection.Write(data)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	_, err = connection.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func printHttpStatus(httpResponse []byte) error {
	responseString := string(httpResponse)
	firstLineIndex := strings.Index(responseString, "\n")
	if firstLineIndex == -1 {
		return fmt.Errorf("invalid http status response")
	}

	fmt.Println("Response from server: ", responseString[0:firstLineIndex])
	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading: ", err)
		return
	}
	fmt.Println("Received request from: ", conn.RemoteAddr().String())
	fmt.Println(string(buffer))

	backendUrl, err := backendServers.nextRoundRobinServer()
	if err != nil {
		fmt.Println(err)
		conn.Write([]byte(err.Error()))
		return
	}
	response, err := forwardMessage(buffer, backendUrl)

	if err != nil {
		fmt.Println("Forward message failed: ", err)
		return
	}
	err = printHttpStatus(response)
	if err != nil {
		fmt.Println("backend response error: ", err)
		return
	}

	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("could not respond to client")
	}

}

var backendServers = NewBackendServers()
var backendUrls = []string{"localhost:8081", "localhost:8082"}

func main() {
	go func() {
		for {
			fmt.Println("doing health check")
			for _, backend := range backendUrls {
				response, err := http.Get("http://" + backend + "/healthCheck")
				if err != nil {
					contains, _ := backendServers.contains(backend)
					if contains {
						fmt.Println("removing " + backend + " from backendList")
						backendServers.remove(backend)
					}
					continue
				}
				contains, _ := backendServers.contains(backend)
				if response.StatusCode != http.StatusOK {
					if contains {
						fmt.Println("removing " + backend + " from backendList")
						backendServers.remove(backend)
					}
				} else {
					if !contains {
						fmt.Println("adding " + backend + " to backendList")
						backendServers.add(backend)
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		go handleConnection(conn)
	}
}
