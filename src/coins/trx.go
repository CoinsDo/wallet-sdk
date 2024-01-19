package coins

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	addr "github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
	"time"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencyTrx = "TRX"

type TrxTxParams struct {
	types.BaseTxParams
	ToAddress              string                                 `json:"toAddress"`
	Amount                 decimal.Decimal                        `json:"amount"`
	Fee                    decimal.Decimal                        `json:"fee"`
	Memo                   string                                 `json:"memo"`
	FromAddress            string                                 `json:"fromAddress"`
	ContractType           core.Transaction_Contract_ContractType `json:"tronContractType"`
	BlockData              Block                                  `json:"tronBlockData"`
	TronFreezeResourceCode core.ResourceCode                      `json:"tronFreezeResourceCode"`
	TronFrozenDuration     int64                                  `json:"tronFrozenDuration"`
	ExpireTime             int64                                  `json:"ExpireTime"`
}

type Block struct {
	BlockId        string `json:"blockId"`
	BlockNumber    int64  `json:"blockNumber"`
	BlockTimeStamp int64  `json:"blockTimeStamp"`
}

var coinTrx Trx

func init() {
	coinTrx = Trx{}
	RegisterCoin(coinTrx)
}

type Trx struct {
}

func (coin Trx) GetCurrency() string {
	return CurrencyTrx
}

func (coin Trx) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Trx) GetBasePath(testNet bool) string {
	return "m/44'/195'/%d'/%d/%d"
}

func (coin Trx) ChainName() string {
	return CurrencyTrx
}

func (coin Trx) GetDecimal() int {
	return 6
}

func (coin Trx) PrivateKeyToString(key types.PrivateKey) (string, error) {
	return hexutil.Encode(key), nil
}

func (coin Trx) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	privateKey, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (coin Trx) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	address := addr.PubkeyToAddress(privateKey.PublicKey)
	var adddress = types.CoinAddress{}
	adddress.AddressStr = address.String()
	return &adddress, nil

}

func (coin Trx) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Trx) GetEmptyTransactionParams() types.TxParams {
	return TrxTxParams{}
}

func (coin Trx) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := params.(TrxTxParams)

	transferContract, err := coin.CreateContract(extraParams)
	if err != nil {
		return nil, err
	}

	return coin.constructTx(transferContract, extraParams)
}

func (coin Trx) constructTx(transferContract proto.Message, extraParams TrxTxParams) (*types.BaseTransaction, error) {

	var contractType = extraParams.ContractType
	var block = extraParams.BlockData

	tx := core.Transaction{}
	any, err := ptypes.MarshalAny(transferContract)
	if err != nil {
		return nil, err
	}

	contract := core.Transaction_Contract{}
	contract.Type = contractType
	contract.Parameter = any
	transactionRaw := core.TransactionRaw{}
	transactionRaw.Contract = append(transactionRaw.Contract, &contract)
	transactionRaw.Timestamp = time.Now().UnixMilli()
	transactionRaw.Expiration = extraParams.ExpireTime
	transactionRaw.FeeLimit = amountToMinUnit(extraParams.Fee, 6).Int64()
	transactionRaw.Data = []byte(extraParams.Memo)
	if !strings.HasPrefix(block.BlockId, "0x") {
		block.BlockId = "0x" + block.BlockId
	}
	blockIdByte, err := hexutil.Decode(block.BlockId)
	if err != nil {
		return nil, err
	}

	blockHeightByte := make([]byte, 8)

	binary.BigEndian.PutUint64(blockHeightByte, uint64(block.BlockNumber))

	blockHeightByte = blockHeightByte[6:8]

	transactionRaw.RefBlockHash = blockIdByte[8:16]
	transactionRaw.RefBlockNum = block.BlockNumber
	transactionRaw.RefBlockBytes = blockHeightByte

	tx.RawData = &transactionRaw
	transaction := types.BaseTransaction{}
	transaction.CoinTransaction = &tx
	return &transaction, nil
}

