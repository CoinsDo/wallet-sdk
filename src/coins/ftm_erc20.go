package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyFtmErc20 = "FTM_ERC20"

var coinFmterc20 FtmErc20

func init() {
	coinFmterc20 = FtmErc20{}
	RegisterEthLikeCoin(coinFmterc20)
}

type FtmErc20 struct {
	Ftm
}

func (coin FtmErc20) GetCurrency() string {
	return CurrencyFtmErc20
}

func (coin FtmErc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)
}

func (coin FtmErc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
