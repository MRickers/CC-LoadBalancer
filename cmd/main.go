package main

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
)

var backendServers = [2]string{"localhost:8081", "localhost:8082"}
var ops atomic.Uint64

func roundRobin() uint64 {
	backendServerIndex := ops.Load()
	fmt.Println("Index: ", backendServerIndex)

	if backendServerIndex >= 2 {
		ops.Store(1)
		return 0
	} else {
		ops.Add(1)
		return backendServerIndex
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

func handleConnection(conn net.Conn, backendServerUrl string) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading: ", err)
		return
	}
	fmt.Println("Received request from: ", conn.RemoteAddr().String())
	fmt.Println(string(buffer))

	response, err := forwardMessage(buffer, backendServerUrl)

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

func main() {

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
		go handleConnection(conn, backendServers[roundRobin()])
	}

}
