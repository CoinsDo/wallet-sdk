package deriver

import (
	ed25519hd "github.com/portto/solana-go-sdk/pkg/hdwallet"
	"github.com/tyler-smith/go-bip39"
	"wallet-sdk/src/types"
)

// for stellar

type StellarDeriver struct {
	seed []byte
}

func (deriver *StellarDeriver) Initialize(mnemonicStr string) error {
	seed := bip39.NewSeed(mnemonicStr, "")
	deriver.seed = seed
	return nil
}

func (deriver *StellarDeriver) Derive(path string) (types.PrivateKey, error) {
	derivedKey, err := ed25519hd.Derived(path, deriver.seed)
	if err != nil {
		return nil, err
	}
	return derivedKey.PrivateKey, nil
}
