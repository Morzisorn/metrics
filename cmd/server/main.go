package main

import (
	"fmt"
	"net/http"

	server "github.com/morzisorn/metrics/internal/server/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", server.Update)

	fmt.Println(http.ListenAndServe(":8080", mux))
}