func (coin Trx) CreateContract(txParams TrxTxParams) (proto.Message, error) {
	var transferContract proto.Message
	var err error
	switch txParams.ContractType {
	case core.Transaction_Contract_TransferContract:
		transferContract, err = coin.CreateTransferContract(txParams.FromAddress, txParams.ToAddress, txParams.Amount)
	case core.Transaction_Contract_FreezeBalanceV2Contract:
		transferContract, err = coin.CreateFreezeBalanceV2Contract(txParams.FromAddress, txParams.Amount,
			txParams.TronFreezeResourceCode)
	case core.Transaction_Contract_DelegateResourceContract:
		transferContract, err = coin.CreateDelegateResourceContract(txParams.FromAddress, txParams.ToAddress, txParams.Amount,
			txParams.TronFreezeResourceCode)
	case core.Transaction_Contract_UnfreezeBalanceV2Contract:
		transferContract, err = coin.CreateUnFreezeBalanceV2Contract(txParams.FromAddress,
			txParams.Amount, txParams.TronFreezeResourceCode)
	case core.Transaction_Contract_UnDelegateResourceContract:
		transferContract, err = coin.CreateUndelegateResourceContract(txParams.FromAddress, txParams.ToAddress, txParams.Amount,
			txParams.TronFreezeResourceCode)
	case core.Transaction_Contract_FreezeBalanceContract:
		var frozenDuration = txParams.TronFrozenDuration
		transferContract, err = coin.CreateFreezeBalanceContract(txParams.FromAddress, txParams.ToAddress,
			txParams.Amount, txParams.TronFreezeResourceCode, frozenDuration)
	case core.Transaction_Contract_UnfreezeBalanceContract:
		transferContract, err = coin.CreateUnFreezeBalanceContract(txParams.FromAddress, txParams.ToAddress,
			txParams.TronFreezeResourceCode)
	case core.Transaction_Contract_WithdrawExpireUnfreezeContract:
		transferContract, err = coin.CreateWithdrawExpireUnfreezeContract(txParams.FromAddress)
	case core.Transaction_Contract_AccountCreateContract:
		transferContract, err = coin.CreateAccountContract(txParams.FromAddress, txParams.ToAddress)

	}

	return transferContract, err
}

func amountToMinUnit(value decimal.Decimal, tokenDecimal int32) *big.Int {
	return value.Mul(decimal.NewFromInt32(10).Pow(decimal.NewFromInt32(tokenDecimal))).BigInt()
}

func (coin Trx) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	transaction := core.Transaction{}
	decodeString, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(decodeString, &transaction)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (coin Trx) CreateTransferContract(fromAddress string, toAddress string, amount decimal.Decimal) (proto.Message, error) {
	transferContract := core.TransferContract{}
	transferContract.Amount = amountToMinUnit(amount, 6).Int64()
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}

	transferContract.OwnerAddress = fromAddressBytes
	toAddressbytes, err := common.DecodeCheck(toAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	transferContract.ToAddress = toAddressbytes
	return &transferContract, nil
}

func (coin Trx) CreateFreezeBalanceContract(fromAddress string, toAddress string, amount decimal.Decimal, sourceCode core.ResourceCode, frozenDuration int64) (proto.Message, error) {
	freezeBalanceContract := core.FreezeBalanceContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	freezeBalanceContract.OwnerAddress = fromAddressBytes
	freezeBalanceContract.FrozenBalance = amountToMinUnit(amount, 6).Int64()
	freezeBalanceContract.Resource = sourceCode
	if toAddress != "" {
		fromAddressBytes, err := common.DecodeCheck(toAddress)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		freezeBalanceContract.ReceiverAddress = fromAddressBytes
	}
	freezeBalanceContract.FrozenDuration = frozenDuration
	return &freezeBalanceContract, nil
}

func (coin Trx) CreateUnFreezeBalanceContract(fromAddress string, toAddress string, sourceCode core.ResourceCode) (proto.Message, error) {
	unfreezeBalanceContract := core.UnfreezeBalanceContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	unfreezeBalanceContract.OwnerAddress = fromAddressBytes
	if toAddress != "" {
		fromAddressBytes, err := common.DecodeCheck(toAddress)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		unfreezeBalanceContract.ReceiverAddress = fromAddressBytes
	}

	unfreezeBalanceContract.Resource = sourceCode
	return &unfreezeBalanceContract, nil
}

