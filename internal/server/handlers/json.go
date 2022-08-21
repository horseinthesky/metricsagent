package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

func extractMetic(r *http.Request) (*storage.Metric, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	metric := &storage.Metric{}

	err = json.Unmarshal(body, metric)
	if err != nil {
		return nil, err
	}

	return metric, nil
}

func HandleSaveJSONMetric(db storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric := &storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(metric)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		if storage.UnsupportedType(metric.MType) {
			http.Error(w, "Unknown type", http.StatusNotImplemented)
			return
		}

		err = db.Set(metric)
		if err != nil {
			http.Error(w, "Invalid value", http.StatusBadRequest)
			return
		}
	})
}
func HandleSaveJSONMetric2(db storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric, err := extractMetic(r)
		if err != nil {
			http.Error(w, "bad or no payload", http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		if storage.UnsupportedType(metric.MType) {
			http.Error(w, "Unknown type", http.StatusNotImplemented)
			return
		}

		err = db.Set(metric)
		if err != nil {
			http.Error(w, "Invalid value", http.StatusBadRequest)
			return
		}
	})
}

func HandleLoadJSONMetric(db storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricRequest, err := extractMetic(r)
		if err != nil {
			http.Error(w, "failed to unmarshal payload", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		if storage.UnsupportedType(metricRequest.MType) {
			http.Error(w, "Unknown type", http.StatusNotImplemented)
			return
		}

		metric, err := db.Get(metricRequest.ID)
		if err != nil {
			http.Error(w, "Unknown metric id", http.StatusNotFound)
			return
		}

		res, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, "Internal error during JSON marshal", http.StatusInternalServerError)
		}

		w.Write([]byte(res))
	})
}
