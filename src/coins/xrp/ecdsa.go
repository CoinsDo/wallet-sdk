package xrpCrypto

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/rubblelabs/ripple/crypto"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
)

var (
	order = btcec.S256().N
	zero  = big.NewInt(0)
	one   = big.NewInt(1)
)

type EcdsaKey struct {
	*btcec.PrivateKey
}

func NewECDSAKeyFromBytes(data []byte) (*EcdsaKey, error) {

	privKeyFromBytes, _ := btcec.PrivKeyFromBytes(data)

	return &EcdsaKey{privKeyFromBytes}, nil
}

func (k *EcdsaKey) Id(sequence *uint32) []byte {
	if sequence == nil {
		return crypto.Sha256RipeMD160(k.PubKey().SerializeCompressed())
	}
	return crypto.Sha256RipeMD160(k.Public(sequence))
}

func (k *EcdsaKey) Private(sequence *uint32) []byte {
	if sequence == nil {
		return k.ToECDSA().D.Bytes()
	}
	b := k.generateKey(*sequence).Key.Bytes()
	return b[:]
}

func (k *EcdsaKey) Public(sequence *uint32) []byte {
	if sequence == nil {
		return k.PubKey().SerializeCompressed()
	}
	return k.generateKey(*sequence).PubKey().SerializeCompressed()
}

func (k *EcdsaKey) generateKey(sequence uint32) *btcec.PrivateKey {
	seed := make([]byte, btcec.PubKeyBytesLenCompressed+4)
	copy(seed, k.PubKey().SerializeCompressed())
	binary.BigEndian.PutUint32(seed[btcec.PubKeyBytesLenCompressed:], sequence)
	key := newKey(seed)
	key.Key = *key.Key.Add(&k.Key)
	return &btcec.PrivateKey{
		Key: key.Key,
	}
}

func newKey(seed []byte) *btcec.PrivateKey {
	inc := big.NewInt(0).SetBytes(seed)
	inc.Lsh(inc, 32)
	for key := big.NewInt(0); ; inc.Add(inc, one) {
		key.SetBytes(crypto.Sha512Half(inc.Bytes()))
		if key.Cmp(zero) > 0 && key.Cmp(order) < 0 {
			privKey, _ := btcec.PrivKeyFromBytes(key.Bytes())
			return privKey
		}
	}
}

// If seed is nil, generate a random one
func NewECDSAKey(seed []byte) (*EcdsaKey, error) {
	if seed == nil {
		seed = make([]byte, 16)
		if _, err := rand.Read(seed); err != nil {
			return nil, err
		}
	}
	return &EcdsaKey{newKey(seed)}, nil
}
