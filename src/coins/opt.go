package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	OPT_TEST    = 420
	OPT_MAIN    = 10
	CurrencyOpt = "OPT"
)

var coinOpt Opt

func init() {
	coinOpt = Opt{}
	RegisterEthLikeCoin(coinOpt)
}

type Opt struct {
	Eth
}

func (coin Opt) GetCurrency() string {

	return CurrencyOpt
}
func (coin Opt) ChainName() string {
	return CurrencyOpt
}

func (coin Opt) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Opt) GetBasePath(testNet bool) string {
	return "m/44'/614'/%d'/%d/%d"
}

func (coin Opt) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(OPT_TEST)
	} else {
		chainId = *big.NewInt(OPT_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
