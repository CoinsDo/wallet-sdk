package coins

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/types"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/near/borsh-go"
	"github.com/shopspring/decimal"
	"github.com/textileio/near-api-go/keys"
	"github.com/textileio/near-api-go/transaction"
)

const CurrencyNear = "NEAR"

type NearTxParams struct {
	types.BaseTxParams
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	Memo        string          `json:"memo"`
	FromAddress string          `json:"fromAddress"`
	ToAddress   string          `json:"toAddress"`
	Nonce       uint64          `json:"nearNonce"`
	BlockHash   string          `json:"nearBlockHash"`
	TxType      int32           `json:"txType"`
}

var coinNear Near

func init() {
	coinNear = Near{}
	RegisterCoin(coinNear)
}

type Near struct {
}

func (coin Near) GetCurrency() string {
	return CurrencyNear
}

func (coin Near) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), index)
}

func (coin Near) GetBasePath(testNet bool) string {
	return "m/44'/397'/%d'"
}

func (coin Near) ChainName() string {
	return CurrencyNear
}

func (coin Near) GetDecimal() int {
	return 24
}

func (coin Near) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	return base58.Decode(key), nil
}

func (coin Near) PrivateKeyToString(key types.PrivateKey) (string, error) {

	return base58.Encode(key), nil
}

func (coin Near) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, privateKey[32:])
	var adddress = types.CoinAddress{}
	adddress.AddressStr = strings.TrimPrefix(hexutil.Encode(publicKey), "0x")
	return &adddress, nil
}

func (coin Near) GetDeriver() deriver.Deriver {
	return &deriver.Ed25519Deriver{}
}

func (coin Near) GetEmptyTransactionParams() types.TxParams {
	return NearTxParams{}
}

func (coin Near) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := txParams.(NearTxParams)
	var fromAddress = extraParams.FromAddress
	if !strings.HasPrefix(extraParams.FromAddress, "0x") {
		fromAddress = "0x" + fromAddress
	}

	decode, err := hexutil.Decode(fromAddress)
	if err != nil {
		return nil, err
	}

	mul := extraParams.Amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(24)))
	hashByte := base58.Decode(extraParams.BlockHash)

	if err != nil {
		return nil, err
	}

	action := transaction.TransferAction(*mul.BigInt())
	var actions []transaction.Action
	keyData := [32]byte{}
	copy(keyData[:], decode)
	pubKey := transaction.PublicKey{
		KeyType: 0,
		Data:    keyData,
	}
	newTransaction := transaction.NewTransaction(strings.TrimPrefix(fromAddress, "0x"), pubKey, extraParams.Nonce, extraParams.ToAddress, hashByte, append(actions, action))
	if extraParams.TxType == 1 {
		action = transaction.DeleteAccountAction(extraParams.ToAddress)
		newTransaction = transaction.NewTransaction(strings.TrimPrefix(fromAddress, "0x"), pubKey, extraParams.Nonce, extraParams.FromAddress, hashByte, append(actions, action))
	}

	tx := types.BaseTransaction{}
	tx.CoinTransaction = newTransaction

	return &tx, nil
}

func (coin Near) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	txBytes, err := base64.StdEncoding.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}
	signedTransaction := transaction.SignedTransaction{}
	err = borsh.Deserialize(&signedTransaction, txBytes)
	if err != nil {
		return nil, err
	}
	return signedTransaction, nil
}

func (coin Near) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := NearTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Near) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	tx := baseTransaction.CoinTransaction.(*transaction.Transaction)
	fromString, err := keys.NewKeyPairFromString(base58.Encode(privateKey))
	if err != nil {
		return nil, err
	}
	_, signedTx, err := transaction.SignTransaction(*tx, fromString, "", "")
	if err != nil {
		return nil, err
	}
	bytes, err := borsh.Serialize(*signedTx)
	if err != nil {
		return nil, err
	}
	toString := base64.StdEncoding.EncodeToString(bytes)
	return &toString, nil
}
