package coins

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/algorand/go-algorand-sdk/types"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/ed25519"
	"wallet-sdk/src/deriver"
	errors2 "wallet-sdk/src/errors"
	types2 "wallet-sdk/src/types"
)

const CurrencyAlgo = "ALGO"

type AlgoTxParams struct {
	types2.BaseTxParams
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	Memo        string          `json:"memo"`
	FromAddress string          `json:"fromAddress"`
	FirstRound  uint64          `json:"algoFirstRound"`
	GenesisID   string          `json:"algoGenesisID"`
	GenesisHash string          `json:"algoGenesisHash"`
	//IsFlatFee =true The fee is independent of the transaction byte size.
	IsFlatFee bool   `json:"algoIsFlatFee"`
	ToAddress string `json:"toAddress"`
}

var coinAlgo Algo

func init() {
	coinAlgo = Algo{}
	RegisterCoin(coinAlgo)
}

type Algo struct {
}

func (coin Algo) GetCurrency() string {
	return CurrencyAlgo
}

func (coin Algo) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Algo) GetBasePath(testNet bool) string {
	return "m/44'/283'/%d'/%d/%d"
}

func (coin Algo) ChainName() string {
	return CurrencyAlgo
}

func (coin Algo) GetDecimal() int {
	return 6
}

func (coin Algo) PrivateKeyToString(key types2.PrivateKey) (string, error) {
	return base58.Encode(key), nil
}
func (coin Algo) PrivateKeyFromString(key string) (types2.PrivateKey, error) {
	privateKeyByte := base58.Decode(key)
	return privateKeyByte, nil
}

func (coin Algo) GenerateAddress(privateKey types2.PrivateKey, testNet bool) (*types2.CoinAddress, error) {
	key := ed25519.PrivateKey{}
	key = []byte(privateKey)
	account, err := crypto.AccountFromPrivateKey(key)
	if err != nil {
		return nil, err
	}

	var adddress = types2.CoinAddress{}
	adddress.AddressStr = account.Address.String()
	return &adddress, nil
}

func (coin Algo) GetDeriver() deriver.Deriver {
	return &deriver.AlgoDeriver{}
}

func (coin Algo) GetEmptyTransactionParams() types2.TxParams {
	return AlgoTxParams{}
}

func (coin Algo) CreateTransaction(txParams types2.TxParams, testNet bool) (*types2.BaseTransaction, error) {

	extraParams := txParams.(AlgoTxParams)
	_, err := types.DecodeAddress(extraParams.FromAddress)
	if err != nil {
		return nil, errors2.ErrorInvalidSendAddress
	}
	_, err = types.DecodeAddress(extraParams.ToAddress)
	if err != nil {
		return nil, errors2.ErrorInvalidAddress
	}

	decodeString, err := base64.StdEncoding.DecodeString(extraParams.GenesisHash)
	if err != nil {
		return nil, err
	}
	mul := extraParams.Amount.Mul(decimal.NewFromInt(1000000))
	feeResult := extraParams.Fee.Mul(decimal.NewFromInt(1000000))
	var txn types.Transaction
	if extraParams.IsFlatFee {
		txn, err = transaction.MakePaymentTxnWithFlatFee(extraParams.FromAddress, extraParams.ToAddress,
			uint64(feeResult.IntPart()), uint64(mul.IntPart()),
			extraParams.FirstRound, extraParams.FirstRound+1000, []byte(extraParams.Memo),
			"", extraParams.GenesisID, decodeString)
	} else {
		txn, err = transaction.MakePaymentTxn(extraParams.FromAddress, extraParams.ToAddress,
			uint64(feeResult.IntPart()), uint64(mul.IntPart()),
			extraParams.FirstRound, extraParams.FirstRound+1000, []byte(extraParams.Memo),
			"", extraParams.GenesisID, decodeString)
	}

	if err != nil {
		return nil, err
	}
	baseTx := types2.BaseTransaction{}
	baseTx.CoinTransaction = txn
	return &baseTx, nil
}

func (coin Algo) EstimateSize(extraParams AlgoTxParams) (uint64, error) {
	decodeString, err := base64.StdEncoding.DecodeString(extraParams.GenesisHash)
	if err != nil {
		return 0, err
	}
	mul := extraParams.Amount.Mul(decimal.NewFromInt(1000000))
	feeResult := extraParams.Fee.Mul(decimal.NewFromInt(1000000))

	fromAddr, err := types.DecodeAddress(extraParams.FromAddress)
	if err != nil {
		return 0, err
	}

	// Decode to address
	toAddr, err := types.DecodeAddress(extraParams.ToAddress)
	if err != nil {
		return 0, err
	}

	// Decode the CloseRemainderTo address, if present
	var closeRemainderToAddr types.Address

	// Decode GenesisHash

	var gh types.Digest
	copy(gh[:], decodeString)

	// Build the transaction
	tx := types.Transaction{
		Type: types.PaymentTx,
		Header: types.Header{
			Sender:      fromAddr,
			Fee:         types.MicroAlgos(feeResult.IntPart()),
			FirstValid:  types.Round(extraParams.FirstRound),
			LastValid:   types.Round(extraParams.FirstRound + 1000),
			Note:        []byte(extraParams.Memo),
			GenesisID:   extraParams.GenesisID,
			GenesisHash: gh,
		},
		PaymentTxnFields: types.PaymentTxnFields{
			Receiver:         toAddr,
			Amount:           types.MicroAlgos(mul.IntPart()),
			CloseRemainderTo: closeRemainderToAddr,
		},
	}
	if err != nil {
		return 0, err
	}
	size, err := transaction.EstimateSize(tx)
	if err != nil {
		return 0, err
	}
	return size, nil

}

func (coin Algo) GetTransactionParamsFromJson(paramsJson string) types2.TxParams {
	params := AlgoTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Algo) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	txBytes, err := base64.StdEncoding.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}
	txn := types.SignedTxn{}
	err = msgpack.Decode(txBytes, &txn)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (coin Algo) SignTx(baseTransaction *types2.BaseTransaction, testNet bool, privateKey types2.PrivateKey) (*string, error) {
	tx := baseTransaction.CoinTransaction.(types.Transaction)
	key := ed25519.PrivateKey{}
	key = []byte(privateKey)
	_, txBytes, err := crypto.SignTransaction(key, tx)
	if err != nil {
		return nil, err
	}

	toString := base64.StdEncoding.EncodeToString(txBytes)

	return &toString, err
}
