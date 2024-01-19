package coins

import (
	"encoding/json"
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/decred/base58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/vedhavyas/go-subkey/sr25519"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	types2 "wallet-sdk/src/types"
)

const CurrencyDot = "DOT"

type DotTxParams struct {
	types2.BaseTxParams
	ToAddress   string          `json:"toAddress"`
	Amount      decimal.Decimal `json:"amount"`
	Meta        string          `json:"dotMeta"`
	AccountInfo DotAccountInfo  `json:"dotAccountInfo"`
}

type DotAccountInfo struct {
	Nonce              int32  `json:"nonce"`
	GenesisHash        string `json:"genesisHash"`
	TransactionVersion int32  `json:"transactionVersion"`
	ImplVersion        int32  `json:"implVersion"`
	SpecVersion        int32  `json:"specVersion"`
}

var coinDot Dot

func init() {
	coinDot = Dot{}
	RegisterCoin(coinDot)
}

type Dot struct {
}

func (coin Dot) GetCurrency() string {
	return CurrencyDot
}

func (coin Dot) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, index)
}

func (coin Dot) GetBasePath(testNet bool) string {
	return "//polkadot//%d//%d"
}

func (coin Dot) ChainName() string {
	return CurrencyDot
}

func (coin Dot) GetDecimal() int {
	return 12
}

func (coin Dot) PrivateKeyFromString(key string) (types2.PrivateKey, error) {
	decode, err := hexutil.Decode(key)
	if err != nil {
		return nil, err
	}
	return decode, nil
}

func (coin Dot) PrivateKeyToString(key types2.PrivateKey) (string, error) {

	return hexutil.Encode(key), nil
}

func (coin Dot) GenerateAddress(privateKey types2.PrivateKey, testNet bool) (*types2.CoinAddress, error) {
	var err error
	scheme := sr25519.Scheme{}
	kyr, err := scheme.FromSeed(privateKey)
	var ss58Address string
	if testNet {
		ss58Address, err = kyr.SS58Address(42)
	} else {
		ss58Address, err = kyr.SS58Address(0)
	}

	if err != nil {
		return nil, err
	}

	var adddress = types2.CoinAddress{}
	adddress.AddressStr = ss58Address

	return &adddress, nil
}

func (coin Dot) GetDeriver() deriver.Deriver {
	return &deriver.DotDeriver{}
}

func (coin Dot) GetEmptyTransactionParams() types2.TxParams {
	return DotTxParams{}
}

func (coin Dot) CreateTransaction(params types2.TxParams, testNet bool) (*types2.BaseTransaction, error) {
	txParams := params.(DotTxParams)
	var accurrency = 10
	if testNet {
		accurrency = 12
	}
	finalAmount := (decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(accurrency))).Mul(txParams.Amount)).IntPart()
	dstAccountID, err := DotAddressToPublicKey(txParams.ToAddress)
	if err != nil {
		return nil, err
	}

	var meta = types.Metadata{}

	err = types.DecodeFromHex(txParams.Meta, &meta)
	if err != nil {
		return nil, err
	}
	dst := types.NewMultiAddressFromAccountID(dstAccountID)

	c, err := types.NewCall(&meta, "Balances.transfer_allow_death", dst, types.NewUCompactFromUInt(uint64(finalAmount)))
	if err != nil {
		return nil, err
	}

	ext := types.NewExtrinsic(c)
	nonce := uint32(txParams.AccountInfo.Nonce)
	genesisHash, err := types.NewHashFromHexString(txParams.AccountInfo.GenesisHash)
	if err != nil {
		return nil, err
	}
	specVersion := txParams.AccountInfo.SpecVersion
	transactionVersio := txParams.AccountInfo.TransactionVersion
	opt := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        types.U32(specVersion),
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: types.U32(transactionVersio),
	}

	return &types2.BaseTransaction{CoinTransaction: &DotTransaction{&ext, &opt}}, nil
}

func (coin Dot) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	return nil, errors.ErrorDecodeNotSupported
}

func (coin Dot) SignTx(baseTransaction *types2.BaseTransaction, testNet bool, privateKey types2.PrivateKey) (*string, error) {

	var kr signature.KeyringPair
	var err error
	encode := hexutil.Encode(privateKey)
	if testNet {
		kr, err = signature.KeyringPairFromSecret(encode, 42)
	} else {
		kr, err = signature.KeyringPairFromSecret(encode, 0)
	}

	if err != nil {
		return nil, err
	}

	tr := baseTransaction.CoinTransaction.(*DotTransaction)

	err = tr.ext.Sign(kr, *tr.opt)

	if err != nil {
		return nil, err
	}
	enc, err := types.EncodeToHex(*tr.ext)

	return &enc, err
}

func (coin Dot) GetTransactionParamsFromJson(paramsJson string) types2.TxParams {
	params := DotTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func DotAddressToPublicKey(address string) ([]byte, error) {
	d58 := base58.Decode(address)
	if len(d58) < 35 {
		return nil, errors.ErrorInvalidAddress
	}
	pk := d58[1 : len(d58)-2]
	return pk, nil
}

type DotTransaction struct {
	ext *types.Extrinsic
	opt *types.SignatureOptions
}
