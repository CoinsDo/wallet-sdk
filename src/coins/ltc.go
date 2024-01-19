package coins

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ltcsuite/ltcd/btcec/v2"
	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/chaincfg/chainhash"
	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/ltcsuite/ltcd/txscript"
	"github.com/ltcsuite/ltcd/wire"
	"github.com/ltcsuite/ltcwallet/wallet/txauthor"
	"github.com/ltcsuite/ltcwallet/wallet/txrules"
	"github.com/ltcsuite/ltcwallet/wallet/txsizes"
	"golang.org/x/crypto/ripemd160"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencyLtc = "LTC"

var coinLtc Ltc

func init() {
	coinLtc = Ltc{}
	RegisterUtxoCoin(coinLtc)

}

type Ltc struct {
	Btc
}

func (coin Ltc) GetCurrency() string {
	return CurrencyLtc
}

func (coin Ltc) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Ltc) GetBasePath(testNet bool) string {
	return "m/44'/2'/%d'/%d/%d"
}

func (coin Ltc) GetSegwitBasePath(testNet bool) string {
	return "m/84'/2'/%d'/%d/%d"
}

func (coin Ltc) GetSegwitPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetSegwitBasePath(testNet), 0, 0, index)
}

func (coin Ltc) ChainName() string {
	return CurrencyLtc
}

func (coin Ltc) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := ltcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}

func (coin Ltc) PrivateKeyToString(key types.PrivateKey) (string, error) {
	btcPrivKey, _ := btcec.PrivKeyFromBytes(key)
	wif, err := ltcutil.NewWIF(btcPrivKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}

func (coin Ltc) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	_, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	params := getLtcNetParams(testNet)

	newAddressPubKey, err := ltcutil.NewAddressPubKey(pubKey.SerializeCompressed(), &params)
	var adddress = types.CoinAddress{}
	adddress.AddressStr = newAddressPubKey.EncodeAddress()
	return &adddress, nil
}

func (coin Ltc) GenerateSegwitAddressByPrivateKey(key string, testnet bool) (*types.CoinAddress, error) {
	wif, err := ltcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}

	privKey := wif.PrivKey

	toECDSA, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Serialize()))
	if err != nil {
		return nil, err
	}
	return coin.GenerateSegwitAddress(toECDSA.D.Bytes(), testnet)
}

func (coin Ltc) GenerateSegwitAddress(keyByte []byte, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	_, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	params := getLtcNetParams(testNet)

	shaHash := sha256.Sum256(pubKey.SerializeCompressed())
	ripeHasher := ripemd160.New()
	_, _ = ripeHasher.Write(shaHash[:])
	publicKeyHash := ripeHasher.Sum(nil)

	hash, err := ltcutil.NewAddressWitnessPubKeyHash(publicKeyHash, &params)
	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = hash.EncodeAddress()
	return &adddress, nil
}

func (coin Ltc) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Ltc) GetEmptyTransactionParams() types.TxParams {
	return BtcTxParams{}
}

