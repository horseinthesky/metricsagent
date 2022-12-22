package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	curDir, err := os.Getwd()
	assert.NoError(t, err)

	pubKey, err := ParsePubKey(curDir + "/testdata/public.pem")
	assert.NoError(t, err)

	privKey, err := ParsePrivKey(curDir + "/testdata/private.pem")
	assert.NoError(t, err)

	testMsg := []byte("plaintext")

	encryptedBytes, err := EncryptWithPublicKey(testMsg, pubKey)
	require.NoError(t, err)
	require.NotZero(t, encryptedBytes)

	decryptedBytes, err := DecryptWithPrivateKey(encryptedBytes, privKey)
	require.NoError(t, err)
	require.Equal(t, testMsg, decryptedBytes)
}
