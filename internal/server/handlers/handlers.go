package server

import (
	"net/http"
	"strings"

	"github.com/morzisorn/metrics/internal/server/storage"
)

func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 5 {
		http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}
	method := splitPath[1]
	if method != "update" {
		http.Error(w, "Invalid method", http.StatusNotFound)
		return
	}

	nameMetric := splitPath[3]
	if nameMetric == "" {
		http.Error(w, "Invalid name metric", http.StatusNotFound)
		return
	}

	valueMetric := splitPath[4]
	if valueMetric == "" {
		http.Error(w, "Invalid value metric", http.StatusNotFound)
		return
	}

	typeMetric := splitPath[2]
	switch typeMetric {
	case "counter":
		s := storage.GetStorage()
		err := s.UpdateCounter(nameMetric, valueMetric)
		if err != nil {
			http.Error(w, "Invalid value metric", http.StatusBadRequest)
			return
		}
	case "gauge":
		s := storage.GetStorage()
		err := s.UpdateGauge(nameMetric, valueMetric)
		if err != nil {
			http.Error(w, "Invalid value metric", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Invalid type metric", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
