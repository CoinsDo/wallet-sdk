package coins

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"wallet-sdk/src/types"
)

const CurrencyDash = "DASH"

var coinDash Dash

func init() {
	coinDash = Dash{}
	RegisterUtxoCoin(coinDash)
}

type Dash struct {
	Btc
}

func (coin Dash) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	netParams := coin.GetDashParams(testNet)
	return coin.GenAddress(keyByte, netParams)
}

func (coin Dash) GenerateAddressByKeyStr(key string, testnet bool) (*types.CoinAddress, error) {
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

func (coin Dash) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Dash) GetCurrency() string {
	return CurrencyDash
}

func (coin Dash) GetBasePath(testNet bool) string {
	return "m/44'/5'/%d'/%d/%d"
}

func (coin Dash) ChainName() string {
	return CurrencyDash
}

func (coin Dash) CreateTransaction(txParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	return coin.Transaction(txParams, coin.GetDashParams(testNet))
}

func (coin Dash) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	netParams := coin.GetDashParams(testNet)
	return coin.Sign(baseTransaction, privateKey, netParams)
}

func (coin Dash) SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, privateKeys map[string]types.PrivateKey) (*string, error) {
	netParams := coin.GetDashParams(testNet)
	return coin.SignMultipleSendAddress(baseTransaction, privateKeys, netParams)
}

func (coin Dash) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {
	params := coin.GetDashParams(testnet)
	return coin.DecodeTx(rawTx, params)

}

func (coin Dash) EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int {

	params := coin.GetDashParams(testNet)
	return coin.EstimateTxSizes(inputCount, outputAddrs, hasExtraChangeAddr, params)

}

func (coin Dash) GetDashParams(testNet bool) chaincfg.Params {
	var params chaincfg.Params
	if testNet {
		params = DashTestNet3Params()
	} else {
		params = DashMainNetParams()
	}
	return params
}
