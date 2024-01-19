package coins

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"github.com/shopspring/decimal"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const (
	trc20TransferMethodSignature = "0xa9059cbb"
	CurrencyTrc20                = "TRC20"
)

type Trc20TxParams struct {
	TrxTxParams
	ContractAddress string `json:"commonContractAddress"`

	TokenDecimal int64 `json:"commonTokenDecimal"`
}

var coinTrc20 Trc20

func init() {
	coinTrc20 = Trc20{}

	RegisterCoin(coinTrc20)
}

type Trc20 struct {
	Trx
}

func (coin Trc20) GetCurrency() string {
	return CurrencyTrc20
}

func (coin Trc20) GetEmptyTransactionParams() types.TxParams {
	return Trc20TxParams{}
}

func (coin Trc20) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {

	extraParams := params.(Trc20TxParams)
	var tokenDecimal = extraParams.TokenDecimal
	var contractAddress = extraParams.ContractAddress
	transferContract, err := coin.CreateContract(extraParams.FromAddress, extraParams.ToAddress, extraParams.Amount, contractAddress, tokenDecimal)
	if err != nil {
		return nil, err
	}
	return coin.constructTx(transferContract, extraParams.TrxTxParams)

}

func (coin Trc20) CreateContract(fromAddress string, toAddress string, amount decimal.Decimal, contractAddress string, tokenDecimal int64) (proto.Message, error) {
	transferContract := core.TriggerSmartContract{}

	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}

	transferContract.OwnerAddress = fromAddressBytes
	toAddressbytes, err := common.DecodeCheck(contractAddress)
	if err != nil {
		return nil, errors.ErrorInvalidContractAddress
	}
	transferContract.ContractAddress = toAddressbytes

	addrB, err := address.Base58ToAddress(toAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	finalAmount := amountToMinUnit(amount, int32(tokenDecimal))
	ab := common.LeftPadBytes(finalAmount.Bytes(), 32)
	req := trc20TransferMethodSignature + "0000000000000000000000000000000000000000000000000000000000000000"[len(addrB.Hex())-4:] + addrB.Hex()[4:]
	req += common.Bytes2Hex(ab)
	dataBytes, err := common.FromHex(req)
	if err != nil {
		return nil, err
	}
	transferContract.Data = dataBytes
	return &transferContract, nil
}

func (coin Trc20) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := Trc20TxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Trc20) CalculateTxLength(txParams Trc20TxParams, privateKey types.PrivateKey) (int64, error) {
	baseTransaction, err := coin.CreateTransaction(txParams, true)
	if err != nil {
		return 0, err
	}
	tx := baseTransaction.CoinTransaction.(*core.Transaction)
	rawData, err := proto.Marshal(tx.GetRawData())
	if err != nil {
		return 0, err
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return 0, err
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	signature, err := crypto.Sign(hash[:], privateKeyECDSA)
	return int64(len(signature) + len(rawData) + 69), nil
}
