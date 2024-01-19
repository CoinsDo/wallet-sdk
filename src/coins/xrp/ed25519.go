package xrpCrypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"github.com/rubblelabs/ripple/crypto"
)

type Ed25519key struct {
	PrivateKey ed25519.PrivateKey
}

func checkSequenceIsNil(seq *uint32) {
	if seq != nil {
		panic("Ed25519 keys do not support account families")
	}
}

func (e *Ed25519key) Id(seq *uint32) []byte {
	checkSequenceIsNil(seq)
	return crypto.Sha256RipeMD160(e.Public(seq))
}

func (e *Ed25519key) Public(seq *uint32) []byte {
	checkSequenceIsNil(seq)
	return append([]byte{0xED}, e.PrivateKey[32:]...)
}

func (e *Ed25519key) Private(seq *uint32) []byte {
	checkSequenceIsNil(seq)
	return e.PrivateKey[:]
}

func NewEd25519Key(seed []byte) (*Ed25519key, error) {
	r := rand.Reader
	if seed != nil {
		r = bytes.NewReader(crypto.Sha512Half(seed))
	}
	_, priv, err := ed25519.GenerateKey(r)
	if err != nil {
		return nil, err
	}
	return &Ed25519key{PrivateKey: priv}, nil
}
