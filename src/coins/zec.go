package coins

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcwallet/wallet/txrules"
	"strings"
	"wallet-sdk/src/coins/zecutil"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gcash/bchutil"
	"github.com/shopspring/decimal"
)

const CurrencyZec = "ZEC"

type ZecTxParams struct {
	BtcTxParams
	ZecLastBlockNum int `json:"zecLastBlockNum"`
}

var coinZec Zec

func init() {
	coinZec = Zec{}
	RegisterUtxoCoin(coinZec)
}

type Zec struct {
	Btc
}

func (coin Zec) GetCurrency() string {
	return CurrencyZec
}

func (coin Zec) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Zec) GetBasePath(testNet bool) string {
	return "m/44'/133'/%d'/%d/%d"
}

func (coin Zec) ChainName() string {
	return CurrencyZec
}

func (coin Zec) GetDecimal() int {
	return 8
}

func (coin Zec) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := bchutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}
func (coin Zec) PrivateKeyToString(key types.PrivateKey) (string, error) {
	btcPrivKey, _ := btcec.PrivKeyFromBytes(key)
	wif, err := btcutil.NewWIF(btcPrivKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}

func (coin Zec) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	ecdsaPrivatekey, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(ecdsaPrivatekey)
	_, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	params := getZecNetParams(testNet)
	encode, err := zecutil.Encode(pubKey.SerializeCompressed(), params.Params)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = encode
	return &adddress, nil
}

func (coin Zec) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Zec) GetEmptyTransactionParams() types.TxParams {
	return ZecTxParams{}
}

func (coin Zec) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {

	extraParams := txParams.(ZecTxParams)

	var unspends = extraParams.Unspends
	var receivers = extraParams.Receivers

	var changeAddress = extraParams.ChangeAddress

	var totalValuInputs = decimal.New(0, 0)
	params := getZecNetParams(testNet)
	newTx := wire.NewMsgTx(4)
	for _, unspend := range unspends {
		ph, err := chainhash.NewHashFromStr(
			unspend.TxHash,
		)
		if err != nil {
			return nil, errors.ErrorInvalidInput
		}
		outValue, err := btcutil.NewAmount(unspend.TxValue.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
		}
		outValueSatoshi := outValue.ToUnit(btcutil.AmountSatoshi)
		var netName string
		if testNet {
			netName = "testnet3"
		} else {
			netName = "mainnet"
		}
		address, err := zecutil.DecodeAddress(unspend.Address, netName)
		if err != nil {
			return nil, err
		}
		addrScript, err := zecutil.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}
		txIn := wire.TxIn{
			PreviousOutPoint: *wire.NewOutPoint(ph, unspend.TxOutputN),
			SignatureScript:  addrScript,
			Witness:          nil,
			Sequence:         uint32(outValueSatoshi),
		}
		newTx.AddTxIn(&txIn)
		totalValuInputs = totalValuInputs.Add(unspend.TxValue)
	}

	var need = decimal.New(0, 0)
	var outputAddrs []string

	for _, receiver := range receivers {

		var netName string
		if testNet {
			netName = "testnet3"
		} else {
			netName = "mainnet"
		}
		_, err := zecutil.DecodeAddress(receiver.Address, netName)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}

		decoded := base58.Decode(receiver.Address)
		var addr *btcutil.AddressPubKeyHash
		addr, err = btcutil.NewAddressPubKeyHash(decoded[2:len(decoded)-4], params.Params)

		if err != nil {
			return nil, err
		}
		outputAddrs = append(outputAddrs, receiver.Address)
		receiverPkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		outValue, err := btcutil.NewAmount(receiver.Value.InexactFloat64())
		if err != nil {
			return nil, err
		}
		outValueSatoshi := outValue.ToUnit(btcutil.AmountSatoshi)
		txOut := wire.NewTxOut(int64(outValueSatoshi), receiverPkScript)
		newTx.AddTxOut(txOut)
		need = need.Add(receiver.Value)
	}

	estimateSize := coin.EstimateSize(len(unspends), outputAddrs, true, testNet)
	fee, err := btcutil.NewAmount(extraParams.Fee.InexactFloat64())
	if err != nil {
		return nil, err
	}
	targetFee := txrules.FeeForSerializeSize(fee, estimateSize)
	toBTC := targetFee.ToBTC()

	if changeAddress != "" && &changeAddress != nil {
		change := totalValuInputs.Sub(need).Sub(decimal.NewFromFloat(toBTC))
		var netName string
		if testNet {
			netName = "testnet3"
		} else {
			netName = "mainnet"
		}
		_, err := zecutil.DecodeAddress(changeAddress, netName)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		change = totalValuInputs.Sub(need).Sub(decimal.NewFromFloat(extraParams.Fee.InexactFloat64()))
		outValue, err := btcutil.NewAmount(change.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
		}

		changeSatoshi := outValue.ToUnit(btcutil.AmountSatoshi)
		decoded := base58.Decode(changeAddress)
		addr, err := btcutil.NewAddressPubKeyHash(decoded[2:len(decoded)-4], params.Params)
		if err != nil {
			return nil, err
		}

		changePkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		txOut := wire.NewTxOut(int64(changeSatoshi), changePkScript)

		if changeSatoshi > 546 {
			newTx.AddTxOut(txOut)
		}
	}

	zecTx := &zecutil.MsgTx{
		MsgTx:        newTx,
		ExpiryHeight: uint32(extraParams.ZecLastBlockNum),
	}

	return &types.BaseTransaction{CoinTransaction: zecTx}, nil
}

