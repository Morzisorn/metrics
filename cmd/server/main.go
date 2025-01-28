package main

import (
	"fmt"
	"net/http"

	server "github.com/morzisorn/metrics/internal/server/handlers"
)

func startServer() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", server.Update)
	return mux
}

func main() {
	mux := startServer()
	fmt.Println(http.ListenAndServe(":8080", mux))
}
