package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyArb1Erc20 = "ARB1_ERC20"

var coinArb1Erc20 Arb1Erc20

func init() {
	coinArb1Erc20 = Arb1Erc20{}
	RegisterEthLikeCoin(coinArb1Erc20)
}

type Arb1Erc20 struct {
	Arb1
}

func (coin Arb1Erc20) GetCurrency() string {
	return CurrencyArb1Erc20
}

func (coin Arb1Erc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin Arb1Erc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
