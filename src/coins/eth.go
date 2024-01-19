package coins

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"strings"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/types"
)

const (
	TESTNET     = 5 //Goerli
	MAINNET     = 1
	CurrencyEth = "ETH"
)

type EthTxParams struct {
	types.BaseTxParams
	ToAddress string          `json:"toAddress"`
	Amount    decimal.Decimal `json:"amount"`
	Nonce     int64           `json:"ethereumNonce"`
	GasPrice  types.BigInt    `json:"ethereumGasPrice"`
	GasLimit  types.BigInt    `json:"ethereumGasLimit"`
	Data      string          `json:"ethereumData"`
}

var coinEth Eth

func init() {
	coinEth = Eth{}
	RegisterEthLikeCoin(coinEth)

}

type Eth struct {
}

func (coin Eth) GetCurrency() string {

	return CurrencyEth
}
func (coin Eth) ChainName() string {
	return CurrencyEth
}

func (coin Eth) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Eth) GetBasePath(testNet bool) string {
	return "m/44'/60'/%d'/%d/%d"
}

func (coin Eth) GetDecimal() int {
	return 18
}

func (coin Eth) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Eth) PrivateKeyToString(key types.PrivateKey) (string, error) {
	return hex.EncodeToString(key), nil
	//return
}

func (coin Eth) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	privateKey, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (coin Eth) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	return createAddress(privateKey, testNet)

}

func (coin Eth) GetEmptyTransactionParams() types.TxParams {
	return EthTxParams{}

}

func (coin Eth) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return createTransaction(params)

}

func (coin Eth) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	return DecodeTx(rawTx)
}

func (coin Eth) CreateDappTransaction(params types.TxParams) (*types.BaseTransaction, error) {
	return createTransaction(params)
}

func (coin Eth) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(TESTNET)
	} else {
		chainId = *big.NewInt(MAINNET)
	}
	return signTx(&chainId, baseTransaction, privateKey)

}

func (coin Eth) SignMessage(message string, privateKey types.PrivateKey) (string, error) {

	var signFormat = "\x19Ethereum Signed Message:\n%d%s"
	newUnsignData := fmt.Sprintf(signFormat, len([]byte(message)), message)
	unsignDataHash := crypto.Keccak256([]byte(newUnsignData))
	key, err := crypto.ToECDSA(privateKey)
	if err != nil {
		log.Println("sign ToECDSA err:", err.Error())
		return "", err
	}
	signatureByte, err := crypto.Sign(unsignDataHash, key)
	if err != nil {
		log.Println("sign Sign err:", err.Error())
		return "", err
	}
	signatureByte[64] += 27
	return hexutil.Encode(signatureByte), nil
}

func (coin Eth) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := EthTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}
