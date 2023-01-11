package server

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	curDir, err := os.Getwd()
	assert.NoError(t, err)

	testAddress := "localhost:8082"
	os.Setenv("CONFIG", curDir+"/testdata/server_config.json")
	os.Setenv("ADDRESS", testAddress)
	os.Setenv("TRUSTED_SUBNET", "0.0.0.0/0")

	config, err := ParseConfig()

	assert.NoError(t, err)
	assert.Equal(t, testAddress, config.Address)
	assert.Equal(t, 100*time.Second, config.StoreInterval)
	assert.Equal(t, "", config.DatabaseDSN)
}
