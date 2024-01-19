package coins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fxamacker/cbor/v2"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"wallet-sdk/src/coins/ada"
	"wallet-sdk/src/crypto/cip1852"
	"wallet-sdk/src/deriver"
	errors2 "wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencyAda = "ADA"

type AdaTxParams struct {
	types.BaseTxParams
	SlotNo    int64      `json:"adaSlotNo"`
	Unspends  []Unspent  `json:"unspends"`
	Receivers []Receiver `json:"receivers"`
	//Ada fee= txbyte length*1000*minFeeA+minFeeB
	MinFeeA decimal.Decimal `json:"minFeeA"`
	MinFeeB decimal.Decimal `json:"minFeeB"`
}

func (params AdaTxParams) GetParams() interface{} {
	return params
}

type Unspent struct {
	Address   string          `json:"address"`
	TxHash    string          `json:"txHash"`
	TxOutputN uint32          `json:"txOutputN"`
	TxValue   decimal.Decimal `json:"txValue"`
}

type Receiver struct {
	Address string          `json:"address"`
	Value   decimal.Decimal `json:"value"`
}

var coinAda Ada

func init() {
	coinAda = Ada{}
	RegisterUtxoCoin(coinAda)
}

type Ada struct {
}

func (coin Ada) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := AdaTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Ada) GetCurrency() string {
	return CurrencyAda
}
func (coin Ada) ChainName() string {
	return CurrencyAda
}

func (coin Ada) GetDecimal() int {
	return 6
}

func (coin Ada) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Ada) GetBasePath(testNet bool) string {
	return "m/1852'/1815'/%d'/%d/%d"
}

func (coin Ada) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	privateKey, err := cip1852.PrivateKeyFromHex(key)
	if err != nil {

		return nil, err
	}
	return privateKey.KeyData, nil
}

func (coin Ada) PrivateKeyToString(key types.PrivateKey) (string, error) {
	privateKey := cip1852.PrivateKeyFromBytes(key)
	return privateKey.ToHex(), nil
}

func (coin Ada) GenerateAddressByKeyStr(key string, testnet bool) (*types.CoinAddress, error) {
	network := coin.GetNet(testnet)
	entropy, err := cip1852.PrivateKeyFromHex(key)
	if err != nil {

		return nil, err
	}
	hash, err := entropy.GetPublicKey().GetKeyHash()
	if err != nil {

		return nil, err
	}
	address, err := ada.GetEnterpriseAddress(hash, network)
	if err != nil {
		return nil, err
	}
	var coinAddress = types.CoinAddress{
		AddressStr: address,
		Currency:   coin.GetCurrency(),
	}
	return &coinAddress, nil
}

func (coin Ada) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	network := coin.GetNet(testNet)
	hdPrivatekey := cip1852.PrivateKeyFromBytes(privateKey)
	keyHash, err := hdPrivatekey.GetPublicKey().GetKeyHash()
	if err != nil {
		return nil, err
	}
	address, err := ada.GetEnterpriseAddress(keyHash, network)

	if err != nil {
		return nil, err
	}
	var coinAddress = types.CoinAddress{
		AddressStr: address,
		Currency:   coin.GetCurrency(),
	}
	return &coinAddress, nil
}

func (coin Ada) GenerateReceivingAddressByKeyByte(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	network := coin.GetNet(testNet)
	keyLen := len(privateKey) / 2
	bussinessKeyByte := make([]byte, keyLen)
	copy(bussinessKeyByte, privateKey[0:keyLen])
	businessKey := cip1852.PrivateKeyFromBytes(bussinessKeyByte)

	stakeKey := cip1852.PrivateKeyFromBytes(privateKey[keyLen:])
	businessKeyHash, err := businessKey.GetPublicKey().GetKeyHash()
	if err != nil {
		return nil, err
	}

	stakeKeyHash, err := stakeKey.GetPublicKey().GetKeyHash()
	if err != nil {
		return nil, err
	}
	address, err := ada.GetBaseAddress(businessKeyHash, stakeKeyHash, network)
	if err != nil {
		return nil, err
	}
	var coinAddress = types.CoinAddress{
		AddressStr: address,
		PrivateKey: businessKey.ToHex(),
		Currency:   coin.GetCurrency(),
	}
	return &coinAddress, nil
}

func (coin Ada) GetDeriver() deriver.Deriver {
	return &deriver.Cip1852Deriver{}
}

func (coin Ada) GetEmptyTransactionParams() types.TxParams {
	return AdaTxParams{}
}

func (coin Ada) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := params.(AdaTxParams)
	var adaUnits = 1000000.0
	builder := ada.NewTxBuilder(ada.ProtocolParams{
		MinimumUtxoValue: 1000000,
		MinFeeA:          uint64(extraParams.MinFeeA.Mul(decimal.NewFromFloat(adaUnits)).IntPart()),
		MinFeeB:          uint64(extraParams.MinFeeB.Mul(decimal.NewFromFloat(adaUnits)).IntPart()),
	})

	var receivers = extraParams.Receivers
	slotNo := extraParams.SlotNo

	for _, unspend := range extraParams.Unspends {
		mul := unspend.TxValue.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt32(6)))
		builder.AddInput(nil, ada.TransactionID(unspend.TxHash), uint64(unspend.TxOutputN), uint64(mul.IntPart()))
	}
	for _, receiver := range receivers {

		_, err := ada.GetAddressBytes(receiver.Address)
		if err != nil {

			return nil, errors2.ErrorInvalidAddress
		}
		value := receiver.Value.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt32(6)))

		builder.AddOutput(receiver.Address, uint64(value.IntPart()))
	}
	err := builder.AddFee(extraParams.Unspends[0].Address)
	if err != nil {
		return nil, err
	}
	builder.SetTtl(uint64(slotNo))
	tx := builder.Build()
	baseTransaction := types.BaseTransaction{
		CoinTransaction: tx,
	}

	return &baseTransaction, nil
}