func (coin Zec) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	netParams := getZecNetParams(testnet)
	return coin.DecodeTx(rawTx, *netParams.Params)
}

func (coin Zec) DecodeTx(rawTx string, params chaincfg.Params) (interface{}, error) {
	decodeString, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}

	var mtx = zecutil.MsgTx{}
	err = mtx.Deserialize(bytes.NewReader(decodeString))
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	rawTxBytes, err := hex.DecodeString(strings.TrimPrefix(rawTx, "0x"))
	if err != nil {
		return nil, err
	}
	firstSHA := sha256.Sum256(rawTxBytes)
	secondSHA := sha256.Sum256(firstSHA[:])

	txid := reverseBytes(secondSHA[:])
	txReply := btcjson.TxRawDecodeResult{
		Txid:     hex.EncodeToString(txid),
		Version:  mtx.Version,
		Locktime: mtx.LockTime,
		Vin:      coin.createVinList(mtx),
		Vout:     coin.createVoutList(mtx, &params, nil),
	}
	return txReply, nil
}

func (coin Zec) witnessToHex(witness wire.TxWitness) []string {
	if len(witness) == 0 {
		return nil
	}

	result := make([]string, 0, len(witness))
	for _, wit := range witness {
		result = append(result, hex.EncodeToString(wit))
	}

	return result
}

func (coin Zec) createVinList(mtx zecutil.MsgTx) []btcjson.Vin {
	vinList := make([]btcjson.Vin, len(mtx.TxIn))
	if zecutil.IsCoinBaseTx(mtx) {
		txIn := mtx.TxIn[0]
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		vinList[0].Witness = coin.witnessToHex(txIn.Witness)
		return vinList
	}

	for i, txIn := range mtx.TxIn {
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := &vinList[i]
		vinEntry.Txid = txIn.PreviousOutPoint.Hash.String()
		vinEntry.Vout = txIn.PreviousOutPoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &btcjson.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}

		if mtx.HasWitness() {
			vinEntry.Witness = coin.witnessToHex(txIn.Witness)
		}
	}

	return vinList
}

