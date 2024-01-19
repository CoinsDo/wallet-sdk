package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyHrc20 = "HRC20"

var coinHrc20 Hrc20

func init() {
	coinHrc20 = Hrc20{}
	RegisterEthLikeCoin(coinHrc20)
}

type Hrc20 struct {
	Ht
}

func (coin Hrc20) GetCurrency() string {
	return CurrencyHrc20
}

func (coin Hrc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin Hrc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
