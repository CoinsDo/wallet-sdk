package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyErc20 = "ERC20"

type Erc20TxParams struct {
	EthTxParams
	TokenDecimal    int64  `json:"commonTokenDecimal"`
	ContractAddress string `json:"commonContractAddress"`
}

var coinErc20 Erc20

func init() {
	coinErc20 = Erc20{}
	RegisterEthLikeCoin(coinErc20)
}

type Erc20 struct {
	Eth
}

func (coin Erc20) GetCurrency() string {
	return CurrencyErc20
}

func (coin Erc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTokenTransaction(params)
}

func (coin Erc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
