package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	Metrics map[string]float64
}

var storage = MemStorage{
	Metrics: make(map[string]float64),
}

type Storage interface {
	GetMetric(name string) (float64, bool)
	UpdateCounter(name, value string) error
	UpdateGauge(name, value string) error
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
		err := storage.UpdateCounter(nameMetric, valueMetric)
		if err != nil {
			http.Error(w, "Invalid value metric", http.StatusBadRequest)
		}
	case "gauge":
		err := storage.UpdateGauge(nameMetric, valueMetric)
		if err != nil {
			http.Error(w, "Invalid value metric", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Invalid type metric", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

}

func (m *MemStorage) GetMetric(name string) (float64, bool) {
	metric, exist := m.Metrics[name]
	return metric, exist
}

func (m *MemStorage) UpdateCounter(name, value string) error {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	metric, exist := storage.GetMetric(name)
	if exist {
		metric += float64(v)
	}
	storage.Metrics[name] = metric
	return nil
}

func (m *MemStorage) UpdateGauge(name, value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	storage.Metrics[name] = v

	return nil
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)
	mux.HandleFunc("/update/", update)

	fmt.Println(http.ListenAndServe(":8080", mux))
}
