package util

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func CmpPubKeyToEVMAddress(pubKey []byte) (common.Address, error) {
	if len(pubKey) != secp256k1.PubKeyBytesLenCompressed {
		return common.Address{}, fmt.Errorf("invalid compressed public key length: %d", len(pubKey))
	}

	uncmpPubKey, err := crypto.DecompressPubkey(pubKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to decompress public key: %w", err)
	}

	return crypto.PubkeyToAddress(*uncmpPubKey), nil
}
