package deriver

import (
	"bytes"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ed25519"
	cip1852 "wallet-sdk/src/crypto/cip1852"
	"wallet-sdk/src/types"
)

//for ALGO

type AlgoDeriver struct {
	rootKey cip1852.HdKeyPair
}

func (deriver *AlgoDeriver) Initialize(mnemonicStr string) error {
	entropy, err := bip39.EntropyFromMnemonic(mnemonicStr)
	if err != nil {
		return err
	}
	keyPair := cip1852.NewRootKey(entropy)
	deriver.rootKey = keyPair
	return nil
}

func (deriver *AlgoDeriver) Derive(path string) (types.PrivateKey, error) {

	derivationPath, err := cip1852.CreateFromPath(path)
	if err != nil {
		return nil, err
	}

	data := cip1852.DeriveByKeyPair(deriver.rootKey, *derivationPath).PrivateKey.KeyData
	reader := bytes.NewReader(data)
	_, sk, err := ed25519.GenerateKey(reader)
	if err != nil {
		return nil, err
	}
	var keyData = make([]byte, len(sk))
	copy(keyData, sk[:])
	return keyData, nil
}
