package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleSaveTextMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	valueString := chi.URLParam(r, "value")

	err := stash.Set(metricName, valueString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	w.Write([]byte("Received a POST request\n"))
}

func HandleLoadTextMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")

	value, err := stash.Get(metricName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	w.Write([]byte(fmt.Sprint(value)))
}
