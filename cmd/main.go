package main

import (
	"bufio"
	"fmt"
	"net/http"
)

func forwardHandler(w http.ResponseWriter, r *http.Request) {
	response := "Received request from " + r.RemoteAddr
	response += "\n" + r.Method + " " + r.RequestURI + " " + r.Proto
	response += "\nHost: " + r.Host
	response += "\nUser-Agent: " + r.UserAgent()
	response += "\nAccept: " + r.Header.Get("Accept") + "\n"

	fmt.Println(response)

	url := "http://localhost:8081"
	resp, err := http.Get(url)
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
	http.HandleFunc("/", forwardHandler)

	http.ListenAndServe(":8080", nil)
}
