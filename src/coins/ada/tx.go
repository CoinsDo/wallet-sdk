package ada

import (
	"encoding/hex"
	"github.com/shopspring/decimal"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

type ProtocolParams struct {
	MinimumUtxoValue uint64
	PoolDeposit      uint64
	KeyDeposit       uint64
	MinFeeA          uint64
	MinFeeB          uint64
}

type TransactionID string

func (id TransactionID) Bytes() []byte {

	bytes, err := hex.DecodeString(strings.TrimPrefix(string(id), "0x"))
	if err != nil {
		panic(err)
	}

	return bytes
}

type Transaction struct {
	_          struct{} `cbor:",toarray"`
	Body       transactionBody
	WitnessSet transactionWitnessSet
	Metadata   *transactionMetadata // or null
}

func (tx *Transaction) Bytes() []byte {
	bytes, err := cbor.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (tx *Transaction) CborHex() string {
	return hex.EncodeToString(tx.Bytes())
}

func (tx *Transaction) ID() TransactionID {
	txHash := blake2b.Sum256(tx.Body.Bytes())
	return TransactionID(hex.EncodeToString(txHash[:]))
}

type transactionWitnessSet struct {
	VKeyWitnessSet []vkeyWitness `cbor:"0,keyasint,omitempty"`
}

type vkeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte   // ed25519 public key
	Signature []byte   // ed25519 signature
}

// Cbor map
type transactionMetadata map[uint64]transactionMetadatum

// This could be cbor map, array, int, bytes or a text
type transactionMetadatum struct{}

type transactionBody struct {
	Inputs       []transactionInput  `cbor:"0,keyasint"`
	Outputs      []transactionOutput `cbor:"1,keyasint"`
	Fee          uint64              `cbor:"2,keyasint"`
	Ttl          uint64              `cbor:"3,keyasint"`
	Certificates []certificate       `cbor:"4,keyasint,omitempty"` // Omit for now
	Withdrawals  *uint               `cbor:"5,keyasint,omitempty"` // Omit for now
	Update       *uint               `cbor:"6,keyasint,omitempty"` // Omit for now
	MetadataHash *uint               `cbor:"7,keyasint,omitempty"` // Omit for now
}

func (body *transactionBody) Bytes() []byte {
	bytes, err := cbor.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes
}

type transactionInput struct {
	_     struct{} `cbor:",toarray"`
	ID    []byte   // HashKey 32 bytes
	Index uint64
}

type transactionOutput struct {
	_       struct{} `cbor:",toarray"`
	Address []byte
	Amount  uint64
}

type certificate struct{}

type HumanReadableTransaction struct {
	Id         string                `json:"id"`
	WitnessSet transactionWitnessSet `json:"witnessSet"`
	Metadata   *transactionMetadata  `json:"metadata"`
	Tx         TxBody                `json:"tx"`
}

type TxBody struct {
	Inputs       []Input         `json:"inputs"`
	Outputs      []Output        `json:"outputs"`
	Fee          decimal.Decimal `json:"fee"`
	Ttl          uint64          `json:"ttl"`
	Certificates []certificate   `json:"certificates"`
	Withdrawals  *uint           `json:"withdrawals"`
	Update       *uint           `json:"update"`
	MetadataHash *uint           `json:"metadata_hash"`
}

type Input struct {
	ID    string `json:"id"`
	Index uint64 `json:"index"`
}

type Output struct {
	TxHash  string          `json:"txHash"`
	Address string          `json:"address"`
	Amount  decimal.Decimal `json:"amount"`
	Index   uint64          `json:"index"`
}