func (coin Trx) CreateDelegateResourceContract(fromAddress string, toAddress string, amount decimal.Decimal, sourceCode core.ResourceCode) (proto.Message, error) {
	delegateResourceContract := core.DelegateResourceContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	delegateResourceContract.OwnerAddress = fromAddressBytes
	toAddressBytes, err := common.DecodeCheck(toAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	delegateResourceContract.ReceiverAddress = toAddressBytes
	delegateResourceContract.Balance = amountToMinUnit(amount, 6).Int64()
	delegateResourceContract.Resource = sourceCode

	return &delegateResourceContract, nil
}

func (coin Trx) CreateUndelegateResourceContract(fromAddress string, toAddress string, amount decimal.Decimal, sourceCode core.ResourceCode) (proto.Message, error) {
	unDelegateResourceContract := core.UnDelegateResourceContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	unDelegateResourceContract.OwnerAddress = fromAddressBytes
	toAddressBytes, err := common.DecodeCheck(toAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	unDelegateResourceContract.ReceiverAddress = toAddressBytes
	unDelegateResourceContract.Balance = amountToMinUnit(amount, 6).Int64()
	unDelegateResourceContract.Resource = sourceCode
	return &unDelegateResourceContract, nil
}

func (coin Trx) CreateFreezeBalanceV2Contract(fromAddress string, amount decimal.Decimal, sourceCode core.ResourceCode) (proto.Message, error) {
	freezeBalanceV2Contract := core.FreezeBalanceV2Contract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	freezeBalanceV2Contract.OwnerAddress = fromAddressBytes
	freezeBalanceV2Contract.FrozenBalance = amountToMinUnit(amount, 6).Int64()
	freezeBalanceV2Contract.Resource = sourceCode
	return &freezeBalanceV2Contract, nil
}

func (coin Trx) CreateUnFreezeBalanceV2Contract(fromAddress string, amount decimal.Decimal, sourceCode core.ResourceCode) (proto.Message, error) {
	unfreezeBalanceV2Contract := core.UnfreezeBalanceV2Contract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	unfreezeBalanceV2Contract.OwnerAddress = fromAddressBytes
	unfreezeBalanceV2Contract.Resource = sourceCode
	unfreezeBalanceV2Contract.UnfreezeBalance = amountToMinUnit(amount, 6).Int64()
	return &unfreezeBalanceV2Contract, nil
}

func (coin Trx) CreateWithdrawExpireUnfreezeContract(fromAddress string) (proto.Message, error) {
	unfreezeBalanceV2Contract := core.WithdrawExpireUnfreezeContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	unfreezeBalanceV2Contract.OwnerAddress = fromAddressBytes
	return &unfreezeBalanceV2Contract, nil
}
func (coin Trx) CreateWithdrawContract(fromAddress string) (proto.Message, error) {
	unfreezeBalanceV2Contract := core.WithdrawBalanceContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	unfreezeBalanceV2Contract.OwnerAddress = fromAddressBytes
	return &unfreezeBalanceV2Contract, nil
}

func (coin Trx) CreateAccountContract(fromAddress string, toAddress string) (proto.Message, error) {

	accountCreateContract := core.AccountCreateContract{}
	fromAddressBytes, err := common.DecodeCheck(fromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	accountCreateContract.OwnerAddress = fromAddressBytes
	createAddressBytes, err := common.DecodeCheck(toAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	accountCreateContract.AccountAddress = createAddressBytes
	accountCreateContract.Type = core.AccountType_Normal
	return &accountCreateContract, nil

}

func (coin Trx) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := TrxTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Trx) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	tx := baseTransaction.CoinTransaction.(*core.Transaction)
	rawData, err := proto.Marshal(tx.GetRawData())
	if err != nil {
		return nil, err
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	signature, err := crypto.Sign(hash[:], privateKeyECDSA)
	if err != nil {
		return nil, err
	}
	tx.Signature = append(tx.Signature, signature)
	txSigBytes, err := proto.Marshal(tx)
	var t = hex.EncodeToString(txSigBytes)
	return &t, nil
}

func (coin Trx) CalculateTxLength(txParams TrxTxParams, privateKey types.PrivateKey) (int64, error) {
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
