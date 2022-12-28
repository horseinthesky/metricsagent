package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// addHash adds hash to metric.
// Only used if hash key is provided.
func addHash(metric Metric, hashKey string) Metric {
	var data string

	h := hmac.New(sha256.New, []byte(hashKey))

	switch metric.MType {
	case "gauge":
		data = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
	case "counter":
		data = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
	}

	h.Write([]byte(data))
	metric.Hash = hex.EncodeToString(h.Sum(nil))

	return metric
}
