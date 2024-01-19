package coins

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gcash/bchutil"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/chaincfg"
	wif2 "github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
	"github.com/shopspring/decimal"
	"wallet-sdk/src/deriver"
	errors2 "wallet-sdk/src/errors"
	"wallet-sdk/src/types"

	btcChaincfg "github.com/btcsuite/btcd/chaincfg"
)

const CurrencyBsv = "BSV"

var coinBsv Bsv

func init() {
	coinBsv = Bsv{}
	RegisterUtxoCoin(coinBsv)
}

type Bsv struct {
	Btc
}

func (coin Bsv) GetCurrency() string {
	return CurrencyBsv
}

func (coin Bsv) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Bsv) GetBasePath(testNet bool) string {
	return "m/44'/236'/%d'/%d/%d"
}

func (coin Bsv) ChainName() string {
	return CurrencyBsv
}

func (coin Bsv) GetDecimal() int {
	return 8
}

func (coin Bsv) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := bchutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}

func (coin Bsv) PrivateKeyToString(key types.PrivateKey) (string, error) {
	//btcPrivKey, _ := bchec.PrivKeyFromBytes(bsvec.S256(), key)
	bytes, _ := bec.PrivKeyFromBytes(bec.S256(), key)
	wif, err := wif2.NewWIF(bytes, &chaincfg.MainNet, true)
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}

func (coin Bsv) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {

	bytes, pubKey := bec.PrivKeyFromBytes(bec.S256(), privateKey)
	params := coin.getNetParams(testNet)
	wif, err := wif2.NewWIF(bytes, &params, true)
	if err != nil {
		return nil, err
	}

	address, err := bscript.NewAddressFromPublicKey(pubKey, !testNet)
	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = address.AddressString
	adddress.PrivateKey = wif.String()
	return &adddress, nil
}

func (coin Bsv) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Bsv) GetEmptyTransactionParams() types.TxParams {
	return BtcTxParams{}
}

func (coin Bsv) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := txParams.(BtcTxParams)
	var unspends = extraParams.Unspends
	var receivers = extraParams.Receivers
	changeAddress := extraParams.ChangeAddress
	tx := bt.NewTx()
	var utxos bt.UTXOs
	for _, unspend := range unspends {
		fundingScript, _ := bscript.NewP2PKHFromAddress(unspend.Address)
		pti, err := hex.DecodeString(unspend.TxHash)
		if err != nil {
			return nil, err
		}
		satoshiValue := unspend.TxValue.Mul(decimal.NewFromFloat(10).Pow(decimal.NewFromInt(8)))
		utxos = append(utxos,
			&bt.UTXO{
				TxID:          pti,
				Vout:          unspend.TxOutputN,
				LockingScript: fundingScript,
				Satoshis:      uint64(satoshiValue.IntPart()),
			},
		)
	}
	err := tx.FromUTXOs(utxos...)
	if err != nil {
		return nil, err
	}
	for _, receiver := range receivers {
		satoshiValue := receiver.Value.Mul(decimal.NewFromFloat(10).Pow(decimal.NewFromInt(8)))
		err := tx.PayToAddress(receiver.Address, uint64(satoshiValue.IntPart()))
		if err != nil {
			return nil, errors2.ErrorInvalidAddress
		}
	}

	if changeAddress != "" {
		satoshiValue := extraParams.Fee.Mul(decimal.NewFromFloat(10).Pow(decimal.NewFromInt(8))).IntPart()
		var FQPoint5SatPerByte = bt.NewFeeQuote().
			AddQuote(bt.FeeTypeStandard, &bt.Fee{
				FeeType: bt.FeeTypeStandard,
				MiningFee: bt.FeeUnit{
					Satoshis: int(satoshiValue),
					Bytes:    1000,
				},
				RelayFee: bt.FeeUnit{
					Satoshis: int(satoshiValue),
					Bytes:    1000,
				},
			}).AddQuote(bt.FeeTypeData, &bt.Fee{
			FeeType: bt.FeeTypeData,
			MiningFee: bt.FeeUnit{
				Satoshis: int(satoshiValue),
				Bytes:    1000,
			},
			RelayFee: bt.FeeUnit{
				Satoshis: int(satoshiValue),
				Bytes:    1000,
			},
		})
		err := tx.ChangeToAddress(changeAddress, FQPoint5SatPerByte)
		if err != nil {
			return nil, errors2.ErrorInvalidAddress
		}

	}

	return &types.BaseTransaction{CoinTransaction: tx}, nil
}

