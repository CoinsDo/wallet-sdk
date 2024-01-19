package deriver

import (
	"github.com/tyler-smith/go-bip39"
	cip18522 "wallet-sdk/src/crypto/cip1852"
	"wallet-sdk/src/types"
)

//for ADA

type Cip1852Deriver struct {
	rootKey cip18522.HdKeyPair
}

func (deriver *Cip1852Deriver) Initialize(mnemonicStr string) error {
	entropy, err := bip39.EntropyFromMnemonic(mnemonicStr)
	if err != nil {
		return err
	}
	keyPair := cip18522.NewRootKey(entropy)
	deriver.rootKey = keyPair
	return nil
}

func (deriver *Cip1852Deriver) Derive(path string) (types.PrivateKey, error) {

	derivationPath, err := cip18522.CreateFromPath(path)
	if err != nil {
		return nil, err
	}
	bussinessKey := cip18522.DeriveByKeyPair(deriver.rootKey, *derivationPath)
	stekeKeyPath := cip18522.CreateStakeAddressPath(0, 0)
	stakeKey := cip18522.DeriveByKeyPair(deriver.rootKey, stekeKeyPath)
	data := append(bussinessKey.PrivateKey.KeyData, bussinessKey.PrivateKey.ChainCode...)
	data = append(data, stakeKey.PrivateKey.KeyData...)
	data = append(data, stakeKey.PrivateKey.ChainCode...)

	return data, nil
}
