package file

import (
	"encoding/json"
	"io"
	"os"
)

type FileStorage struct {
	Producer *FileStorageProducer
	Consumer *FileStorageConsumer
}

type FileStorageProducer struct {
	filename string
	file     io.WriteCloser
	encoder  *json.Encoder
}

type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func NewStorage(filepath string) (*FileStorage, error) {
	cons, err := NewFileStorageConsumer(filepath)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		Producer: NewFileStorageProducer(filepath),
		Consumer: cons,
	}, nil
}

func NewFileStorageProducer(filepath string) *FileStorageProducer {
	return &FileStorageProducer{filename: filepath}
}

type FileStorageConsumer struct {
	file    io.ReadCloser
	decoder *json.Decoder
}

func NewFileStorageConsumer(filepath string) (*FileStorageConsumer, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &FileStorageConsumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (p *FileStorageProducer) WriteMetric(name string, value float64) error {
	metric := Metric{
		Name:  name,
		Value: value,
	}
	return p.encoder.Encode(metric)
}

func (c *FileStorageConsumer) ReadMetric() (*Metric, error) {
	var metric Metric
	err := c.decoder.Decode(&metric)
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func (c *FileStorageConsumer) Close() error {
	return c.file.Close()
}

func (p *FileStorageProducer) Close() error {
	return p.file.Close()
}

func (f *FileStorage) Close() error {
	err := f.Producer.Close()
	if err != nil {
		return err
	}
	return f.Consumer.Close()
}

func (p *FileStorageProducer) WriteMetrics(metrics *map[string]float64) error {
	var err error
	p.file, err = os.OpenFile(p.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer p.file.Close()

	p.encoder = json.NewEncoder(p.file)
	for name, value := range *metrics {
		err := p.WriteMetric(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *FileStorageConsumer) ReadMetrics() (*map[string]float64, error) {
	var metrics = make(map[string]float64)
	for {
		metric, err := c.ReadMetric()
		if err != nil {
			break
		}
		metrics[metric.Name] = metric.Value
	}
	return &metrics, nil
}
