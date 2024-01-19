package types

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type BigInt struct {
	big.Int
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *BigInt) UnmarshalJSON(p []byte) error {

	all := strings.ReplaceAll(string(p), "\"", "")
	if all == "null" || all == "" {
		return nil
	}
	n := new(big.Int)
	_, b2 := n.SetString(all, 10)
	if !b2 {
		return fmt.Errorf("not a valid big integer: %s", p)
	}
	b.Int = *n
	return nil
}

func ParseBigintFromString(valueStr string) (*BigInt, error) {
	bigInt := BigInt{}
	n := big.Int{}
	_, b2 := n.SetString(valueStr, 10)
	if !b2 {
		return nil, errors.New(fmt.Sprintf("not a valid big integer: %s", valueStr))
	}
	bigInt.Int = n
	return &bigInt, nil
}

type TxParams interface {
	GetParams() interface{}
}

type BaseTxParams struct {
}

func (params BaseTxParams) GetParams() interface{} {
	return params
}

type PrivateKey []byte

type Path string

type BaseTransaction struct {
	CoinTransaction interface{}
}

type CoinAddress struct {
	AddressStr string `json:"addressStr"`
	PrivateKey string `json:"privateKey"`
	Path       Path   `json:"path"`
	Index      int64  `json:"index"`
	Currency   string `json:"currency"`
}
