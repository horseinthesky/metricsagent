package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/horseinthesky/metricsagent/internal/server"
)

const (
	listenOn = "localhost:8080"
)

var (
	address string
)

func init() {
	address = os.Getenv("ADDRESS")
	if address == "" {
		address = listenOn
	}
}

func main() {
	metricsServer := server.New()
	log.Fatal(fmt.Errorf("server crashed due to %w", http.ListenAndServe(address, metricsServer)))
}
