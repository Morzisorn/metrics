package file

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/logger"
	"github.com/morzisorn/metrics/internal/server/storage/models"
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
	instanceStorage models.Storage
	onceStorage     sync.Once

	instanceFile *FileStorage
	onceFile     sync.Once
)

func GetStorage() models.Storage {
	onceStorage.Do(func() {
		instanceStorage = GetFileStorage()
	})
	return instanceStorage
}

func GetFileStorage() *FileStorage {
	onceFile.Do(func() {
		var err error
		service := config.GetService("server")
		instanceFile, err = NewFileStorage(service.Config.FileStoragePath)
		if err != nil {
			logger.Log.Panic("Error file loading storage", zap.Error(err))
		}
	})
	return instanceFile
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
	metric := models.StorageMetric{
		Name:  name,
		Value: value,
	}
	return p.encoder.Encode(metric)
}

func (c *FileStorageConsumer) ReadMetric() (*models.StorageMetric, error) {
	var metric models.StorageMetric
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

func (f *FileStorage) GetMetric(name string) (float64, bool) {
	for {
		metric, err := f.Consumer.ReadMetric()
		if err != nil {
			break
		}
		if metric.Name == name {
			return metric.Value, true
		}
	}
	return 0, false
}

func (f *FileStorage) GetMetrics() (*map[string]float64, error) {
	return f.Consumer.ReadMetrics()
}

func (f *FileStorage) UpdateCounter(name string, value float64) (float64, error) {
	return value, f.Producer.WriteMetric(name, value)
}

func (f *FileStorage) UpdateGauge(name string, value float64) error {
	return f.Producer.WriteMetric(name, value)
}

func (f *FileStorage) SetMetrics(metrics *map[string]float64) error {
	return f.Producer.WriteMetrics(metrics)
}
