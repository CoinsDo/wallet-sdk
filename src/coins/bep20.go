package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyBep20 = "BEP20"

var coinBep20 Bep20

func init() {
	coinBep20 = Bep20{}
	RegisterEthLikeCoin(coinBep20)
}

type Bep20 struct {
	Bnb
}

func (coin Bep20) GetCurrency() string {
	return CurrencyBep20
}

func (coin Bep20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin Bep20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
