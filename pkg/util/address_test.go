package util_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/piplabs/story-staking-api/pkg/util"
)

func TestCmpPubKeyToEVMAddress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cmpPubKeyBase64 := "AqBVHHkyOfiie29Wrez6hMvC644kbZfPgXA1jFEs7Uwq"

		cmpPubKey, err := base64.StdEncoding.DecodeString(cmpPubKeyBase64)
		require.NoError(t, err)

		evmAddr, err := util.CmpPubKeyToEVMAddress(cmpPubKey)
		require.NoError(t, err)
		require.Equal(t, "0xC5c0BEEAC8B37eD52F6A675eE2154D926a88E3ec", evmAddr.Hex())
	})

	t.Run("invalid compressed public key length", func(t *testing.T) {
		cmpPubKeyBase64 := "AqBVHHkyOfiie29Wrez6hMvC644kbZfPgXA1jFEs7Uwq"

		cmpPubKey, err := base64.StdEncoding.DecodeString(cmpPubKeyBase64)
		require.NoError(t, err)

		cmpPubKey = cmpPubKey[:len(cmpPubKey)-1]

		_, err = util.CmpPubKeyToEVMAddress(cmpPubKey)
		require.Error(t, err)
	})

	t.Run("invalid prefix", func(t *testing.T) {
		cmpPubKeyBase64 := "AqBVHHkyOfiie29Wrez6hMvC644kbZfPgXA1jFEs7Uwq"

		cmpPubKey, err := base64.StdEncoding.DecodeString(cmpPubKeyBase64)
		require.NoError(t, err)

		cmpPubKey[0] = 0x01

		_, err = util.CmpPubKeyToEVMAddress(cmpPubKey)
		require.Error(t, err)
	})
}
