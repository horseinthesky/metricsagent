package agent

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	curDir, err := os.Getwd()
	assert.NoError(t, err)

	os.Setenv("CONFIG", curDir + "/testdata/agent_config.json")
	os.Setenv("REPORT_INTERVAL", "50s")

	config, err := ParseConfig()

	assert.NoError(t, err)
	assert.Equal(t, "localhost:8081", config.Address)
	assert.Equal(t, 3 * time.Second, config.PollInterval)
	assert.Equal(t, 50 * time.Second, config.ReportInterval)
}
