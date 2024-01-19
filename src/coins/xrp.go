package coins

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rubblelabs/ripple/crypto"
	"github.com/rubblelabs/ripple/data"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	xrpCrypto "wallet-sdk/src/coins/xrp"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencyXrp = "XRP"

type XrpTxParams struct {
	types.BaseTxParams
	ToAddress          string          `json:"toAddress"`
	Amount             decimal.Decimal `json:"amount"`
	Fee                decimal.Decimal `json:"fee"`
	Memo               string          `json:"memo"`
	FromAddress        string          `json:"fromAddress"`
	Flags              uint32          `json:"flags"`
	Sequence           uint32          `json:"sequence"`
	LastLedgerSequence uint32          `json:"lastLedgerSequence"`
	TxType             XrpTxType       `json:"txType"`
}

type XrpTxType int

const (
	TxXrpTransaction   XrpTxType = 0
	TxXrpDeleteAccount XrpTxType = 1
	TxXrpSetAccountTag XrpTxType = 2
)

var coinXrp Xrp

func init() {
	coinXrp = Xrp{}
	RegisterCoin(coinXrp)
}

type Xrp struct {
}

func (coin Xrp) GetCurrency() string {
	return CurrencyXrp
}

func (coin Xrp) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Xrp) GetBasePath(testNet bool) string {
	return "m/44'/144'/%d'/%d/%d"
}

func (coin Xrp) ChainName() string {
	return CurrencyXrp
}

func (coin Xrp) GetDecimal() int {
	return 6
}

func (coin Xrp) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	privateKey, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (coin Xrp) PrivateKeyToString(key types.PrivateKey) (string, error) {
	return hexutil.Encode(key), nil
}

func (coin Xrp) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	ecdsaKey, err := crypto.NewECDSAKey(privateKey)
	if err != nil {
		return nil, err
	}
	sequence := uint32(0)
	address, err := crypto.AccountId(ecdsaKey, &sequence)
	if err != nil {
		return nil, err
	}

	var adddress = types.CoinAddress{}
	adddress.AddressStr = address.String()
	return &adddress, nil
}

func (coin Xrp) GenerateAddressFromSecret(secret string) (*types.CoinAddress, error) {
	decodeSeed, err := xrpCrypto.DecodeSeed(secret)
	if err != nil {
		return nil, err
	}
	deriveKey, err := decodeSeed.DeriveKey()
	if err != nil {
		return nil, err
	}

	md160 := crypto.Sha256RipeMD160(deriveKey.Public(nil))
	accountId, err := crypto.NewAccountId(md160)
	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = accountId.String()
	adddress.PrivateKey = hexutil.Encode(deriveKey.Private(nil))

	return &adddress, nil
}

func (coin Xrp) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Xrp) GetEmptyTransactionParams() types.TxParams {
	return XrpTxParams{}
}

func (coin Xrp) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {

	extraParams := txParams.(XrpTxParams)
	if extraParams.TxType == TxXrpTransaction {
		if extraParams.Amount.LessThan(decimal.NewFromFloat(0.000001)) {
			return nil, errors.ErrorLessThanMinimum
		}
	}
	num := strconv.FormatFloat(extraParams.Amount.Mul(decimal.NewFromInt(1000000)).InexactFloat64(), 'f', 0, 64)
	fees := strconv.FormatFloat(extraParams.Fee.Mul(decimal.NewFromInt(1000000)).InexactFloat64(), 'f', 0, 64)
	var txType = "Payment"
	if extraParams.TxType == TxXrpDeleteAccount {
		txType = "AccountDelete"
	} else if extraParams.TxType == TxXrpSetAccountTag {
		txType = "AccountSet"
	}

	_, err := crypto.NewRippleHashCheck(extraParams.FromAddress, crypto.RIPPLE_ACCOUNT_ID)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	_, err = crypto.NewRippleHashCheck(extraParams.ToAddress, crypto.RIPPLE_ACCOUNT_ID)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	paramsInfo := TxExtraParams{
		TransactionType:    txType,
		Account:            extraParams.FromAddress,
		Amount:             num,
		Destination:        extraParams.ToAddress,
		Fee:                fees,
		Flags:              extraParams.Flags,
		Sequence:           extraParams.Sequence,
		LastLedgerSequence: extraParams.LastLedgerSequence,
	}
	if extraParams.Memo != "" {
		memoInt, err := strconv.ParseUint(extraParams.Memo, 10, 32)
		if err != nil {
			return nil, err
		}
		u := uint32(memoInt)
		paramsInfo.DestinationTag = &u
	}
	tx := *data.NewTransactionWithMetadata(data.TransactionType(data.ECDSA))
	b, _ := json.Marshal(paramsInfo)
	txErr := tx.UnmarshalJSON(b)
	if txErr != nil {
		fmt.Println("Invalid JSON transaction: ", txErr)
		return nil, txErr
	}

	hashTx := types.BaseTransaction{CoinTransaction: tx}
	return &hashTx, nil
}

