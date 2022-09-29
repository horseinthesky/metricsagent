package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func (a *Agent) addHash(metric *Metric) {
	var data string

	h := hmac.New(sha256.New, []byte(a.key))

	switch metric.MType {
	case "gauge":
		data = fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
	case "counter":
		data = fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType,  *metric.Delta)
	}

	h.Write([]byte(data))
	metric.Hash = hex.EncodeToString(h.Sum(nil))
}
