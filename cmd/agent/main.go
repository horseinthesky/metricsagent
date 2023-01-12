package main

import (
	"fmt"
	"log"

	"github.com/horseinthesky/metricsagent/internal/agent"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// Start agent
	cfg, err := agent.ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse agent config: %w", err))
	}

	// Log build info
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	if cfg.GRPC {
		grpcAgent, err := agent.NewGRPCAgent(cfg)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create HTTP agent: %w", err))
		}

		grpcAgent.Run()
	} else {
		agent, err := agent.NewAgent(cfg)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create gRPC agent: %w", err))
		}

		agent.Run()
	}
}
