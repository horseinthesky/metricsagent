package main

import (
	"net/http"
	"strconv"
	"strings"
)

var gaugeStorage = map[string]float64{}
var counterStorage = map[string]int64{}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/update/" {
	//         http.NotFound(w, r)
	//         return
	// }
	// switch r.Method {
	//       case "GET":
	//               for k, v := range r.URL.Query() {
	//                       fmt.Printf("%s: %s\n", k, v)
	//               }
	//               w.Write([]byte("Received a GET request\n"))
	// case "GET":
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	metricType := params[2]
	metricName := params[3]
	valueString := params[4]
	// _, metricType, metricName, valueString := splitPath(r.URL.Path, "/")
	if metricType == "gauge" {
		// value, err := strconv.ParseFloat("100", 64)
		// fmt.Println(value)
		// fmt.Println(err)
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}
		gaugeStorage[metricName] = value
	}

	if metricType == "counter" {
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}

		if oldValue, ok := counterStorage[metricName]; ok {
			newValue := oldValue + value
			counterStorage[metricName] = newValue
		} else {
			counterStorage[metricName] = value
		}
	}

	// fmt.Println(gaugeStorage)
	// fmt.Println(counterStorage)
	w.Write([]byte("Received a POST request\n"))
	// default:
	// 	w.WriteHeader(http.StatusNotImplemented)
	// 	w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	// }

}

func main() {
	http.HandleFunc("/update/", metricsHandler)
	http.ListenAndServe(":8080", nil)
}
