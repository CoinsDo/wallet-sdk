package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyOptErc20 = "OPT_ERC20"

var coinOpt20 OptErc20

func init() {
	coinOpt20 = OptErc20{}
	RegisterEthLikeCoin(coinOpt20)
}

type OptErc20 struct {
	Opt
}

func (coin OptErc20) GetCurrency() string {
	return CurrencyOptErc20
}

func (coin OptErc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin OptErc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
