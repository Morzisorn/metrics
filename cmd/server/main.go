package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/morzisorn/metrics/internal/storage"
)

var s = storage.MemStorage{
	Metrics: make(map[string]float64),
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
	}

	typeMetric := splitPath[2]
	nameMetric := splitPath[3]
	valueMetric := splitPath[4]

	if nameMetric == "" {
		http.Error(w, "Invalid name or value metric", http.StatusBadRequest)
	}

	switch typeMetric {
	case "counter":
		err := s.UpdateCounter(nameMetric, valueMetric)
		if err != nil {
			http.Error(w, "Invalid value metric", http.StatusBadRequest)
		}
	case "gauge":
		err := s.UpdateGauge(nameMetric, valueMetric)
		if err != nil {
			http.Error(w, "Invalid value metric", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Invalid type metric", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)
	mux.HandleFunc("/update/", update)

	fmt.Println(http.ListenAndServe(":8080", mux))
}
