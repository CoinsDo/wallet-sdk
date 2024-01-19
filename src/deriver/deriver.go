package deriver

import "wallet-sdk/src/types"

//hd wallet derivation

type Deriver interface {

	// Derive Derive a private key based on the specified path
	Derive(path string) (types.PrivateKey, error)

	// Initialize Initialize the root private key using a mnemonic phrase
	Initialize(mnemonicStr string) error
}