func (coin Zec) createVoutList(mtx zecutil.MsgTx, chainParams *chaincfg.Params, filterAddrMap map[string]struct{}) []btcjson.Vout {
	voutList := make([]btcjson.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		disbuf, _ := txscript.DisasmString(v.PkScript)
		scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(
			v.PkScript, chainParams)
		passesFilter := len(filterAddrMap) == 0
		encodedAddrs := make([]string, len(addrs))

		for j, addr := range addrs {
			encodedAddr := addr.EncodeAddress()
			encodedAddrs[j], _ = zecutil.ParsePkScript(v.PkScript, chainParams)
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
		vout.Value = btcutil.Amount(v.Value).ToBTC()
		vout.ScriptPubKey.Addresses = encodedAddrs
		vout.ScriptPubKey.Asm = disbuf
		vout.ScriptPubKey.Hex = hex.EncodeToString(v.PkScript)
		vout.ScriptPubKey.Type = scriptClass.String()
		vout.ScriptPubKey.ReqSigs = int32(reqSigs)

		voutList = append(voutList, vout)
	}

	return voutList
}

func (coin Zec) SignTx(baseTransaction *types.BaseTransaction, testNet bool, keyByte types.PrivateKey) (*string, error) {
	transaction := baseTransaction.CoinTransaction.(*zecutil.MsgTx)

	params := getZecNetParams(testNet)

	ecdsaPrivatekey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(ecdsaPrivatekey)
	btcPrivKey, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	encode, err := zecutil.Encode(pubKey.SerializeCompressed(), params.Params)
	if err != nil {
		return nil, err
	}

	wif, err := btcutil.NewWIF(btcPrivKey, params.Params, true)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	var netName string
	if testNet {
		netName = "testnet3"
	} else {
		netName = "mainnet"
	}
	address, err := zecutil.DecodeAddress(encode, netName)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	addrScript, err := zecutil.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}

	for index, txIn := range transaction.TxIn {
		netParams := getZecNetParams(testNet)
		sigScript, err := zecutil.SignTxOutput(
			netParams.Params,
			transaction,
			index,
			addrScript,
			txscript.SigHashAll,
			txscript.KeyClosure(func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
				return wif.PrivKey, wif.CompressPubKey, nil
			}),
			nil,
			nil,
			int64(txIn.Sequence))
		if err != nil {
			return nil, err
		}
		txIn.SignatureScript = sigScript
	}
	var buf bytes.Buffer
	if err = transaction.ZecEncode(&buf, 0, wire.BaseEncoding); err != nil {
		return nil, err
	}
	sprintf := fmt.Sprintf("%x", buf.Bytes())
	return &sprintf, nil
}

func (coin Zec) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error) {
	transaction := baseTransaction.CoinTransaction.(*zecutil.MsgTx)
	params := getZecNetParams(testNet)
	var netName string
	if testNet {
		netName = "testnet3"
	} else {
		netName = "mainnet"
	}
	for index, txIn := range transaction.TxIn {
		netParams := getZecNetParams(testNet)

		sigScript, err := zecutil.SignTxOutput(
			netParams.Params,
			transaction,
			index,
			txIn.SignatureScript,
			txscript.SigHashAll,
			txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
				bool, error) {
				var privateKeyBytes types.PrivateKey
				for addrString, key := range keys {
					address, err := zecutil.DecodeAddress(addrString, netName)
					if err != nil {
						return nil, false, err
					}
					addrScript, err := zecutil.PayToAddrScript(address)
					if err != nil {
						return nil, false, err
					}
					_, addrs, _, err := txscript.ExtractPkScriptAddrs(addrScript, netParams.Params)
					if err != nil {
						return nil, false, err
					}
					if addr.EncodeAddress() == addrs[0].EncodeAddress() {
						privateKeyBytes = key
						break
					}
				}
				if privateKeyBytes == nil {
					return nil, false, errors.ErrorKeyNotFound
				}

				privateKey, err := crypto.ToECDSA(privateKeyBytes)
				if err != nil {
					return nil, false, err
				}
				privateKeyBytes = crypto.FromECDSA(privateKey)

				btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

				wif, err := btcutil.NewWIF(btcPrivKey, params.Params, true)
				if err != nil {
					return nil, false, err
				}

				return wif.PrivKey, wif.CompressPubKey, nil
			}),
			txscript.ScriptClosure(func(
				addr btcutil.Address) ([]byte, error) {
				address, err := zecutil.DecodeAddress(addr.EncodeAddress(), netName)
				if err != nil {
					return nil, err
				}
				addrScript, err := zecutil.PayToAddrScript(address)
				if err != nil {
					return nil, err
				}
				return addrScript, nil
			}),
			nil,
			int64(txIn.Sequence))
		if err != nil {
			return nil, err
		}
		txIn.SignatureScript = sigScript
	}
	var buf bytes.Buffer
	if err := transaction.ZecEncode(&buf, 0, wire.BaseEncoding); err != nil {
		return nil, err
	}

	sprintf := fmt.Sprintf("%x", buf.Bytes())
	return &sprintf, nil
}

func (coin Zec) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := ZecTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Zec) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := getZecNetParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, *params.Params)

}

func getZecNetParams(testNet bool) zecutil.Params {
	var params zecutil.Params

	if testNet {
		params = zecutil.TestNet3Params
	} else {
		params = zecutil.MainNetParams
	}
	return params
}

func reverseBytes(data []byte) []byte {
	reversed := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		reversed[i] = data[len(data)-1-i]
	}
	return reversed
}
