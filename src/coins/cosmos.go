package coins

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	xauthTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/shopspring/decimal"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/types"
)

const CurrencyAtom = "ATOM"

type AtomTxParams struct {
	types.BaseTxParams
	ToAddress   string          `json:"toAddress"`
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	Memo        string          `json:"memo"`
	FromAddress string          `json:"fromAddress"`

	GasLimit        uint64 `json:"gasLimit"`
	ChainId         string `json:"chainId"`
	AccountNumber   uint64 `json:"accountNumber"`
	AccountSequence uint64 `json:"accountSequence"`
}

var coinAtom = Atom{}

func init() {
	RegisterCoin(coinAtom)
}

type Atom struct {
}

func (coin Atom) GetCurrency() string {
	return CurrencyAtom
}

func (coin Atom) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Atom) GetBasePath(testNet bool) string {
	return "m/44'/118'/%d'/%d/%d"
}

func (coin Atom) ChainName() string {
	return CurrencyAtom
}

func (coin Atom) GetDecimal() int {
	return 6
}

func (coin Atom) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	decodeString, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return decodeString, nil
}

func (coin Atom) PrivateKeyToString(key types.PrivateKey) (string, error) {

	return hex.EncodeToString(key), nil
}

func (coin Atom) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	priv := secp256k1.PrivKey{Key: privateKey}
	addr := sdk.AccAddress(priv.PubKey().Address()).String()
	address := types.CoinAddress{
		AddressStr: addr,
		PrivateKey: hex.EncodeToString(privateKey),
	}
	return &address, nil
}

func (coin Atom) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Atom) GetEmptyTransactionParams() types.TxParams {
	return BtcTxParams{}
}

func (coin Atom) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := params.(AtomTxParams)
	fromAccAddress, err := sdk.AccAddressFromBech32(extraParams.FromAddress)
	if err != nil {
		return nil, err
	}
	amountCoin := sdk.NewCoin("uatom", sdk.NewInt(extraParams.Amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(6))).IntPart()))

	feeCoin := sdk.NewCoin("uatom", sdk.NewInt(extraParams.Fee.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(6))).IntPart()))
	toAccAddress, err := sdk.AccAddressFromBech32(extraParams.ToAddress)
	if err != nil {
		return nil, err
	}
	msgSend := banktypes.NewMsgSend(fromAccAddress, toAccAddress, sdk.NewCoins(amountCoin))

	interfaceRegistry := cosmosTypes.NewInterfaceRegistry()
	protcalCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := xauthTx.NewTxConfig(protcalCodec, xauthTx.DefaultSignModes)
	if err != nil {
		return nil, err
	}
	txfNoKeybase := tx.Factory{}.
		WithTxConfig(txConfig).
		WithAccountNumber(extraParams.AccountNumber).
		WithSequence(extraParams.AccountSequence).
		WithGas(extraParams.GasLimit).
		WithFees(feeCoin.String()).
		WithMemo(extraParams.Memo).
		WithChainID(extraParams.ChainId)
	unsignedTx, err := txfNoKeybase.BuildUnsignedTx(msgSend)
	txWithParams := TxWithParams{
		TxParams:    extraParams,
		Transaction: unsignedTx,
	}
	transaction := types.BaseTransaction{
		CoinTransaction: txWithParams,
	}
	return &transaction, nil
}

func (coin Atom) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	txWithParams := baseTransaction.CoinTransaction.(TxWithParams)
	txBuilder := txWithParams.Transaction
	txParams := txWithParams.TxParams
	config, err := network.DefaultConfigWithAppConfig(network.MinimumAppConfig())
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	priv := secp256k1.PrivKey{Key: privateKey}

	var sigsV2 []signing.SignatureV2
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  config.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: txParams.AccountSequence,
	}

	sigsV2 = append(sigsV2, sigV2)
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	signerData := xauthsigning.SignerData{
		ChainID:       txParams.ChainId,
		AccountNumber: txParams.AccountNumber,
		Sequence:      txParams.AccountSequence,
	}
	sigs, err := tx.SignWithPrivKey(
		config.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, &priv, config.TxConfig, txParams.AccountSequence)
	if err != nil {
		return nil, err
	}

	sigsV2 = append(sigsV2, sigs)
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	txBytes, err := config.TxConfig.TxEncoder()(txBuilder.GetTx())

	toString := base64.StdEncoding.EncodeToString(txBytes)
	return &toString, nil
}

func (coin Atom) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := AtomTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

type TxWithParams struct {
	TxParams    AtomTxParams
	Transaction client.TxBuilder
}
