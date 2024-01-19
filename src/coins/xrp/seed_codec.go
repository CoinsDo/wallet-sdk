package xrpCrypto

import (
	"errors"
	"github.com/rubblelabs/ripple/crypto"
)

type Version []byte

var ED25519_SEED = Version{0x01, 0xE1, 0x4B}
var FAMILY_SEED = Version{0x21}

type KeyType uint

const ED25519 = KeyType(0)
const SECP256K1 = KeyType(1)

type Decoded struct {
	Version Version
	Bytes   []byte
	Type    KeyType
}

func (decoded Decoded) DeriveKey() (crypto.Key, error) {

	keyType := decoded.Type
	switch keyType {
	case ED25519:
		return NewEd25519Key(decoded.Bytes)
	case SECP256K1:
		return NewECDSAKey(decoded.Bytes)
	default:

	}

	return nil, nil
}

func DecodeSeed(base58EncodedSecret string) (*Decoded, error) {
	base58Decode, err := crypto.Base58Decode(base58EncodedSecret, crypto.ALPHABET)
	if err != nil {
		return nil, err
	}

	var decodedDataWithoutChecksum = make([]byte, len(base58Decode)-4)
	copy(decodedDataWithoutChecksum, base58Decode[0:len(base58Decode)-4])

	expectLen := uint32(16)
	decoded, err := decodeBytes(decodedDataWithoutChecksum, []KeyType{ED25519, SECP256K1}, []Version{ED25519_SEED, FAMILY_SEED}, &expectLen)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func decodeBytes(base58Decode []byte, types []KeyType, versions []Version, expectLength *uint32) (*Decoded, error) {
	var versionLengthGuess = len(versions[0])
	var payloadLength uint32
	if expectLength != nil {
		payloadLength = *expectLength
	} else {
		payloadLength = uint32(len(base58Decode) - versionLengthGuess)
	}
	var versionBytes = make([]byte, len(base58Decode)-int(payloadLength))
	copy(versionBytes, base58Decode[:len(base58Decode)-int(payloadLength)])

	var payload = make([]byte, int(payloadLength))
	copy(payload, base58Decode[len(base58Decode)-int(payloadLength):])

	for i, version := range versions {
		versionEqual := byteArrayEquay(versionBytes, version)

		if versionEqual {
			var keyType KeyType
			if i < len(types) {
				keyType = types[i]
			}

			decoded := Decoded{
				Version: version,
				Bytes:   payload,
				Type:    keyType,
			}
			return &decoded, nil
		}
	}
	return nil, errors.New("Version is invalid. Version bytes do not match any of the provided versions.")

}

func byteArrayEquay(arr1, arr2 []byte) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i, b := range arr1 {
		if b != arr2[i] {
			return false
		}
	}

	return true
}
