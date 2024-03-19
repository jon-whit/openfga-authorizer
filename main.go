package main

import (
	"log"
	"net/http"
)

func main() {
	if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
