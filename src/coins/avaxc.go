package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	AVAX_TEST = 43113
	AVAX_MAIN = 43114
)

const CurrencyAvaxc = "AVAXC"

var coinAvaxc Avaxc

func init() {
	coinAvaxc = Avaxc{}

	RegisterEthLikeCoin(coinAvaxc)
}

type Avaxc struct {
	Eth
}

func (coin Avaxc) GetCurrency() string {
	return CurrencyAvaxc
}

func (coin Avaxc) ChainName() string {
	return CurrencyAvaxc
}

func (coin Avaxc) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Avaxc) GetBasePath(testNet bool) string {
	return "m/44'/9005'/%d'/%d/%d"
}

func (coin Avaxc) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(AVAX_TEST)
	} else {
		chainId = *big.NewInt(AVAX_MAIN)
	}

	return signTx(&chainId, baseTransaction, privateKey)

}