func (coin Bsv) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	tx := baseTransaction.CoinTransaction.(*bt.Tx)
	params := coin.getNetParams(testNet)

	bytes, _ := bec.PrivKeyFromBytes(bec.S256(), privateKey)
	wif, err := wif2.NewWIF(bytes, &params, true)
	if err != nil {
		return nil, err
	}
	if err := tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: wif.PrivKey}); err != nil {
		return nil, err
	}
	s := tx.String()
	return &s, nil

}

func (coin Bsv) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error) {
	params := coin.getNetParams(testNet)
	tx := baseTransaction.CoinTransaction.(*bt.Tx)

	var getter = KeyGetter{
		keys:   keys,
		params: params,
	}

	if err := tx.FillAllInputs(context.Background(), getter); err != nil {
		return nil, err
	}
	s := tx.String()
	return &s, nil
}

type KeyGetter struct {
	keys   map[string]types.PrivateKey
	params chaincfg.Params
}

func (getter KeyGetter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error) {
	addresses, err := lockingScript.Addresses()
	if err != nil {
		return nil, err
	}
	var privateKey types.PrivateKey
	for addr, key := range getter.keys {
		fundingScript, _ := bscript.NewP2PKHFromAddress(addr)
		strings, err := fundingScript.Addresses()
		if err != nil {
			return nil, err
		}
		if addresses[0] == strings[0] {
			privateKey = key
		}
	}
	if privateKey == nil {
		return nil, errors.New("key not found")
	}
	keyBytes, _ := bec.PrivKeyFromBytes(bec.S256(), privateKey)
	wif, err := wif2.NewWIF(keyBytes, &getter.params, true)
	if err != nil {
		return nil, err
	}
	return &unlocker.Simple{PrivateKey: wif.PrivKey}, nil
}

//func (coin Bsv) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
//	//params := coin.getNetParams(testnet)
//	tx, err := bt.NewTxFromString(rawTx)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return tx, nil
//
//}

func (coin Bsv) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := coin.GetNetParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, params)

}

func (coin Bsv) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {

	var params btcChaincfg.Params

	if testnet {
		params = BsvTestNet3Params()
	} else {
		params = BsvMainNetParams()
	}

	return coin.DecodeTx(rawTx, params)

}

func (coin Bsv) DecodeTx(rawTx string, params btcChaincfg.Params) (interface{}, error) {
	decodeString, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}

	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(decodeString))
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

// witnessToHex formats the passed witness stack as a slice of hex-encoded
// strings to be used in a JSON response.
func (coin Bsv) witnessToHex(witness wire.TxWitness) []string {
	// Ensure nil is returned when there are no entries versus an empty
	// slice so it can properly be omitted as necessary.
	if len(witness) == 0 {
		return nil
	}

	result := make([]string, 0, len(witness))
	for _, wit := range witness {
		result = append(result, hex.EncodeToString(wit))
	}

	return result
}

func (coin Bsv) createVinList(mtx *wire.MsgTx) []btcjson.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]btcjson.Vin, len(mtx.TxIn))
	if blockchain.IsCoinBaseTx(mtx) {
		txIn := mtx.TxIn[0]
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		vinList[0].Witness = coin.witnessToHex(txIn.Witness)
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

		if mtx.HasWitness() {
			vinEntry.Witness = coin.witnessToHex(txIn.Witness)
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func (coin Bsv) createVoutList(mtx *wire.MsgTx, chainParams *btcChaincfg.Params, filterAddrMap map[string]struct{}) []btcjson.Vout {
	voutList := make([]btcjson.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.PkScript)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(
			v.PkScript, chainParams)

		// Encode the addresses while checking if the address passes the
		// filter when needed.
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

func (coin Bsv) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := BtcTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Bsv) getNetParams(testNet bool) chaincfg.Params {
	var params chaincfg.Params

	if testNet {
		params = chaincfg.TestNet
	} else {
		params = chaincfg.MainNet
	}
	return params
}
