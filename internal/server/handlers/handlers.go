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
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	typeMetric := splitPath[2]
	nameMetric := splitPath[3]
	valueMetric := splitPath[4]

	if nameMetric == "" {
		http.Error(w, "Invalid name metric", http.StatusNotFound)
		return
	}

	if valueMetric == "" {
		http.Error(w, "Invalid value metric", http.StatusBadRequest)
		return
	}

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
