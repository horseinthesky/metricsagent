package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

func HandleSaveJSONMetric(db storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		metric := &storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(metric)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if storage.UnsupportedType(metric.MType) {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusNotImplemented)
			return
		}

		err = db.Set(metric)
		if err != nil {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusBadRequest)
			return
		}

		w.Write([]byte(`{"result": "metric saved"}`))
	})
}

func HandleLoadJSONMetric(db storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		metricRequest := &storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(metricRequest)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if storage.UnsupportedType(metricRequest.MType) {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusNotImplemented)
			return
		}

		metric, err := db.Get(metricRequest.ID)
		if err != nil {
			http.Error(w, `{"result": "unknown metric id"}`, http.StatusNotFound)
			return
		}

		res, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, `{"error": "faied to marshal metric"}`, http.StatusInternalServerError)
			return
		}

		w.Write([]byte(res))
	})
}