func (coin Ada) DecodeTransaction(signedTx string, testnet bool) (interface{}, error) {
	if !strings.HasPrefix(signedTx, "0x") {
		signedTx = "0x" + signedTx
	}
	decode, err := hexutil.Decode(signedTx)
	if err != nil {
		return "", err
	}
	decoder := cbor.NewDecoder(bytes.NewReader(decode))
	transaction := ada.Transaction{}
	err = decoder.Decode(&transaction)
	if err != nil {
		return nil, err
	}
	readableTransaction := ada.HumanReadableTransaction{}
	readableTransaction.Id = string(transaction.ID())
	fmt.Println(transaction.ID())
	readableTransaction.Metadata = transaction.Metadata
	readableTransaction.WitnessSet = transaction.WitnessSet
	txBody := ada.TxBody{}

	for i, output := range transaction.Body.Outputs {
		address, err := ada.ByteToAddress(output.Address)
		if err != nil {
			return "", err
		}
		div := decimal.NewFromInt(int64(output.Amount)).Div(decimal.NewFromFloat(1000000))
		unspent := ada.Output{
			TxHash:  string(transaction.ID()),
			Amount:  div,
			Address: address,
			Index:   uint64(i),
		}
		txBody.Outputs = append(txBody.Outputs, unspent)
	}

	for _, input := range transaction.Body.Inputs {
		inputId := hexutil.Encode(input.ID)
		input := ada.Input{
			ID:    inputId,
			Index: input.Index,
		}
		txBody.Inputs = append(txBody.Inputs, input)
	}
	txBody.Fee = decimal.NewFromInt(int64(transaction.Body.Fee)).Div(decimal.NewFromFloat(1000000))
	txBody.Ttl = transaction.Body.Ttl
	txBody.Certificates = transaction.Body.Certificates
	txBody.MetadataHash = transaction.Body.MetadataHash
	txBody.Update = transaction.Body.Update
	txBody.Withdrawals = transaction.Body.Withdrawals

	readableTransaction.Tx = txBody
	return readableTransaction, nil
}

func (coin Ada) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {

	return coin.SignTxByKeys(baseTransaction, []types.PrivateKey{privateKey})
}

func (coin Ada) SignTxByKeys(baseTransaction *types.BaseTransaction, privateKeys []types.PrivateKey) (*string, error) {
	keyBag := ada.NewKeyBag()
	for i, _ := range privateKeys {
		privateKey := privateKeys[i]
		hdPrivatekey := cip1852.PrivateKeyFromBytes(privateKey)
		keyBag.AddKey(hdPrivatekey)
	}
	t := baseTransaction.CoinTransaction.(ada.Transaction)
	keyBag.Sign(&t)
	hex := t.CborHex()
	return &hex, nil
}

func (coin Ada) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error) {
	var privateKey []types.PrivateKey
	for s, _ := range keys {
		privateKey = append(privateKey, keys[s])
	}
	return coin.SignTxByKeys(baseTransaction, privateKey)
}

func (coin Ada) CheckPrivateKey(privateKey string, address string, testnet bool) (bool, error) {
	addressBytes, err := ada.GetAddressBytes(address)
	if err != nil {
		return false, err
	}
	hex, err := cip1852.PrivateKeyFromHex(privateKey)
	if err != nil {
		return false, err
	}
	hash, err := hex.GetPublicKey().GetKeyHash()
	if err != nil {
		return false, err
	}
	if len(addressBytes) < 29 {
		return false, errors.New("address not supported")
	}
	return byteEq(hash, addressBytes[1:29]), nil
}

func (coin Ada) CalculateFee(extraParams AdaTxParams) (float64, error) {
	var adaUnits = 1000000.0
	builder := ada.NewTxBuilder(ada.ProtocolParams{
		MinimumUtxoValue: 1000000,
		MinFeeA:          uint64(extraParams.MinFeeA.Mul(decimal.NewFromFloat(adaUnits)).IntPart()),
		MinFeeB:          uint64(extraParams.MinFeeB.Mul(decimal.NewFromFloat(adaUnits)).IntPart()),
	})
	var unspends = extraParams.Unspends
	var receivers = extraParams.Receivers

	for _, unspend := range unspends {

		parseUint, err := strconv.ParseFloat(unspend.TxValue.String(), 64)
		if err != nil {
			return 0, err
		}
		parseUint = parseUint * 1000000
		builder.AddInput(nil, ada.TransactionID(unspend.TxHash), uint64(unspend.TxOutputN), uint64(parseUint))
	}

	for _, receiver := range receivers {

		parseUint, err := strconv.ParseFloat(receiver.Value.String(), 64)
		if err != nil {
			return 0, err
		}
		parseUint = parseUint * 1000000
		builder.AddOutput(receiver.Address, uint64(parseUint))
	}

	builder.AddFee(receivers[0].Address)
	tx := builder.Build()
	fee := tx.Body.Fee
	f := float64(fee)
	f /= 1000000

	return f, nil

}

func (coin Ada) GetNet(testnet bool) ada.Network {
	var network ada.Network
	if testnet {
		network = ada.TestNet
	} else {
		network = ada.MainNet
	}
	return network
}

func byteEq(a, b []byte) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
