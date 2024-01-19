package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	BNB_TEST    = 97
	BNB_MAIN    = 56
	CurrencyBnb = "BNB"
)

var coinBnb Bnb

func init() {
	coinBnb = Bnb{}
	RegisterEthLikeCoin(coinBnb)
}

type Bnb struct {
	Eth
}

func (coin Bnb) GetCurrency() string {

	return CurrencyBnb
}
func (coin Bnb) ChainName() string {
	return CurrencyBnb
}

func (coin Bnb) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Bnb) GetBasePath(testNet bool) string {
	return "m/44'/9006'/%d'/%d/%d"
}

func (coin Bnb) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(BNB_TEST)
	} else {
		chainId = *big.NewInt(BNB_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
