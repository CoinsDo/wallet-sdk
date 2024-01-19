package coins

import (
	"encoding/json"
	"wallet-sdk/src/types"
)

const CurrencyErc1155 = "ERC1155"

type Erc1155TxParams struct {
	EthTxParams
	ContractAddress string             `json:"commonContractAddress"`
	FromAddress     string             `json:"fromAddress"`
	BatchData       []Erc1155BatchData `json:"erc1155BatchData"`
}

type Erc1155BatchData struct {
	TokenId types.BigInt `json:"tokenId"`
	Amount  types.BigInt `json:"amount"`
}

var coinErc1155 Erc1155

func init() {
	coinErc1155 = Erc1155{}
	RegisterEthLikeCoin(coinErc1155)
	RegisterNftToken(coinErc1155)
}

type Erc1155 struct {
	Eth
}

func (coin Erc1155) GetCurrency() string {
	return CurrencyErc1155
}

func (coin Erc1155) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createErc1155TokenTransaction(params)
}

func (coin Erc1155) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Erc1155TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