func (coin Ltc) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := txParams.(BtcTxParams)
	var unspends = extraParams.Unspends
	var receivers = extraParams.Receivers
	changeAddress := extraParams.ChangeAddress
	var totalAmount = ltcutil.Amount(0)
	params := getLtcNetParams(testNet)
	var currentInputs []*wire.TxIn
	var currentInputValues []ltcutil.Amount
	var inputScripts [][]byte
	for _, unspend := range unspends {

		amount, err := ltcutil.NewAmount(unspend.TxValue.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
		}
		totalAmount += amount
		address, err := ltcutil.DecodeAddress(unspend.Address, &params)
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
	feeAmount, err := ltcutil.NewAmount(extraParams.Fee.InexactFloat64())
	if err != nil {
		return nil, err
	}
	var txOut []*wire.TxOut
	for _, receiver := range receivers {
		decAddr, err := ltcutil.DecodeAddress(receiver.Address, &params)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		a, err := ltcutil.NewAmount(receiver.Value.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}

		script, err := txscript.PayToAddrScript(decAddr)
		if err != nil {
			return nil, err
		}
		txOut = append(txOut, &wire.TxOut{
			Value:    int64(a),
			PkScript: script,
		})
	}
	inputSource := func(target ltcutil.Amount) (total ltcutil.Amount, inputs []*wire.TxIn, inputValues []ltcutil.Amount, scripts [][]byte, err error) {
		return totalAmount, currentInputs, currentInputValues, inputScripts, nil
	}

	changeSource := txauthor.ChangeSource{}

	changeSource.NewScript = func() ([]byte, error) {
		if changeAddress == "" {
			return nil, nil
		} else {
			changeAddr, err := ltcutil.DecodeAddress(changeAddress, &params)
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

	unsignedTransaction, err := coin.NewUnsignedTransaction(txOut, feeAmount, inputSource, &changeSource, changeAddress != "")
	if err != nil {
		return nil, err
	}

	return &types.BaseTransaction{CoinTransaction: unsignedTransaction}, nil
}

func (coin Ltc) NewUnsignedTransaction(outputs []*wire.TxOut, relayFeePerKb ltcutil.Amount, fetchInputs txauthor.InputSource, fetchChange *txauthor.ChangeSource, hasChange bool) (*txauthor.AuthoredTx, error) {

	targetAmount := coin.SumOutputValues(outputs)
	estimatedSize := txsizes.EstimateSerializeSize(0, outputs, hasChange)
	targetFee := txrules.FeeForSerializeSize(relayFeePerKb, estimatedSize)

	for {
		inputAmount, inputs, inputValues, scripts, err := fetchInputs(targetAmount + targetFee)
		if err != nil {
			return nil, err
		}
		if inputAmount < targetAmount+targetFee {
			return nil, errors.ErrorInsufficientFunds
		}

		maxSignedSize := txsizes.EstimateSerializeSize(len(inputs), outputs, true)
		maxRequiredFee := txrules.FeeForSerializeSize(relayFeePerKb, maxSignedSize)
		remainingAmount := inputAmount - targetAmount
		if remainingAmount < maxRequiredFee {
			targetFee = maxRequiredFee
			continue
		}

		unsignedTransaction := &wire.MsgTx{
			Version:  wire.TxVersion,
			TxIn:     inputs,
			TxOut:    outputs,
			LockTime: 0,
		}
		changeIndex := -1
		if fetchChange != nil {
			changeAmount := inputAmount - targetAmount - maxRequiredFee
			if changeAmount != 0 && !isDustAmount(changeAmount,
				txsizes.P2PKHPkScriptSize, relayFeePerKb) {
				changeScript, err := fetchChange.NewScript()
				if changeScript != nil {
					if err != nil {
						return nil, err
					}
					if len(changeScript) > txsizes.P2PKHPkScriptSize {
						return nil, errors.ErrorFeeAddressError
					}
					change := wire.NewTxOut(int64(changeAmount), changeScript)
					l := len(outputs)
					unsignedTransaction.TxOut = append(outputs[:l:l], change)
					changeIndex = l
				}
			}
		}

		return &txauthor.AuthoredTx{
			Tx:              unsignedTransaction,
			PrevScripts:     scripts,
			PrevInputValues: inputValues,
			TotalInput:      inputAmount,
			ChangeIndex:     changeIndex,
		}, nil
	}
}

func (coin Ltc) SignTx(baseTransaction *types.BaseTransaction, testNet bool, keyByte types.PrivateKey) (*string, error) {
	authoredTx := baseTransaction.CoinTransaction.(*txauthor.AuthoredTx)
	params := getLtcNetParams(testNet)
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	wif, err := ltcutil.NewWIF(btcPrivKey, &params, true)
	if err != nil {
		return nil, err
	}
	key, err := ltcutil.NewAddressPubKeyHash(ltcutil.Hash160(wif.SerializePubKey()), &params)
	if err != nil {
		return nil, err
	}
	mkGetKey := func() txscript.KeyDB {

		return txscript.KeyClosure(func(addr ltcutil.Address) (*btcec.PrivateKey,
			bool, error) {
			return wif.PrivKey, wif.CompressPubKey, nil
		})
	}

	getScript := txscript.ScriptClosure(func(
		addr ltcutil.Address) ([]byte, error) {

		return txscript.PayToAddrScript(key)
	})

	source := LtcSecProvider{testNet,
		mkGetKey(), getScript,
	}
	err = authoredTx.AddAllInputScripts(source)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Grow(authoredTx.Tx.SerializeSize())
	err = authoredTx.Tx.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	toString := hex.EncodeToString(buf.Bytes())
	return &toString, nil
}

func (coin Ltc) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error) {

	return coin.SignMultipleSendAddress(baseTransaction, keys, testNet)
}

func (coin Ltc) SignMultipleSendAddress(tx *types.BaseTransaction, keys map[string]types.PrivateKey, testNet bool) (*string, error) {
	netParams := getLtcNetParams(testNet)
	authoredTx := tx.CoinTransaction.(*txauthor.AuthoredTx)

	mkGetKey := func() txscript.KeyDB {

		return txscript.KeyClosure(func(addr ltcutil.Address) (*btcec.PrivateKey,
			bool, error) {

			derivedKey := keys[addr.EncodeAddress()]

			privateKey, err := crypto.ToECDSA(derivedKey)
			if err != nil {
				return nil, false, err
			}
			privateKeyBytes := crypto.FromECDSA(privateKey)

			btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

			wif, err := ltcutil.NewWIF(btcPrivKey, &netParams, true)
			if err != nil {
				return nil, false, err
			}

			return wif.PrivKey, wif.CompressPubKey, nil
		})
	}

	getScript := txscript.ScriptClosure(func(
		addr ltcutil.Address) ([]byte, error) {
		derivedKey := keys[addr.EncodeAddress()]

		privateKey, err := crypto.ToECDSA(derivedKey)
		if err != nil {
			return nil, err
		}
		privateKeyBytes := crypto.FromECDSA(privateKey)

		btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		wif, err := ltcutil.NewWIF(btcPrivKey, &netParams, true)
		if err != nil {
			return nil, err
		}
		key, err := ltcutil.NewAddressPubKeyHash(ltcutil.Hash160(wif.SerializePubKey()), &netParams)
		if err != nil {
			return nil, err
		}
		return txscript.PayToAddrScript(key)
	})

	source := LtcSecProvider{testNet,
		mkGetKey(), getScript,
	}
	err := authoredTx.AddAllInputScripts(source)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Grow(authoredTx.Tx.SerializeSize())
	err = authoredTx.Tx.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	toString := hex.EncodeToString(buf.Bytes())
	return &toString, nil
}

func (coin Ltc) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := getLtcNetParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, params)

}

