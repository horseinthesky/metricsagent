package agent

import (
	"fmt"
	"log"
)

func Example() {
	// Start agent
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse agent config: %w", err))
	}

	agent, err := NewAgent(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create agent: %w", err))
	}

	agent.Run()
}
