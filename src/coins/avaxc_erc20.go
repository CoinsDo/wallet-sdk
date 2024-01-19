package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyAvaxErc20 = "AVAXC_ERC20"

var coinAvaxErc20 AvaxErc20

func init() {
	coinAvaxErc20 = AvaxErc20{}
	RegisterEthLikeCoin(coinAvaxErc20)
}

type AvaxErc20 struct {
	Avaxc
}

func (coin AvaxErc20) GetCurrency() string {
	return CurrencyAvaxErc20
}

func (coin AvaxErc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)

}

func (coin AvaxErc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
