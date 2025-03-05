package models

type StorageMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Storage interface {
	GetMetric(name string) (float64, bool)
	GetMetrics() (*map[string]float64, error)
	UpdateCounter(name string, value float64) (float64, error)
	UpdateGauge(name string, value float64) error
	WriteMetrics(*map[string]float64) error
	UpdateCounters(*map[string]float64) error
	UpdateGauges(*map[string]float64) error
}
