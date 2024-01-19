package coins

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcwallet/wallet/txrules"
	"github.com/btcsuite/btcwallet/wallet/txsizes"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

const (
	MinNondustOutput = 546        // satoshis
	omniHex          = "6f6d6e69" // Hex-encoded: "omni"
	CurrencyBtc      = "BTC"
)

type BtcTxParams struct {
	types.BaseTxParams
	Unspends      []Unspent       `json:"unspends"`
	Receivers     []Receiver      `json:"receivers"`
	ChangeAddress string          `json:"changeAddress"`
	Fee           decimal.Decimal `json:"fee"`
}

var coinBtc Btc

func init() {
	coinBtc = Btc{}
	RegisterUtxoCoin(coinBtc)
}

type Btc struct {
}

func (coin Btc) GetCurrency() string {
	return CurrencyBtc
}

func (coin Btc) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Btc) GetBasePath(testNet bool) string {
	if testNet {
		return "m/44'/1'/%d'/%d/%d"
	}
	return "m/44'/0'/%d'/%d/%d"
}

func (coin Btc) ChainName() string {
	return CurrencyBtc
}

func (coin Btc) GetDecimal() int {
	return 8
}

func (coin Btc) PrivateKeyToString(key types.PrivateKey) (string, error) {
	btcPrivKey, _ := btcec.PrivKeyFromBytes(key)
	wif, err := btcutil.NewWIF(btcPrivKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}

func (coin Btc) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}

func (coin Btc) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.GenAddress(keyByte, netParams)
}

func (coin Btc) GenAddress(keyByte []byte, netParams chaincfg.Params) (*types.CoinAddress, error) {
	_, pubKey := btcec.PrivKeyFromBytes(keyByte)

	newAddressPubKey, err := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(), &netParams)
	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = newAddressPubKey.EncodeAddress()
	return &adddress, nil
}

func (coin Btc) GenerateTaprootAddressByPrivateKey(key string, testnet bool) (*types.CoinAddress, error) {
	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}

	privKey := wif.PrivKey

	toECDSA, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Serialize()))
	if err != nil {
		return nil, err
	}
	return coin.GenerateTaprootAddress(toECDSA.D.Bytes(), testnet)
}

func (coin Btc) GenerateTaprootAddress(keyByte []byte, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	coinPointer := &coin
	netParams := coinPointer.GetNetParams(testNet)
	wif, err := btcutil.NewWIF(btcPrivKey, &netParams, true)
	if err != nil {
		return nil, err
	}

	utxoTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(wif.PrivKey.PubKey())), &netParams)

	segwitNested := utxoTaprootAddress.EncodeAddress()

	fmt.Println(segwitNested)

	var adddress = types.CoinAddress{}
	adddress.AddressStr = segwitNested
	return &adddress, nil
}

func (coin Btc) GenerateNestedSegitAddressByPrivateKey(key string, testnet bool) (*types.CoinAddress, error) {
	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}

	privKey := wif.PrivKey

	toECDSA, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Serialize()))
	if err != nil {
		return nil, err
	}
	return coin.GenerateNestedSegitAddress(toECDSA.D.Bytes(), testnet)
}

func (coin Btc) GenerateNestedSegitAddress(keyByte []byte, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	coinPointer := &coin
	netParams := coinPointer.GetNetParams(testNet)
	wif, err := btcutil.NewWIF(btcPrivKey, &netParams, true)
	if err != nil {
		return nil, err
	}

	witnessProg := btcutil.Hash160(wif.SerializePubKey())

	addressWitnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, &netParams)
	if err != nil {
		return nil, err
	}

	serializedScript, err := txscript.PayToAddrScript(addressWitnessPubKeyHash)

	if err != nil {
		return nil, err
	}
	addressScriptHash, err := btcutil.NewAddressScriptHash(serializedScript, &netParams)

	if err != nil {
		return nil, err
	}
	segwitNested := addressScriptHash.EncodeAddress()

	var adddress = types.CoinAddress{}
	adddress.AddressStr = segwitNested
	return &adddress, nil
}

func (coin Btc) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Btc) GetEmptyTransactionParams() types.TxParams {
	return BtcTxParams{}
}

func (coin Btc) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.Transaction(txParams, netParams)
}

