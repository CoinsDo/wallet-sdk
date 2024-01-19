package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	ETHW_TEST    = 10001
	ETHW_MAIN    = 10001
	CurrencyEthw = "ETHW"
)

var coinEthw Ethw

func init() {
	coinEthw = Ethw{}
	RegisterEthLikeCoin(coinEthw)

}

type Ethw struct {
	Eth
}

func (coin Ethw) GetCurrency() string {

	return CurrencyEthw
}
func (coin Ethw) ChainName() string {
	return CurrencyEthw
}

func (coin Ethw) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Ethw) GetBasePath(testNet bool) string {
	return "m/44'/63'/%d'/%d/%d"
}

func (coin Ethw) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(ETHW_TEST)
	} else {
		chainId = *big.NewInt(ETHW_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
