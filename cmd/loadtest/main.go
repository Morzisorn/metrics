package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	concurrency             = 5
	requestsPerWorker       = 1000
	sleepBeforeStartSeconds = 3
	targetURL               = "http://localhost:8080/updates"

	client *http.Client
)

type Metric struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

func main() {
	run()
}

func run() {
	time.Sleep(time.Duration(sleepBeforeStartSeconds) * time.Second)
	var wg sync.WaitGroup
	client = &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Println("Starting load test...")

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			baseValue := 1.0
			baseDelta := int64(1)

			for j := 0; j < requestsPerWorker; j++ {
				metrics := initMetrics(baseValue, baseDelta, j)
				err := makeRequest(metrics, workerID, j)
				if err != nil {
					continue
				}
			}
		}(i)
	}

	wg.Wait()
	log.Println("Load test completed.")
}

func initMetrics(baseValue float64, baseDelta int64, j int) []Metric {
	return []Metric{
		{
			ID:    "RandomValue",
			Type:  "gauge",
			Value: floatPtr(baseValue + float64(j)),
		},
		{
			ID:    "PollCount",
			Type:  "counter",
			Delta: intPtr(baseDelta + int64(j)),
		},
		{
			ID:    "Alloc",
			Type:  "gauge",
			Value: floatPtr(baseValue + float64(j) + 0.3),
		},
	}
}

func makeRequest(metrics []Metric, workerID, j int) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Worker %d: JSON marshal error: %v", workerID, err)
		return err
	}

	resp, err := client.Post(targetURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("Worker %d: request %d failed: %v", workerID, j, err)
		return err
	}
	resp.Body.Close()
	return nil
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int64) *int64 {
	return &v
}
