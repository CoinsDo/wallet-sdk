package coins

import (
	"encoding/json"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/shopspring/decimal"
	"strconv"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencyOmni = "OMNI"

type OmniTxParams struct {
	BtcTxParams
	Amount          decimal.Decimal `json:"amount"`
	ToAddress       string          `json:"toAddress"`
	ContractAddress string          `json:"commonContractAddress"`
}

var coinOmni Omni

func init() {
	coinOmni = Omni{}
	RegisterUtxoCoin(coinOmni)

}

type Omni struct {
	Btc
}

func (coin Omni) GetCurrency() string {
	return CurrencyOmni
}

func (coin Omni) GetEmptyTransactionParams() types.TxParams {
	return OmniTxParams{}
}

func (coin Omni) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := OmniTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Omni) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := txParams.(OmniTxParams)
	var unspends = extraParams.Unspends
	contractAddress := extraParams.ContractAddress
	var propertyID, _ = strconv.ParseUint(contractAddress, 10, 32)
	changeAddress := extraParams.ChangeAddress
	var totalAmount = btcutil.Amount(0)
	params := coin.GetNetParams(testNet)
	var currentInputs []*wire.TxIn
	var currentInputValues []btcutil.Amount
	var inputScripts [][]byte
	for _, unspend := range unspends {

		amount, err := btcutil.NewAmount(unspend.TxValue.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
		}
		totalAmount += amount
		address, err := btcutil.DecodeAddress(unspend.Address, &params)
		if err != nil {
			return nil, errors.ErrorInvalidSendAddress
		}
		script, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}
		inputScripts = append(inputScripts, script)
		hash, err := chainhash.NewHashFromStr(unspend.TxHash)
		if err != nil {
			return nil, errors.ErrorInvalidInput
		}
		nextInput := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: unspend.TxOutputN,
		}, nil, nil)
		currentInputs = append(currentInputs, nextInput)
		currentInputValues = append(currentInputValues, amount)
	}

	feeAmount, err := btcutil.NewAmount(extraParams.Fee.InexactFloat64())
	if err != nil {
		return nil, err
	}
	var txOut []*wire.TxOut

	var need = decimal.New(0, 0)

	decAddr, err := btcutil.DecodeAddress(extraParams.ToAddress, &params)

	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	script, err := txscript.PayToAddrScript(decAddr)
	if err != nil {
		return nil, err
	}
	txOut = append(txOut, &wire.TxOut{
		Value:    int64(MinNondustOutput),
		PkScript: script,
	})

	opreturnScript, err := GetClassCOpreturnDataScript(uint(propertyID), extraParams.Amount.InexactFloat64(), true)
	if err != nil {
		return nil, err
	}
	opreturnTxOut := wire.NewTxOut(0, opreturnScript)
	txOut = append(txOut, opreturnTxOut)

	need.Add(decimal.NewFromFloat(extraParams.Fee.InexactFloat64()))
	if err != nil {
		return nil, err
	}
	inputSource := func(target btcutil.Amount) (total btcutil.Amount, inputs []*wire.TxIn, inputValues []btcutil.Amount, scripts [][]byte, err error) {
		return totalAmount, currentInputs, currentInputValues, inputScripts, nil
	}

	changeSource := txauthor.ChangeSource{}

	changeSource.NewScript = func() ([]byte, error) {
		if changeAddress == "" {
			return nil, nil
		} else {
			changeAddr, err := btcutil.DecodeAddress(changeAddress, &params)
			if err != nil {
				return nil, errors.ErrorInvalidAddress
			}
			script, err := txscript.PayToAddrScript(changeAddr)
			if err != nil {
				return nil, err
			}
			return script, nil
		}
	}

	unsignedTransaction, err := coin.NewUnsignedTransaction(currentInputs, txOut, feeAmount, inputSource, &changeSource, changeAddress != "")
	if err != nil {
		return nil, err
	}

	return &types.BaseTransaction{CoinTransaction: unsignedTransaction}, nil
}
