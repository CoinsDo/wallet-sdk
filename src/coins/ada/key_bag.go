package ada

import (
	"encoding/hex"
	"github.com/echovl/cardano-go/crypto"
	"golang.org/x/crypto/blake2b"
	"wallet-sdk/src/crypto/cip1852"
)

type keyBag struct {
	pkeys map[string]crypto.ExtendedSigningKey
}

func NewKeyBag() keyBag {

	return keyBag{
		pkeys: make(map[string]crypto.ExtendedSigningKey),
	}
}

func (keyBag keyBag) AddKey(key cip1852.HdPrivateKey) {
	pkeyHashBytes := blake2b.Sum256(key.KeyData)
	pkeyHashString := hex.EncodeToString(pkeyHashBytes[:])
	keyBag.pkeys[pkeyHashString] = key.KeyData
}

func (keyBag keyBag) Sign(tx *Transaction) {
	witnessSet := transactionWitnessSet{}
	for _, pkey := range keyBag.pkeys {
		txHash := blake2b.Sum256(tx.Body.Bytes())
		publicKey := pkey.ExtendedVerificationKey()[:32]
		signature := pkey.Sign(txHash[:])
		witness := vkeyWitness{VKey: publicKey, Signature: signature}

		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}
	tx.WitnessSet = witnessSet

}
