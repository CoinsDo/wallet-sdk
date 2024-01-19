package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyXdaiErc20 = "XDAI_ERC20"

var coinXdaiErc20 XdaiErc20

func init() {
	coinXdaiErc20 = XdaiErc20{}
	RegisterEthLikeCoin(coinXdaiErc20)
}

type XdaiErc20 struct {
	Xdai
}

func (coin XdaiErc20) GetCurrency() string {
	return CurrencyXdaiErc20
}

func (coin XdaiErc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin XdaiErc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