func (coin Xrp) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	return nil, errors.ErrorDecodeNotSupported
}

func (coin Xrp) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	key, err := crypto.NewECDSAKey(privateKey)
	if err != nil {
		return nil, err
	}
	defer zeroKey(key)
	transaction := baseTransaction.CoinTransaction.(data.TransactionWithMetaData).Transaction
	lastLedgerSequence := *transaction.GetBase().LastLedgerSequence + 4
	base := transaction.GetBase()
	var sequence uint32
	accountId, err := crypto.AccountId(key, &sequence)
	if err != nil {
		return nil, err
	}
	account, err := data.NewAccountFromAddress(accountId.String())
	if err != nil {
		return nil, err
	}
	base.Account = *account
	base.LastLedgerSequence = &lastLedgerSequence

	err = data.Sign(transaction, key, &sequence)
	if err != nil {
		return nil, err
	}

	_, txRaw, err := data.Raw(transaction)
	if err != nil {
		return nil, err
	}
	encodedTx := fmt.Sprintf("%X", txRaw)

	return &encodedTx, nil
}

func (coin Xrp) SignBySecret(baseTransaction *types.BaseTransaction, secret string) (*string, error) {

	decodeSeed, err := xrpCrypto.DecodeSeed(secret)
	if err != nil {
		return nil, err
	}
	deriveKey, err := decodeSeed.DeriveKey()
	if err != nil {
		return nil, err
	}

	md160 := crypto.Sha256RipeMD160(deriveKey.Public(nil))
	accountId, err := crypto.NewAccountId(md160)
	if err != nil {
		return nil, err
	}

	// Sign the transaction
	transaction := baseTransaction.CoinTransaction.(data.TransactionWithMetaData).Transaction
	lastLedgerSequence := *transaction.GetBase().LastLedgerSequence + 4
	base := transaction.GetBase()
	if err != nil {
		return nil, err
	}
	account, err := data.NewAccountFromAddress(accountId.String())
	if err != nil {
		return nil, err
	}
	base.Account = *account
	base.LastLedgerSequence = &lastLedgerSequence

	err = data.Sign(transaction, deriveKey, nil)
	if err != nil {
		return nil, err
	}

	// Convert to hex
	_, txRaw, err := data.Raw(transaction)
	if err != nil {
		return nil, err
	}
	encodedTx := fmt.Sprintf("%X", txRaw)

	return &encodedTx, nil
}

func (coin Xrp) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := XrpTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

type TxExtraParams struct {
	TransactionType    string  `json:"TransactionType"`
	Amount             string  `json:"Amount"`
	Destination        string  `json:"Destination"`
	Account            string  `json:"ccount"`
	Fee                string  `json:"fee"`
	Sequence           uint32  `json:"sequence"`
	Flags              uint32  `json:"flags"`
	LastLedgerSequence uint32  `json:"lastLedgerSequence"`
	DestinationTag     *uint32 `json:"DestinationTag"`
	//unmarshal
}

// ZeroKey Removes key data from memory
func zeroKey(k crypto.Key) {
	b := k.Private(nil)
	for i := range b {
		b[i] = 0
	}
}
