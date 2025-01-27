package main

import (
	"fmt"
	"net/http"

	server "github.com/morzisorn/metrics/internal/server/handlers"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)
	mux.HandleFunc("/update/", server.Update)

	fmt.Println(http.ListenAndServe(":8080", mux))
}
