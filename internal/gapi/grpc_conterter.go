package server

import (
	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

func MetricFromPB(pbMetric *pb.Metric) storage.Metric {
	return storage.Metric{
		ID:    pbMetric.Id,
		MType: pbMetric.Mtype,
		Delta: &pbMetric.Delta,
		Value: &pbMetric.Value,
		Hash:  pbMetric.Hash,
	}
}

func MetricToPB(metric storage.Metric) *pb.Metric {
	pbMetric := &pb.Metric{
		Id:    metric.ID,
		Mtype: metric.MType,
		Hash:  metric.Hash,
	}

	if metric.Delta != nil {
		pbMetric.Delta = *metric.Delta
	}

	if metric.Value != nil {
		pbMetric.Value = *metric.Value
	}

	return pbMetric
}
