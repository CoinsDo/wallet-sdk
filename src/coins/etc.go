package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	ETC_TEST    = 63
	ETC_MAIN    = 61
	CurrencyEtc = "ETC"
)

var coinEtc Etc

func init() {
	coinEtc = Etc{}
	RegisterEthLikeCoin(coinEtc)
}

type Etc struct {
	Eth
}

func (coin Etc) GetCurrency() string {

	return CurrencyEtc
}
func (coin Etc) ChainName() string {
	return CurrencyEtc
}

func (coin Etc) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Etc) GetBasePath(testNet bool) string {
	return "m/44'/61'/%d'/%d/%d"
}

func (coin Etc) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(ETC_TEST)
	} else {
		chainId = *big.NewInt(ETC_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
