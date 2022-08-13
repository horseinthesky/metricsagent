package main

import (
	"log"
	"net/http"

	"github.com/horseinthesky/metricsagent/internal/server"
)

const (
	listenOn = ":8080"
)

func main() {
	metricsServer := server.New()
	log.Fatal(http.ListenAndServe(listenOn, metricsServer))
}
