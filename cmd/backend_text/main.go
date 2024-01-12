package main

import (
	"fmt"
	"net/http"
)

func forwardHandler(w http.ResponseWriter, r *http.Request) {
	response := "\nReceived request from " + r.RemoteAddr
	response += "\n" + r.Method + " " + r.RequestURI + " " + r.Proto
	response += "\nHost: " + r.Host
	response += "\nUser-Agent: " + r.UserAgent()
	response += "\nAccept: " + r.Header.Get("Accept")

	fmt.Println(response)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from backend server"))
}

func main() {
	http.HandleFunc("/", forwardHandler)

	http.ListenAndServe(":8081", nil)
}
