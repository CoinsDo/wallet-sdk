package coins

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/shopspring/decimal"
	"strings"
	"wallet-sdk/src/coins/filecoin"
	"wallet-sdk/src/coins/filecoin/local"
	"wallet-sdk/src/coins/filecoin/secp256k1"
	"wallet-sdk/src/coins/filecoin/types"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	types2 "wallet-sdk/src/types"
)

const CurrencyFileCoin = "FIL"

type FileTxParams struct {
	types2.BaseTxParams
	FromAddress string          `json:"fromAddress"`
	ToAddress   string          `json:"toAddress"`
	Amount      decimal.Decimal `json:"amount"`
	Nonce       uint64          `json:"fileNonce"`
	GasLimit    int64           `json:"fileGasLimit"`
	GasFeeCap   int64           `json:"fileGasFeeCap"`
	GasPremium  int64           `json:"fileGasPremium"`
}

var coinFile FileCoin

func init() {
	coinFile = FileCoin{}
	RegisterCoin(coinFile)
}

type FileCoin struct {
}

func (coin FileCoin) GetCurrency() string {
	return CurrencyFileCoin
}
func (coin FileCoin) ChainName() string {
	return CurrencyFileCoin
}

func (coin FileCoin) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin FileCoin) GetBasePath(testNet bool) string {
	return "m/44'/461'/%d'/%d/%d"
}

func (coin FileCoin) GetDecimal() int {
	return 18
}

func (coin FileCoin) PrivateKeyFromString(key string) (types2.PrivateKey, error) {
	privateKey, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (coin FileCoin) PrivateKeyToString(key types2.PrivateKey) (string, error) {
	return hexutil.Encode(key), nil
}

func (coin FileCoin) GenerateAddress(privateKey types2.PrivateKey, testNet bool) (*types2.CoinAddress, error) {
	privateKeyBytes, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	publicKey := secp256k1.PublicKey(crypto.FromECDSA(privateKeyBytes))

	k1Address, err := address.NewSecp256k1Address(publicKey)
	if err != nil {
		return nil, err

	}
	if testNet {
		address.CurrentNetwork = address.Testnet
	} else {
		address.CurrentNetwork = address.Mainnet
	}

	var adddress = types2.CoinAddress{}
	adddress.AddressStr = k1Address.String()

	return &adddress, nil
}

func (coin FileCoin) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin FileCoin) GetEmptyTransactionParams() types2.TxParams {
	return FileTxParams{}
}

func (coin FileCoin) CreateTransaction(params types2.TxParams, testNet bool) (*types2.BaseTransaction, error) {
	extraParams := params.(FileTxParams)

	to, err := address.NewFromString(extraParams.ToAddress)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	from, err := address.NewFromString(extraParams.FromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}

	msg := &types.Message{
		Version:    0,
		To:         to,
		From:       from,
		Nonce:      extraParams.Nonce,
		Value:      filecoin.FromFil(extraParams.Amount),
		GasLimit:   extraParams.GasLimit,
		GasFeeCap:  abi.NewTokenAmount(extraParams.GasFeeCap),
		GasPremium: abi.NewTokenAmount(extraParams.GasPremium),
		Method:     0,
		Params:     nil,
	}
	return &types2.BaseTransaction{CoinTransaction: msg}, nil
}

func (coin FileCoin) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {

	message := types.Message{}
	err := json.Unmarshal([]byte(rawTx), &message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (coin FileCoin) GetTransactionParamsFromJson(paramsJson string) types2.TxParams {
	params := FileTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin FileCoin) SignTx(baseTransaction *types2.BaseTransaction, testNet bool, privateKey types2.PrivateKey) (*string, error) {
	toECDSA, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(toECDSA)
	tx := baseTransaction.CoinTransaction.(*types.Message)
	s, err := local.WalletSignMessage(types.KTSecp256k1, privateKeyBytes, tx)
	if err != nil {
		return nil, err
	}
	if err := local.WalletVerifyMessage(s); err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	s2 := string(marshal)
	return &s2, nil
}
