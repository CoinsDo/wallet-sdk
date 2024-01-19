package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	OKT_TEST    = 65
	OKT_MAIN    = 66
	CurrencyOkt = "OKT"
)

var coinOkt Okt

func init() {
	coinOkt = Okt{}
	RegisterEthLikeCoin(coinOkt)
}

type Okt struct {
	Eth
}

func (coin Okt) GetCurrency() string {

	return CurrencyOkt
}
func (coin Okt) ChainName() string {
	return CurrencyOkt
}

func (coin Okt) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Okt) GetBasePath(testNet bool) string {
	return "m/44'/996'/%d'/%d/%d"
}

func (coin Okt) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(OKT_TEST)
	} else {
		chainId = *big.NewInt(OKT_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
