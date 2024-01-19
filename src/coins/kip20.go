package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyKip20 = "KIP20"

var coinKip20 Kip20

func init() {
	coinKip20 = Kip20{}
	RegisterEthLikeCoin(coinKip20)
}

type Kip20 struct {
	Okt
}

func (coin Kip20) GetCurrency() string {
	return CurrencyKip20
}

func (coin Kip20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin Kip20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
