package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	XDAI_TEST    = 100
	XDAI_MAIN    = 100
	CurrencyXdai = "XDAI"
)

var coinXdai Xdai

func init() {
	coinXdai = Xdai{}
	RegisterEthLikeCoin(coinXdai)
}

type Xdai struct {
	Eth
}

func (coin Xdai) GetCurrency() string {

	return CurrencyXdai
}
func (coin Xdai) ChainName() string {
	return CurrencyXdai
}

func (coin Xdai) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Xdai) GetBasePath(testNet bool) string {
	return "m/44'/700'/%d'/%d/%d"
}

func (coin Xdai) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(XDAI_TEST)
	} else {
		chainId = *big.NewInt(XDAI_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
