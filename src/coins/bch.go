package coins

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/schancel/cashaddr-converter/cashaddress"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/blockchain"
	"github.com/gcash/bchd/btcjson"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"github.com/gcash/bchutil"
	"github.com/gcash/bchwallet/wallet/txauthor"
	"github.com/gcash/bchwallet/wallet/txrules"
	"github.com/gcash/bchwallet/wallet/txsizes"
	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/schancel/cashaddr-converter/address"
	"github.com/schancel/cashaddr-converter/legacy"
)

const CurrencyBch = "BCH"

var coinBch Bch

func init() {
	coinBch = Bch{}
	RegisterUtxoCoin(coinBch)
}

type Bch struct {
}

func (coin Bch) GetCurrency() string {
	return CurrencyBch
}

func (coin Bch) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Bch) GetBasePath(testNet bool) string {
	return "m/44'/145'/%d'/%d/%d"
}

func (coin Bch) ChainName() string {
	return CurrencyBch
}

func (coin Bch) GetDecimal() int {
	return 8
}

func (coin Bch) PrivateKeyToString(key types.PrivateKey) (string, error) {
	privateKey, err := crypto.ToECDSA(key)

	btcPrivKey, _ := bchec.PrivKeyFromBytes(privateKey.PublicKey.Curve, key)
	wif, err := bchutil.NewWIF(btcPrivKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}

func (coin Bch) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := bchutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}

func (coin Bch) CashAddressToLegacyAddress(cashAddress string) (*string, error) {
	fromString, err := address.NewFromString(cashAddress)
	if err != nil {
		return nil, err
	}
	lagacyAddress, err := fromString.Legacy()
	if err != nil {
		return nil, err
	}
	encode, err := lagacyAddress.Encode()
	if err != nil {
		return nil, err
	}
	return &encode, nil

}

func (coin Bch) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	_, pubKey := bchec.PrivKeyFromBytes(privateKey.PublicKey.Curve, keyByte)
	netParams := coin.getNetParams(testNet)
	if err != nil {
		return nil, err
	}
	newAddressPubKey, err := bchutil.NewAddressPubKey(pubKey.SerializeCompressed(), &netParams)
	if err != nil {
		return nil, err
	}

	decode, err := legacy.Decode(newAddressPubKey.EncodeAddress())

	if err != nil {
		return nil, err
	}
	fromLegacy, err := address.NewFromLegacy(decode)
	if err != nil {
		return nil, err
	}
	cashAddress, err := fromLegacy.CashAddress()
	var adddress = types.CoinAddress{}
	adddress.AddressStr = cashAddress.String()
	return &adddress, nil
}

func (coin Bch) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Bch) GetEmptyTransactionParams() types.TxParams {
	return BtcTxParams{}
}

