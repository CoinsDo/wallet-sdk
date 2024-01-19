package coins

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"wallet-sdk/src/types"
)

const CurrencyDoge = "DOGE"

var coinDoge Doge

func init() {
	coinDoge = Doge{}
	RegisterUtxoCoin(coinDoge)
}

type Doge struct {
	Btc
}

func (coin Doge) GetCurrency() string {
	return CurrencyDoge
}

func (coin Doge) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.GenAddress(privateKey, netParams)
}

func (coin Doge) GenerateAddressByKeyStr(key string, testnet bool) (*types.CoinAddress, error) {
	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}

	privKey := wif.PrivKey
	toECDSA, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Serialize()))
	if err != nil {
		return nil, err
	}
	return coin.GenerateAddress(toECDSA.D.Bytes(), testnet)
}

func (coin Doge) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Doge) GetBasePath(testNet bool) string {
	return "m/44'/3'/%d'/%d/%d"
}

func (coin Doge) ChainName() string {
	return CurrencyDoge
}

func (coin Doge) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.Transaction(txParams, netParams)
}

func (coin Doge) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.Sign(baseTransaction, privateKey, netParams)
}

func (coin Doge) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, privateKeys map[string]types.PrivateKey) (*string, error) {
	netParams := coin.GetNetParams(testNet)
	return coin.SignMultipleSendAddress(baseTransaction, privateKeys, netParams)
}

func (coin Doge) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	params := coin.GetNetParams(testnet)
	return coin.DecodeTx(rawTx, params)

}

func (coin Doge) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := coin.GetNetParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, params)

}

func (coin Doge) GetNetParams(testNet bool) chaincfg.Params {
	var params chaincfg.Params
	if testNet {
		params = DogeTestNet3Params()
	} else {
		params = DogeMainNetParams()
	}
	return params
}
