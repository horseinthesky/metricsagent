package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	config, err := ParseConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.Address, defaultListenOn)
	assert.Equal(t, config.StoreFile, defaultStoreFile)
}
