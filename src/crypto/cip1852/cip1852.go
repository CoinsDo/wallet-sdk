package cip1852

import "github.com/tyler-smith/go-bip39"

func GetKeyPairFromMnemonic(mnemonic string, path DerivationPath) (*HdKeyPair, error) {
	fromMnemonic, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	entropy := GetKeyPaiFromEntropy(fromMnemonic, path)
	return &entropy, nil
}

func GetKeyPaiFromEntropy(entropy []byte, path DerivationPath) HdKeyPair {
	rootKey := NewRootKey(entropy)

	indexKey := rootKey.DeriveChild(path.Purpose.Value, path.Purpose.IsHarden).
		DeriveChild(path.CoinType.Value, path.CoinType.IsHarden).
		DeriveChild(path.Account.Value, path.Account.IsHarden).
		DeriveChild(path.Role.Value, path.Role.IsHarden).
		DeriveChild(path.Index.Value, path.Index.IsHarden)

	return indexKey

}

func DeriveByKeyPair(keypair HdKeyPair, path DerivationPath) HdKeyPair {
	indexKey := keypair.DeriveChild(path.Purpose.Value, path.Purpose.IsHarden).
		DeriveChild(path.CoinType.Value, path.CoinType.IsHarden).
		DeriveChild(path.Account.Value, path.Account.IsHarden).
		DeriveChild(path.Role.Value, path.Role.IsHarden).
		DeriveChild(path.Index.Value, path.Index.IsHarden)

	return indexKey

}
