package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyErc721 = "ERC721"

type Erc721 struct {
	Eth
}

type Erc721TxParams struct {
	EthTxParams
	ContractAddress string       `json:"commonContractAddress"`
	FromAddress     string       `json:"fromAddress"`
	TokenId         types.BigInt `json:"ethereumTokenId"`
}

var coinErc721 Erc721

func init() {
	coinErc721 = Erc721{}
	RegisterEthLikeCoin(coinErc721)
	RegisterNftToken(coinErc721)
}

func (coin Erc721) GetCurrency() string {
	return CurrencyErc721
}

func (coin Erc721) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createErc721TokenTransaction(params)
}

func (coin Erc721) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc721TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
