package coins

import (
	"fmt"
	"math/big"
	"wallet-sdk/src/types"
)

const (
	FTM_TEST    = 4002
	FTM_MAIN    = 250
	CurrencyFtm = "FTM"
)

var coinFtm Ftm

func init() {
	coinFtm = Ftm{}
	RegisterEthLikeCoin(coinFtm)
}

type Ftm struct {
	Eth
}

func (coin Ftm) GetCurrency() string {

	return CurrencyFtm
}
func (coin Ftm) ChainName() string {
	return CurrencyFtm
}

func (coin Ftm) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Ftm) GetBasePath(testNet bool) string {
	return "m/44'/1007'/%d'/%d/%d"
}

func (coin Ftm) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	var chainId big.Int
	if testNet {
		chainId = *big.NewInt(FTM_TEST)
	} else {
		chainId = *big.NewInt(FTM_MAIN)
	}
	return signTx(&chainId, baseTransaction, privateKey)
}
