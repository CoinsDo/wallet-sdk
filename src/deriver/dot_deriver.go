package deriver

import (
	"github.com/vedhavyas/go-subkey"
	"github.com/vedhavyas/go-subkey/sr25519"
	"wallet-sdk/src/types"
)

// for dot ksm

type DotDeriver struct {
	mnemonic string
}

func (deriver *DotDeriver) Initialize(mnemonicStr string) error {
	deriver.mnemonic = mnemonicStr
	return nil
}

func (deriver *DotDeriver) Derive(path string) (types.PrivateKey, error) {
	var str = deriver.mnemonic + path
	bytes := []byte(str)
	scheme := sr25519.Scheme{}
	kyr, err := subkey.DeriveKeyPair(scheme, string(bytes))
	if err != nil {
		return nil, err
	}
	return kyr.Seed(), nil
}
