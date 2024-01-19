package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	HT_TEST    = 256
	HT_MAIN    = 128
	CurrencyHt = "HT"
)

var coinHt Ht

func init() {
	coinHt = Ht{}
	RegisterEthLikeCoin(coinHt)

}

type Ht struct {
	Eth
}

func (coin Ht) GetCurrency() string {

	return CurrencyHt
}
func (coin Ht) ChainName() string {
	return CurrencyHt
}

func (coin Ht) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Ht) GetBasePath(testNet bool) string {
	return "m/44'/1010'/%d'/%d/%d"
}

func (coin Ht) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(HT_TEST)
	} else {
		chainId = *big.NewInt(HT_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
