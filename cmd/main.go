package main

import (
	"bufio"
	"fmt"
	"net/http"
	"sync/atomic"
)

var backendServers = [2]string{"http://localhost:8081", "http://localhost:8082"}
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

func roundRobinBalancer(forwardHandler func(w http.ResponseWriter, r *http.Request, backendUrl string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backendServer := backendServers[roundRobin()]
		forwardHandler(w, r, backendServer)
	}
}

func forwardHandler(w http.ResponseWriter, r *http.Request, backendUrl string) {
	response := "Received request from " + r.RemoteAddr
	response += "\n" + r.Method + " " + r.RequestURI + " " + r.Proto
	response += "\nHost: " + r.Host
	response += "\nUser-Agent: " + r.UserAgent()
	response += "\nAccept: " + r.Header.Get("Accept") + "\n"

	fmt.Println(response)

	resp, err := http.Get(backendUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response from server: " + resp.Proto + " " + resp.Status + "\n\n")

	backend_response := ""
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		backend_response += scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println(backend_response)

	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(backend_response))

}

func main() {
	http.HandleFunc("/", roundRobinBalancer(forwardHandler))

	http.ListenAndServe(":8080", nil)
}
