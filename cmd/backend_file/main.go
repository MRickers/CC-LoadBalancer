package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("cmd/backend_file")))
	http.ListenAndServe(":8082", nil)
}
