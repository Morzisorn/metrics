package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"go.uber.org/zap"
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

var (
	fileInstance *FileStorage
	onceFile     sync.Once
)

func GetFileStorage() *FileStorage {
	onceFile.Do(func() {
		var err error
		service := config.GetService("server")
		fileInstance, err = NewFileStorage(service.Config.FileStoragePath)
		if err != nil {
			logger.Log.Panic("Error file loading storage", zap.Error(err))
		}
	})
	return fileInstance
}

func NewFileStorageProducer(filename string) (*FileStorageProducer, error) {
	return &FileStorageProducer{
		filename: filename,
	}, nil
}

type FileStorageConsumer struct {
	file    io.ReadCloser
	decoder *json.Decoder
}

func NewFileStorageConsumer(filename string) (*FileStorageConsumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &FileStorageConsumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func NewFileStorage(filepath string) (*FileStorage, error) {
	producer, err := NewFileStorageProducer(filepath)
	if err != nil {
		return nil, err
	}

	consumer, err := NewFileStorageConsumer(filepath)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		Producer: producer,
		Consumer: consumer,
	}, nil
}

func (p *FileStorageProducer) WriteMetric(name string, value float64) error {
	metric := StorageMetric{
		Name:  name,
		Value: value,
	}
	return p.encoder.Encode(metric)
}

func (c *FileStorageConsumer) ReadMetric() (*StorageMetric, error) {
	var metric StorageMetric
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