func (coin Bch) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := txParams.(BtcTxParams)
	var unspends = extraParams.Unspends
	var receivers = extraParams.Receivers
	changeAddress := extraParams.ChangeAddress
	var totalAmount = bchutil.Amount(0)
	params := coin.getNetParams(testNet)
	currentInputs := make([]*wire.TxIn, 0, len(unspends))
	currentInputValues := make([]bchutil.Amount, 0, len(unspends))
	var inputScripts [][]byte
	for _, unspend := range unspends {

		amount, err := bchutil.NewAmount(unspend.TxValue.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
		}
		totalAmount += amount
		hash, err := chainhash.NewHashFromStr(unspend.TxHash)
		if err != nil {
			return nil, errors.ErrorInvalidInput
		}
		nextInput := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: unspend.TxOutputN,
		}, nil)
		inputaddr, err := bchutil.DecodeAddress(unspend.Address, &params)
		if err != nil {
			return nil, err
		}
		script, err := txscript.PayToAddrScript(inputaddr)
		if err != nil {
			return nil, err
		}
		inputScripts = append(inputScripts, script)
		currentInputs = append(currentInputs, nextInput)
		currentInputValues = append(currentInputValues, amount)
	}
	feeAmount, err := bchutil.NewAmount(extraParams.Fee.InexactFloat64())
	if err != nil {
		return nil, errors.ErrorInvalidAmount
	}
	var txOut []*wire.TxOut
	for _, receiver := range receivers {
		decAddr, err := bchutil.DecodeAddress(receiver.Address, &params)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		a, err := bchutil.NewAmount(receiver.Value.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
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

	inputSource := func(target bchutil.Amount) (total bchutil.Amount, inputs []*wire.TxIn, inputValues []bchutil.Amount, scripts [][]byte, err error) {
		return totalAmount, currentInputs, currentInputValues, inputScripts, nil
	}

	changeSource := func() ([]byte, error) {
		if changeAddress == "" {
			return nil, nil
		} else {
			changeAddr, err := bchutil.DecodeAddress(changeAddress, &params)
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

	unsignedTransaction, err := coin.NewUnsignedTransaction(currentInputs, txOut, feeAmount, inputSource, changeSource, changeAddress != "")
	if err != nil {
		return nil, err
	}

	return &types.BaseTransaction{CoinTransaction: unsignedTransaction}, nil
}

func (coin Bch) NewUnsignedTransaction(inputs []*wire.TxIn, outputs []*wire.TxOut, relayFeePerKb bchutil.Amount, fetchInputs txauthor.InputSource, fetchChange txauthor.ChangeSource, hasChange bool) (*txauthor.AuthoredTx, error) {

	targetAmount := SumOutputValues(outputs)
	estimatedSize := txsizes.EstimateSerializeSize(len(inputs), outputs, hasChange)
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
			if changeAmount != 0 && !txrules.IsDustAmount(changeAmount,
				txsizes.P2PKHPkScriptSize, relayFeePerKb) {
				changeScript, err := fetchChange()
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

func SumOutputValues(outputs []*wire.TxOut) (totalOutput bchutil.Amount) {
	for _, txOut := range outputs {
		totalOutput += bchutil.Amount(txOut.Value)
	}
	return totalOutput
}

func (coin Bch) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := coin.getNetParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, params)

}

func (coin Bch) EstimateTxSizes(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet chaincfg.Params) int {
	changeSize := 0
	outputCount := len(outputAddrs)
	if hasExtraChangeAddr {
		changeSize = txsizes.P2PKHOutputSize
		outputCount++
	}

	//8 additional bytes are for version and locktime
	return 8 + wire.VarIntSerializeSize(uint64(inputCount)) +
		wire.VarIntSerializeSize(uint64(outputCount)) +
		inputCount*txsizes.RedeemP2PKHInputSize +
		coin.SumOutputSerializeSizesOfChainParams(outputAddrs, testNet) +
		changeSize
}

func (coin Bch) SumOutputSerializeSizesOfChainParams(outputAddrs []string, params chaincfg.Params) int {
	var sizeSum = 0
	for _, addr := range outputAddrs {
		decodeAddress, err := bchutil.DecodeAddress(addr, &params)
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

func (coin Bch) DecodeTransaction(hexStr string, testnet bool) (interface{}, error) {
	params := coin.getNetParams(testnet)
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	// Create and return the result.
	txReply := btcjson.TxRawDecodeResult{
		Txid:     mtx.TxHash().String(),
		Version:  mtx.Version,
		Locktime: mtx.LockTime,
		Vin:      coin.createVinList(&mtx),
		Vout:     coin.createVoutList(&mtx, &params, nil),
	}
	return txReply, nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed
// transaction.
func (coin Bch) createVinList(mtx *wire.MsgTx) []btcjson.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]btcjson.Vin, len(mtx.TxIn))
	if blockchain.IsCoinBaseTx(mtx) {
		txIn := mtx.TxIn[0]
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		return vinList
	}

	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := &vinList[i]
		vinEntry.Txid = txIn.PreviousOutPoint.Hash.String()
		vinEntry.Vout = txIn.PreviousOutPoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &btcjson.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
	}

	return vinList
}

func (coin Bch) createVoutList(mtx *wire.MsgTx, chainParams *chaincfg.Params, filterAddrMap map[string]struct{}) []btcjson.Vout {
	voutList := make([]btcjson.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		disbuf, _ := txscript.DisasmString(v.PkScript)
		scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(
			v.PkScript, chainParams)
		passesFilter := len(filterAddrMap) == 0
		encodedAddrs := make([]string, len(addrs))
		for j, addr := range addrs {
			encodedAddr := addr.EncodeAddress()
			encodedAddrs[j] = encodedAddr

			// No need to check the map again if the filter already
			// passes.
			if passesFilter {
				continue
			}
			if _, exists := filterAddrMap[encodedAddr]; exists {
				passesFilter = true
			}
		}

		if !passesFilter {
			continue
		}

		var vout btcjson.Vout
		vout.N = uint32(i)
		vout.Value = bchutil.Amount(v.Value).ToBCH()
		vout.ScriptPubKey.Addresses = encodedAddrs
		vout.ScriptPubKey.Asm = disbuf
		vout.ScriptPubKey.Hex = hex.EncodeToString(v.PkScript)
		vout.ScriptPubKey.Type = scriptClass.String()
		vout.ScriptPubKey.ReqSigs = int32(reqSigs)

		voutList = append(voutList, vout)
	}

	return voutList
}

func (coin Bch) SignTx(baseTransaction *types.BaseTransaction, testNet bool, derivedKey types.PrivateKey) (*string, error) {
	authoredTx := baseTransaction.CoinTransaction.(*txauthor.AuthoredTx)
	//signTransaction, err := btc2.SignTransaction(transaction, privateKey, GetNetParams(testNet), true)
	netParams := coin.getNetParams(testNet)
	privateKey, err := crypto.ToECDSA(derivedKey)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	pKey, _ := bchec.PrivKeyFromBytes(privateKey.PublicKey.Curve, privateKeyBytes)
	wif, err := bchutil.NewWIF(pKey, &netParams, true)
	key, err := bchutil.NewAddressPubKeyHash(bchutil.Hash160(wif.SerializePubKey()), &netParams)

	if err != nil {
		return nil, err
	}

	script, err := txscript.PayToAddrScript(key)
	if err != nil {
		return nil, err
	}
	mkGetKey := func() txscript.KeyDB {

		return txscript.KeyClosure(func(addr bchutil.Address) (*bchec.PrivateKey,
			bool, error) {
			return wif.PrivKey, wif.CompressPubKey, nil
		})
	}

	getScript := txscript.ScriptClosure(func(
		addr bchutil.Address) ([]byte, error) {

		return txscript.PayToAddrScript(key)
	})

	for i, txIn := range authoredTx.Tx.TxIn {
		output, err := txscript.SignTxOutput(&netParams, authoredTx.Tx,
			i,
			int64(authoredTx.PrevInputValues[i]),
			script,
			txscript.SigHashAll,
			mkGetKey(),
			getScript,
			nil,
		)

		txIn.SignatureScript = output
		if err != nil {
			return nil, err
		}
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

func (coin Bch) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error) {

	return coin.SignMultipleSendAddress(baseTransaction, keys, testNet)
}

func (coin Bch) SignMultipleSendAddress(tx *types.BaseTransaction, keys map[string]types.PrivateKey, testNet bool) (*string, error) {
	netParams := coin.getNetParams(testNet)
	authoredTx := tx.CoinTransaction.(*txauthor.AuthoredTx)

	mkGetKey := func() txscript.KeyDB {

		return txscript.KeyClosure(func(addr bchutil.Address) (*bchec.PrivateKey,
			bool, error) {

			derivedKey := keys[addr.EncodeAddress()]
			if derivedKey == nil {
				var prefix = cashaddress.TestNet
				if !testNet {
					prefix = cashaddress.MainNet
				}
				derivedKey = keys[prefix+":"+addr.EncodeAddress()]
			}
			if derivedKey == nil {
				var prefix = cashaddress.TestNet
				if !testNet {
					prefix = cashaddress.MainNet
				}

				decode, err := cashaddress.Decode(addr.EncodeAddress(), prefix)
				if err != nil {
					return nil, false, err
				}
				legacyAddress, err := coin.CashAddressToLegacyAddress(decode.String())
				if err != nil {
					return nil, false, err
				}
				derivedKey = keys[*legacyAddress]
			}

			privateKey, err := crypto.ToECDSA(derivedKey)
			if err != nil {
				return nil, false, err
			}
			privateKeyBytes := crypto.FromECDSA(privateKey)

			btcPrivKey, _ := bchec.PrivKeyFromBytes(secp256k1.S256(), privateKeyBytes)

			wif, err := bchutil.NewWIF(btcPrivKey, &netParams, true)
			if err != nil {
				return nil, false, err
			}

			return wif.PrivKey, wif.CompressPubKey, nil
		})
	}

	getScript := txscript.ScriptClosure(func(
		addr bchutil.Address) ([]byte, error) {
		derivedKey := keys[addr.EncodeAddress()]

		privateKey, err := crypto.ToECDSA(derivedKey)
		if err != nil {
			return nil, err
		}
		privateKeyBytes := crypto.FromECDSA(privateKey)

		btcPrivKey, _ := bchec.PrivKeyFromBytes(secp256k1.S256(), privateKeyBytes)

		wif, err := bchutil.NewWIF(btcPrivKey, &netParams, true)
		if err != nil {
			return nil, err
		}
		key, err := bchutil.NewAddressPubKeyHash(ltcutil.Hash160(wif.SerializePubKey()), &netParams)
		if err != nil {
			return nil, err
		}
		return txscript.PayToAddrScript(key)
	})

	source := BchSecProvider{&netParams,
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

func (coin Bch) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := BtcTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Bch) getNetParams(testNet bool) chaincfg.Params {
	var params chaincfg.Params

	if testNet {
		params = chaincfg.TestNet3Params
	} else {
		params = chaincfg.MainNetParams
	}
	return params
}

type BchSecProvider struct {
	Params *chaincfg.Params
	txscript.KeyDB
	txscript.ScriptDB
}

func (receiver BchSecProvider) ChainParams() *chaincfg.Params {
	return receiver.Params
}
