package coins

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/portto/solana-go-sdk/types"
	"github.com/shopspring/decimal"
	"time"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	types2 "wallet-sdk/src/types"
)

const CurrencySol = "SOL"

type SolanaTxParams struct {
	types2.BaseTxParams
	FromAddress     string          `json:"fromAddress"`
	ToAddress       string          `json:"toAddress"`
	Amount          decimal.Decimal `json:"amount"`
	Fee             decimal.Decimal `json:"fee"`
	Memo            string          `json:"memo"`
	RecentBlockHash string          `json:"solRecentBlockHash"`
}

var coinSol Sol

func init() {
	coinSol = Sol{}

	RegisterCoin(coinSol)

}

type Sol struct {
}

func (coin Sol) GetCurrency() string {
	return CurrencySol
}
func (coin Sol) ChainName() string {
	return CurrencySol
}

func (coin Sol) GetPath(index int64, testNet bool) string {

	return fmt.Sprintf("m/44'/501'/%d'/0'", index)
}

func (coin Sol) GetDeriver() deriver.Deriver {
	return &deriver.Ed25519Deriver{}
}

func (coin Sol) GetBasePath(testNet bool) string {
	return "m/44'/501'/%d'/%d'"
}

func (coin Sol) GetDecimal() int {
	return 9
}

func (coin Sol) PrivateKeyFromString(key string) (types2.PrivateKey, error) {
	return base58.Decode(key), nil
}

func (coin Sol) PrivateKeyToString(key types2.PrivateKey) (string, error) {

	return base58.Encode(key), nil
}

func (coin Sol) GenerateAddress(privateKey types2.PrivateKey, testNet bool) (*types2.CoinAddress, error) {
	publicKey := make([]byte, 32)
	copy(publicKey, privateKey[32:])
	var adddress = types2.CoinAddress{}
	adddress.AddressStr = base58.Encode(publicKey)
	return &adddress, nil
}

func (coin Sol) GetEmptyTransactionParams() types2.TxParams {
	return SolanaTxParams{}
}

func (coin Sol) CreateTransaction(params types2.TxParams, testNet bool) (*types2.BaseTransaction, error) {
	txParams := params.(SolanaTxParams)
	fromPubKey, err := solana.PublicKeyFromBase58(txParams.FromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}

	toPubKey, err := solana.PublicKeyFromBase58(txParams.ToAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	f := txParams.Amount.Mul(decimal.NewFromInt(int64(solana.LAMPORTS_PER_SOL))).IntPart()
	//mul := decimal.NewFromFloat(amount).Mul(decimal.New(int64(solana.LAMPORTS_PER_SOL), 10))

	build := system.NewTransferInstruction(
		uint64(f),
		fromPubKey,
		toPubKey,
	).Build()
	meta := solana.NewAccountMeta(fromPubKey, true, true)

	instruction := solana.NewInstruction(solana.MemoProgramID, solana.AccountMetaSlice{meta}, []byte(fmt.Sprint(time.Now().UnixMilli())))

	builder := solana.NewTransactionBuilder()
	builder.AddInstruction(build)
	builder.AddInstruction(instruction)
	if txParams.Memo != "" {
		builder.AddInstruction(solana.NewInstruction(solana.MemoProgramID, solana.AccountMetaSlice{meta}, []byte(txParams.Memo)))
	}
	builder.SetFeePayer(fromPubKey)

	recentBlockHash, err := solana.HashFromBase58(txParams.RecentBlockHash)
	builder.SetRecentBlockHash(recentBlockHash)

	transaction, err := builder.Build()

	return &types2.BaseTransaction{
		CoinTransaction: transaction,
	}, nil
}

func (coin Sol) SignTx(baseTransaction *types2.BaseTransaction, testNet bool, privateKey types2.PrivateKey) (*string, error) {
	transaction := baseTransaction.CoinTransaction.(*solana.Transaction)

	keyByte := []byte(privateKey)
	solanaKey := solana.PrivateKey{}
	solanaKey = keyByte

	privateKeyGetter := func(key solana.PublicKey) *solana.PrivateKey {
		return &solanaKey
	}
	_, err := transaction.Sign(privateKeyGetter)
	if err != nil {
		return nil, err
	}
	rawTx, err := transaction.ToBase64()
	if err != nil {
		return nil, err
	}

	return &rawTx, err
}

func (coin Sol) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	data, err := base64.StdEncoding.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}
	transaction, err := solana.TransactionFromDecoder(bin.NewBinDecoder(data))
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (coin Sol) GetTransactionParamsFromJson(paramsJson string) types2.TxParams {
	params := SolanaTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Sol) ResetBlockHash(rawTx string, blockhash, privateKey string) (*string, error) {
	data, err := base64.StdEncoding.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}

	deserialize, err := types.TransactionDeserialize(data)
	if err != nil {
		return nil, err
	}
	deserialize.Message.RecentBlockHash = blockhash
	fromBase58, err := types.AccountFromBase58(privateKey)
	if err != nil {
		return nil, err
	}
	param := types.NewTransactionParam{
		Message: deserialize.Message,
		Signers: []types.Account{fromBase58},
	}
	transaction, err := types.NewTransaction(param)
	if err != nil {
		return nil, err
	}

	serialize, err := transaction.Serialize()
	if err != nil {
		return nil, err
	}
	toString := base64.StdEncoding.EncodeToString(serialize)
	return &toString, err
}
