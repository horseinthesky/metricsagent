package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	config, err := ParseConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.Address, defaultAddress)
	assert.Equal(t, config.PollInterval, defaultPollInterval)
}
