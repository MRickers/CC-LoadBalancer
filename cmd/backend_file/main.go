package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("cmd/backend_file")))
	http.HandleFunc("/healthCheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.ListenAndServe(":8082", nil)
}
