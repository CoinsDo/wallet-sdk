package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyMaticErc20 = "MATIC_ERC20"

var coinMaticErc20 MaticErc20

func init() {
	coinMaticErc20 = MaticErc20{}
	RegisterEthLikeCoin(coinMaticErc20)
}

type MaticErc20 struct {
	Matic
}

func (coin MaticErc20) GetCurrency() string {
	return CurrencyMaticErc20
}

func (coin MaticErc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin MaticErc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
