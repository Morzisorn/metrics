package proto

import (
	"strings"

	"github.com/morzisorn/metrics/internal/models"
	"github.com/morzisorn/metrics/internal/server/logger"
)

func ConvertMetricToPB(m *models.Metric) *Metric {
	p := Metric{
		Id:    m.ID,
		Mtype: stringToPBMType(m.MType),
	}
	if m.Delta == nil && m.Value == nil {
		logger.Log.Error("Both delta and value are nil")
	}

	if m.Delta != nil {
		p.Delta = *m.Delta
	}
	if m.Value != nil {
		p.Value = *m.Value
	}
	return &p
}

func ConvertMetricFromPB(mPB *Metric) *models.Metric {
	return &models.Metric{
		ID:    mPB.Id,
		Delta: &mPB.Delta,
		Value: &mPB.Value,
		MType: strings.ToLower(mPB.GetMtype().String()),
	}
}

func stringToPBMType(s string) Metric_MType {
	switch s {
	case "GAUGE":
		return Metric_GAUGE
	case "COUNTER":
		return Metric_COUNTER
	default:
		return Metric_GAUGE
	}
}
