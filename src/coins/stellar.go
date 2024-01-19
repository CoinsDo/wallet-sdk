package coins

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/txnbuild"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencyStellar = "XLM"

type StellarTxParams struct {
	types.BaseTxParams

	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	Memo        string          `json:"memo"`
	FromAddress string          `json:"fromAddress"`
	ToAddress   string          `json:"toAddress"`

	NeedCreateAccount bool  `json:"stellarNeedCreateAccount"`
	Sequence          int64 `json:"stellarSequence"`
	TimeOut           int64 `json:"stellarTimeOut"`
}

var coinStellar Stellar

func init() {
	coinStellar = Stellar{}
	RegisterCoin(coinStellar)
}

type Stellar struct {
}

func (coin Stellar) GetCurrency() string {
	return CurrencyStellar
}

func (coin Stellar) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), index)
}

func (coin Stellar) GetBasePath(testNet bool) string {
	return "m/44'/148'/%d'"
}

func (coin Stellar) ChainName() string {
	return CurrencyStellar
}

func (coin Stellar) GetDecimal() int {
	return 7
}

func (coin Stellar) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	return strkey.Decode(strkey.VersionByteSeed, key)
}

func (coin Stellar) PrivateKeyToString(key types.PrivateKey) (string, error) {
	seed, err := strkey.Encode(strkey.VersionByteSeed, key[:])
	if err != nil {
		return "", err
	}
	return seed, nil
}

func (coin Stellar) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	var rawSeed [32]byte
	copy(rawSeed[:], privateKey[0:32])
	seed, err := keypair.FromRawSeed(rawSeed)
	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = seed.Address()
	return &adddress, nil
}

func (coin Stellar) GetDeriver() deriver.Deriver {
	return &deriver.StellarDeriver{}
}

func (coin Stellar) GetEmptyTransactionParams() types.TxParams {
	return StellarTxParams{}
}

func (coin Stellar) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {

	extraParams := txParams.(StellarTxParams)
	fromAccount, err := keypair.Parse(extraParams.FromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	_, err = keypair.Parse(extraParams.ToAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	var timeOut = int64(60 * 60)
	if extraParams.TimeOut != 0 {
		timeOut = extraParams.TimeOut
	}
	transactionParams := txnbuild.TransactionParams{
		SourceAccount: &txnbuild.SimpleAccount{
			AccountID: fromAccount.Address(),
			Sequence:  extraParams.Sequence,
		},
		IncrementSequenceNum: true,
		BaseFee:              extraParams.Fee.IntPart(),
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewTimeout(timeOut),
		},
	}
	if extraParams.NeedCreateAccount {
		transactionParams.Operations =
			[]txnbuild.Operation{&txnbuild.CreateAccount{
				Destination: extraParams.ToAddress,
				Amount:      extraParams.Amount.String(),
			}}
	} else {
		transactionParams.Operations =
			[]txnbuild.Operation{&txnbuild.Payment{
				Destination:   extraParams.ToAddress,
				Amount:        extraParams.Amount.String(),
				SourceAccount: extraParams.FromAddress,
				Asset:         txnbuild.NativeAsset{},
			}}
	}

	if extraParams.Memo != "" {
		textMemo := txnbuild.MemoText(extraParams.Memo)
		transactionParams.Memo = textMemo
	}

	tx, err := txnbuild.NewTransaction(transactionParams)
	if err != nil {
		return nil, err
	}
	return &types.BaseTransaction{
		CoinTransaction: tx,
	}, nil
}

func (coin Stellar) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	return txnbuild.TransactionFromXDR(rawTx)
}

func (coin Stellar) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := StellarTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Stellar) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	transaction := baseTransaction.CoinTransaction.(*txnbuild.Transaction)
	var rawSeed [32]byte
	copy(rawSeed[:], privateKey[0:32])
	stellarKeyPair, err := keypair.FromRawSeed(rawSeed)
	var signText = network.TestNetworkPassphrase
	if !testNet {
		signText = network.PublicNetworkPassphrase
	}
	if err != nil {
		return nil, err
	}
	sign, err := transaction.Sign(signText, stellarKeyPair)
	if err != nil {
		return nil, err
	}
	base64, err := sign.Base64()
	if err != nil {
		return nil, err
	}
	return &base64, nil
}
