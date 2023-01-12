package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

// generateHash adds hash to metric.
// Only used if hash key is provided.
func GenerateHash(metric storage.Metric, hashKey string) []byte {
	hash := hmac.New(sha256.New, []byte(hashKey))

	var data string

	switch metric.MType {
	case storage.Gauge.String():
		data = fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value)
	case storage.Counter.String():
		data = fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta)
	}

	hash.Write([]byte(data))

	return hash.Sum(nil)
}
