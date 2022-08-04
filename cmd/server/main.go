package main

	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var gaugeStorage = map[string]float64{}
var counterStorage = map[string]int64{}

func splitPath(s, sep string) (string, string, string, string) {
	x := strings.Split(s, sep)
	return x[1], x[2], x[3], x[4]
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/update/" {
	//         http.NotFound(w, r)
	//         return
	// }
	switch r.Method {
	//       case "GET":
	//               for k, v := range r.URL.Query() {
	//                       fmt.Printf("%s: %s\n", k, v)
	//               }
	//               w.Write([]byte("Received a GET request\n"))
	case "GET":
		_, metricType, metricName, valueString := splitPath(r.URL.Path, "/")
		if metricType == "gauge" {
			value, _ := strconv.ParseFloat(valueString, 64)
			gaugeStorage[metricName] = value
		}

		if metricType == "counter" {
			value, _ := strconv.ParseInt(valueString, 10, 64)

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
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}

}

func main() {
	http.HandleFunc("/update/", metricsHandler)
	http.ListenAndServe(":8080", nil)
}
