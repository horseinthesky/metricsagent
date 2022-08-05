package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/horseinthesky/metricsagent/cmd/server/storage"
)

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/update/" {
	//         http.NotFound(w, r)
	//         return
	// }

	switch r.Method {
	// case "GET":
	// 	for k, v := range r.URL.Query() {
	// 		fmt.Printf("%s: %s\n", k, v)
	// 	}
	// 	w.Write([]byte("Received a GET request\n"))
	case "POST":
		params := strings.Split(r.URL.Path, "/")
		if len(params) < 5 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}
		metricType := params[2]
		metricName := params[3]
		valueString := params[4]

		if metricType != "gauge" && metricType != "counter" {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
			return
		}

		if metricType == "gauge" {
			value, err := strconv.ParseFloat(valueString, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(http.StatusText(http.StatusBadRequest)))
				return
			}
			storage.GaugeStorage[metricName] = value
		}

		if metricType == "counter" {
			value, err := strconv.ParseInt(valueString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(http.StatusText(http.StatusBadRequest)))
				return
			}

			if oldValue, ok := storage.CounterStorage[metricName]; ok {
				newValue := oldValue + value
				storage.CounterStorage[metricName] = newValue
			} else {
				storage.CounterStorage[metricName] = value
			}
		}

		w.Write([]byte("Received a POST request\n"))
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}

}
