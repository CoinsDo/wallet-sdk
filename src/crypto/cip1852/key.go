package cip1852

import (
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/echovl/ed25519"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/pbkdf2"
	"strconv"
	"strings"
)

type HdKey struct {
	Version     []byte
	Depth       int
	ChildNumber []byte
	ChainCode   []byte
	KeyData     []byte
}

type HdPrivateKey HdKey

func (key HdPrivateKey) ToHex() string {
	bytes := make([]byte, len(key.KeyData)+len(key.ChainCode))
	copy(bytes, key.KeyData)
	copy(bytes[len(key.KeyData):], key.ChainCode)
	return hexutil.Encode(bytes)
}

type HdPublicKey HdKey

type HdKeyPair struct {
	PrivateKey HdPrivateKey
	PublicKey  HdPublicKey
	Path       string
}

func PrivateKeyFromHex(privateKey string) (key *HdPrivateKey, error error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			key = nil
			error = errors.New(fmt.Sprintf("%+v\n", err))
		}
	}()

	privateKey = strings.TrimPrefix(privateKey, "0x")

	bytes := common.Hex2Bytes(privateKey)

	keyData := make([]byte, 64)
	chainCode := make([]byte, 32)

	copy(keyData, bytes[:64])
	copy(chainCode, bytes[64:])

	hdPrivateKey := HdPrivateKey{
		ChainCode: chainCode,
		KeyData:   keyData,
	}
	return &hdPrivateKey, nil
}

func PrivateKeyFromBytes(key []byte) HdPrivateKey {

	keyData := make([]byte, 64)
	chainCode := make([]byte, 32)

	copy(keyData, key[:64])
	copy(chainCode, key[64:])

	hdPrivateKey := HdPrivateKey{
		ChainCode: chainCode,
		KeyData:   keyData,
	}
	return hdPrivateKey
}

func (pubKey HdPublicKey) GetKeyHash() ([]byte, error) {
	data := pubKey.KeyData
	out, err2 := BlakeTo224(data)
	if err2 != nil {
		return nil, err2
	}
	return out, nil
}

func PubKeyFromBytes(data []byte) (*HdPublicKey, error) {
	if data == nil || len(data) != 64 {
		return nil, errors.New("Invalid key length. Key length should be 64")
	}
	keyBytes := make([]byte, 32)
	chainCode := make([]byte, 32)
	copy(keyBytes, data[:32])
	copy(chainCode, data[32:])
	key := HdPublicKey{
		ChainCode: chainCode,
		KeyData:   keyBytes,
	}
	return &key, nil
}

func (privatekey HdPrivateKey) GetPublicKey() HdPublicKey {
	from := ed25519.PublicKeyFrom(privatekey.KeyData)
	key := HdPublicKey{
		ChainCode: privatekey.ChainCode,
		KeyData:   from,
	}
	return key
}

func NewRootKey(entropy []byte) HdKeyPair {
	xprv := pbkdf2.Key([]byte(""), entropy, 4096, 96, sha512.New)
	xprv[0] &= 0xf8
	xprv[31] = (xprv[31] & 0x1f) | 0x40

	var IL = make([]byte, 64)
	var IR = make([]byte, 32)

	copy(IL, xprv[:64])
	copy(IR, xprv[64:])
	pk := ed25519.PublicKeyFrom(IL)
	var publicKey = HdPublicKey{
		ChainCode: IR,
		KeyData:   pk,
	}

	var privateKey = HdPrivateKey{
		ChainCode: IR,
		KeyData:   IL,
	}

	return HdKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

}

func (parent HdKeyPair) DeriveChild(child uint64, isHarden bool) HdKeyPair {
	if isHarden {
		child += 0x80000000
	}
	chainCode := parent.PrivateKey.ChainCode
	xpriv := parent.PrivateKey.KeyData
	zmac := hmac.New(sha512.New, chainCode)
	ccmac := hmac.New(sha512.New, chainCode)

	sindex := serializeIndex(child)
	if isHarden {
		zmac.Write([]byte{0x0})
		zmac.Write(xpriv)
		zmac.Write(sindex)
		ccmac.Write([]byte{0x1})
		ccmac.Write(xpriv)
		ccmac.Write(sindex)
	} else {
		pub := ed25519.PublicKeyFrom(xpriv)
		zmac.Write([]byte{0x2})
		zmac.Write(pub)
		zmac.Write(sindex)
		ccmac.Write([]byte{0x3})
		ccmac.Write(pub)
		ccmac.Write(sindex)
	}
	z := zmac.Sum(nil)
	zl := z[:32]
	zr := z[32:64]
	kl := add28Mul8(xpriv[:32], zl)
	kr := addMod256(xpriv[32:], zr)

	cc := ccmac.Sum(nil)
	cc = cc[32:]

	cxsk := make([]byte, 96)
	copy(cxsk[:32], kl)
	copy(cxsk[32:64], kr)
	copy(cxsk[64:], cc)
	privateKey := HdPrivateKey{
		ChainCode: cc,
		KeyData:   cxsk[:64],
	}

	pk := ed25519.PublicKeyFrom(cxsk[:64])

	publicKey := HdPublicKey{
		ChainCode: cc,
		KeyData:   pk,
	}

	return HdKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Path:       getPath(parent.Path, child, isHarden),
	}

}

const MASTER_PATH = "m"

func serializeIndex(index uint64) []byte {
	return []byte{byte(index), byte(index >> 8), byte(index >> 16), byte(index >> 24)}
}
func getPath(parentPath string, child uint64, isHardened bool) string {

	if parentPath == "" {
		parentPath = MASTER_PATH
	}
	var hardenSymble string
	if isHardened {
		hardenSymble = "'"
	} else {
		hardenSymble = ""
	}
	return parentPath + "/" + strconv.FormatInt(int64(child), 10) + hardenSymble
}

func add28Mul8(x, y []byte) []byte {
	out := make([]byte, 32)
	var carry uint16

	for i, xi := range x[:28] {
		r := uint16(xi) + ((uint16(y[i])) << 3) + carry
		out[i] = byte(r & 0xff)
		carry = r >> 8
	}
	for i, xi := range x[28:32] {
		r := uint16(xi) + carry
		out[i+28] = byte(r & 0xff)
		carry = r >> 8
	}

	return out
}

func addMod256(x, y []byte) []byte {
	out := make([]byte, 32)
	var carry uint16

	for i, xi := range x[:32] {
		r := uint16(xi) + uint16(y[i]) + carry
		out[i] = byte(r)
		carry = r >> 8

	}

	return out
}

func BlakeTo224(data []byte) ([]byte, error) {
	hash, err := blake2b.New(28, nil)
	if err != nil {
		return nil, err
	}
	hash.Write(data)
	sum := hash.Sum(nil)
	return sum, nil
}
