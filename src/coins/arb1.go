package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	ARB1_TEST = 421613
	ARB1_MAIN = 42161
)

const CurrencyArb1 = "ARB1"

var coinArb1 Arb1

func init() {
	coinArb1 = Arb1{}
	RegisterEthLikeCoin(coinArb1)
}

type Arb1 struct {
	Eth
}

func (coin Arb1) GetCurrency() string {
	return CurrencyArb1
}

func (coin Arb1) ChainName() string {
	return CurrencyArb1
}

func (coin Arb1) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Arb1) GetBasePath(testNet bool) string {
	return "m/44'/9001'/%d'/%d/%d"
}

func (coin Arb1) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(ARB1_TEST)
	} else {
		chainId = *big.NewInt(ARB1_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
