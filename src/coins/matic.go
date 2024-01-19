package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	MATIC_TEST    = 80001
	MATIC_MAIN    = 137
	CurrencyMatic = "MATIC"
)

var coinMatic Matic

func init() {
	coinMatic = Matic{}

	RegisterEthLikeCoin(coinMatic)
}

type Matic struct {
	Eth
}

func (coin Matic) GetCurrency() string {

	return CurrencyMatic
}
func (coin Matic) ChainName() string {
	return CurrencyMatic
}

func (coin Matic) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Matic) GetBasePath(testNet bool) string {
	return "m/44'/966'/%d'/%d/%d"
}

func (coin Matic) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(MATIC_TEST)
	} else {
		chainId = *big.NewInt(MATIC_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