func (coin Btc) Transaction(txParams types.TxParams, netParams chaincfg.Params) (*types.BaseTransaction, error) {
	extraParams := txParams.(BtcTxParams)
	var unspends = extraParams.Unspends
	var receivers = extraParams.Receivers
	changeAddress := extraParams.ChangeAddress
	var totalAmount = btcutil.Amount(0)

	var currentInputs []*wire.TxIn
	var currentInputValues []btcutil.Amount
	var inputScripts [][]byte
	for _, unspend := range unspends {

		amount, err := btcutil.NewAmount(unspend.TxValue.InexactFloat64())
		if err != nil {
			return nil, errors.ErrorInvalidAmount
		}
		totalAmount += amount
		address, err := btcutil.DecodeAddress(unspend.Address, &netParams)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		script, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}
		inputScripts = append(inputScripts, script)
		hash, err := chainhash.NewHashFromStr(unspend.TxHash)
		nextInput := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: unspend.TxOutputN,
		}, nil, nil)
		currentInputs = append(currentInputs, nextInput)
		currentInputValues = append(currentInputValues, amount)
	}

	//changeAddr, err := btcutil.DecodeAddress(fromAddress, &netParams)

	feeAmount, err := btcutil.NewAmount(extraParams.Fee.InexactFloat64())
	if err != nil {
		return nil, errors.ErrorInvalidAmount
	}
	var txOut []*wire.TxOut
	for _, receiver := range receivers {
		decAddr, err := btcutil.DecodeAddress(receiver.Address, &netParams)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		a, err := btcutil.NewAmount(receiver.Value.InexactFloat64())
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
	inputSource := func(target btcutil.Amount) (total btcutil.Amount, inputs []*wire.TxIn, inputValues []btcutil.Amount, scripts [][]byte, err error) {
		return totalAmount, currentInputs, currentInputValues, inputScripts, nil
	}

	changeSource := txauthor.ChangeSource{}

	changeSource.NewScript = func() ([]byte, error) {
		if changeAddress == "" {
			return nil, nil
		} else {
			changeAddr, err := btcutil.DecodeAddress(changeAddress, &netParams)
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

func (coin Btc) NewUnsignedTransaction(inputs []*wire.TxIn, outputs []*wire.TxOut, relayFeePerKb btcutil.Amount,
	fetchInputs txauthor.InputSource, fetchChange *txauthor.ChangeSource, needChange bool) (*txauthor.AuthoredTx, error) {
	targetAmount := coin.SumOutputValues(outputs)
	estimatedSize := txsizes.EstimateSerializeSize(len(inputs), outputs, true)
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
			if changeAmount != 0 && !IsDustAmount(changeAmount) {
				changeScript, err := fetchChange.NewScript()
				if changeScript != nil {
					if err != nil {
						return nil, err
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

func IsDustAmount(amount btcutil.Amount) bool {
	return amount < MinNondustOutput
}

func (coin Btc) SumOutputValues(outputs []*wire.TxOut) (totalOutput btcutil.Amount) {
	for _, txOut := range outputs {
		totalOutput += btcutil.Amount(txOut.Value)
	}
	return totalOutput
}

func (coin Btc) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	params := coin.GetNetParams(testnet)
	return coin.DecodeTx(rawTx, params)

}

func (coin Btc) DecodeTx(rawTx string, params chaincfg.Params) (interface{}, error) {
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

	txReply := btcjson.TxRawDecodeResult{
		Txid:     mtx.TxHash().String(),
		Version:  mtx.Version,
		Locktime: mtx.LockTime,
		Vin:      coin.createVinList(&mtx),
		Vout:     coin.createVoutList(&mtx, &params, nil),
	}
	return txReply, nil
}

func (coin Btc) witnessToHex(witness wire.TxWitness) []string {
	if len(witness) == 0 {
		return nil
	}

	result := make([]string, 0, len(witness))
	for _, wit := range witness {
		result = append(result, hex.EncodeToString(wit))
	}

	return result
}

func (coin Btc) createVinList(mtx *wire.MsgTx) []btcjson.Vin {
	vinList := make([]btcjson.Vin, len(mtx.TxIn))
	if blockchain.IsCoinBaseTx(mtx) {
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

func (coin Btc) createVoutList(mtx *wire.MsgTx, chainParams *chaincfg.Params, filterAddrMap map[string]struct{}) []btcjson.Vout {
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

func (coin Btc) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.Sign(baseTransaction, privateKey, netParams)
}

func (coin Btc) Sign(tx *types.BaseTransaction, derivedKey types.PrivateKey, netParams chaincfg.Params) (*string, error) {
	authoredTx := tx.CoinTransaction.(*txauthor.AuthoredTx)
	privateKey, err := crypto.ToECDSA(derivedKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	wif, err := btcutil.NewWIF(btcPrivKey, &netParams, true)
	if err != nil {
		return nil, err
	}

	key, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.SerializePubKey()), &netParams)
	if err != nil {
		return nil, err
	}

	mkGetKey := func() txscript.KeyDB {

		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
			bool, error) {
			return wif.PrivKey, wif.CompressPubKey, nil
		})
	}

	getScript := txscript.ScriptClosure(func(
		addr btcutil.Address) ([]byte, error) {

		return txscript.PayToAddrScript(key)
	})

	source := SecProvider{&netParams,
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
func (coin Btc) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.SignMultipleSendAddress(baseTransaction, keys, netParams)
}

func (coin Btc) SignMultipleSendAddress(tx *types.BaseTransaction, keys map[string]types.PrivateKey, netParams chaincfg.Params) (*string, error) {
	authoredTx := tx.CoinTransaction.(*txauthor.AuthoredTx)

	mkGetKey := func() txscript.KeyDB {

		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
			bool, error) {

			derivedKey := keys[addr.EncodeAddress()]

			privateKey, err := crypto.ToECDSA(derivedKey)
			if err != nil {
				return nil, false, err
			}
			privateKeyBytes := crypto.FromECDSA(privateKey)

			btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

			wif, err := btcutil.NewWIF(btcPrivKey, &netParams, true)
			if err != nil {
				return nil, false, err
			}

			return wif.PrivKey, wif.CompressPubKey, nil
		})
	}

	getScript := txscript.ScriptClosure(func(
		addr btcutil.Address) ([]byte, error) {
		derivedKey := keys[addr.EncodeAddress()]

		privateKey, err := crypto.ToECDSA(derivedKey)
		if err != nil {
			return nil, err
		}
		privateKeyBytes := crypto.FromECDSA(privateKey)

		btcPrivKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		wif, err := btcutil.NewWIF(btcPrivKey, &netParams, true)
		if err != nil {
			return nil, err
		}
		key, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.SerializePubKey()), &netParams)
		if err != nil {
			return nil, err
		}
		return txscript.PayToAddrScript(key)
	})

	source := SecProvider{&netParams,
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

func (coin Btc) GetNetParams(testNet bool) chaincfg.Params {
	var params chaincfg.Params

	if testNet {
		params = chaincfg.TestNet3Params
	} else {
		params = chaincfg.MainNetParams
	}
	return params
}

func (coin Btc) GetAddressType(str string, testNet bool) (int8, error) {
	params := coin.GetNetParams(testNet)
	decAddr, err := btcutil.DecodeAddress(str, &params)
	if err != nil {
		return -1, err
	}

	switch decAddr.(type) {
	case *btcutil.AddressPubKeyHash:
		return 0, nil
	case *btcutil.AddressScriptHash:
		return 1, nil
	case *btcutil.AddressPubKey:
		return 2, nil
	case *btcutil.AddressWitnessPubKeyHash:
		return 3, nil
	case *btcutil.AddressWitnessScriptHash:
		return 4, nil
	case *btcutil.AddressTaproot:
		return 5, nil
	}
	return 6, nil
}

func (coin Btc) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := BtcTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

type SecProvider struct {
	Params *chaincfg.Params
	txscript.KeyDB
	txscript.ScriptDB
}

func (receiver SecProvider) ChainParams() *chaincfg.Params {
	return receiver.Params
}

func (coin Btc) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := coin.GetNetParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, params)

}

func (coin Btc) EstimateTxSizes(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet chaincfg.Params) int {
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

func (coin Btc) SumOutputSerializeSizesOfChainParams(outputAddrs []string, params chaincfg.Params) int {
	var sizeSum = 0
	for _, addr := range outputAddrs {
		decodeAddress, err := btcutil.DecodeAddress(addr, &params)
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
