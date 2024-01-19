package deriver

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
	"strconv"
	"strings"
	"wallet-sdk/src/types"
)

// for standard bip39 path

type Bip39Deriver struct {
	rootKey hdkeychain.ExtendedKey
}

func (deriver *Bip39Deriver) Initialize(mnemonicStr string) error {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonicStr, "")
	if err != nil {
		return err
	}
	key, err := hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return err
	}
	deriver.rootKey = *key
	return nil

}

func (deriver *Bip39Deriver) Derive(path string) (types.PrivateKey, error) {
	deriveKey, err := deriver.DeriveKey(deriver.rootKey, path)
	if err != nil {
		return nil, err
	}
	ecPrivKey, err := deriveKey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	bytes := ecPrivKey.Serialize()
	return bytes, nil
}

func (deriver Bip39Deriver) DeriveKey(key hdkeychain.ExtendedKey, path string) (*hdkeychain.ExtendedKey, error) {
	split := strings.Split(path, "/")[1:]
	var derivedKey = &key
	for _, ele := range split {
		contains := strings.Contains(ele, "'")
		var id uint32
		if contains {
			str := strings.ReplaceAll(ele, "'", "")
			atoi, err := strconv.ParseInt(str, 10, 32)
			if err != nil {

				return nil, err
			}
			id = hdkeychain.HardenedKeyStart + uint32(atoi)
		} else {
			atoi, err := strconv.ParseInt(ele, 10, 32)
			if err != nil {

				return nil, err
			}
			id = uint32(atoi)
		}

		childKey, err := derivedKey.Derive(id)
		if err != nil {
			return nil, err
		}
		derivedKey = childKey
	}
	return derivedKey, nil
}