func (coin Ltc) EstimateTxSizes(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet chaincfg.Params) int {
	changeSize := 0
	outputCount := len(outputAddrs)
	if hasExtraChangeAddr {
		changeSize = txsizes.P2PKHOutputSize
		outputCount++
	}

	return 8 + wire.VarIntSerializeSize(uint64(inputCount)) +
		wire.VarIntSerializeSize(uint64(outputCount)) +
		inputCount*txsizes.RedeemP2PKHInputSize +
		coin.SumOutputSerializeSizesOfChainParams(outputAddrs, testNet) +
		changeSize
}

func (coin Ltc) SumOutputSerializeSizesOfChainParams(outputAddrs []string, params chaincfg.Params) int {
	var sizeSum = 0
	for _, addr := range outputAddrs {
		decodeAddress, err := ltcutil.DecodeAddress(addr, &params)
		if err != nil {
			return 0
		}
		script, err := txscript.PayToAddrScript(decodeAddress)
		if err != nil {
			return 0
		}
		sizeSum += 8 + wire.VarIntSerializeSize(uint64(len(script))) + len(script)

	}
	return sizeSum
}

func getLtcNetParams(testNet bool) chaincfg.Params {
	var params chaincfg.Params

	if testNet {
		params = chaincfg.TestNet4Params
	} else {
		params = chaincfg.MainNetParams
	}
	return params
}

func isDustAmount(amount ltcutil.Amount, scriptSize int, relayFeePerKb ltcutil.Amount) bool {
	return amount < MinNondustOutput
}

func (coin Ltc) SumOutputValues(outputs []*wire.TxOut) (totalOutput ltcutil.Amount) {
	for _, txOut := range outputs {
		totalOutput += ltcutil.Amount(txOut.Value)
	}
	return totalOutput
}

func (coin Ltc) GetAddressType(str string, testNet bool) (int8, error) {
	params := getLtcNetParams(testNet)
	decAddr, err := ltcutil.DecodeAddress(str, &params)
	if err != nil {
		return -1, err
	}

	switch decAddr.(type) {
	case *ltcutil.AddressPubKeyHash:
		return 0, nil
	case *ltcutil.AddressScriptHash:
		return 1, nil
	case *ltcutil.AddressPubKey:
		return 2, nil
	case *ltcutil.AddressWitnessPubKeyHash:
		return 3, nil
	case *ltcutil.AddressWitnessScriptHash:
		return 4, nil
	case *ltcutil.AddressTaproot:
		return 5, nil
	}
	return 6, nil
}

type LtcSecProvider struct {
	TestNet bool
	txscript.KeyDB
	txscript.ScriptDB
}

func (receiver LtcSecProvider) ChainParams() *chaincfg.Params {
	params := getLtcNetParams(receiver.TestNet)
	return &params
}
