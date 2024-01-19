package deriver

import (
	"bytes"
	"crypto/ed25519"
	ed25519hd "github.com/portto/solana-go-sdk/pkg/hdwallet"
	"github.com/tyler-smith/go-bip39"
	"wallet-sdk/src/types"
)

// for sol near

type Ed25519Deriver struct {
	seed []byte
}

func (deriver *Ed25519Deriver) Initialize(mnemonicStr string) error {
	seed := bip39.NewSeed(mnemonicStr, "")
	deriver.seed = seed
	return nil
}

func (deriver *Ed25519Deriver) Derive(path string) (types.PrivateKey, error) {
	derivedKey, err := ed25519hd.Derived(path, deriver.seed)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(derivedKey.PrivateKey)
	_, privKey, err := ed25519.GenerateKey(reader)
	if err != nil {
		return nil, err
	}
	var keyData = make([]byte, len(privKey))
	copy(keyData, privKey[:])
	return keyData, nil
}
