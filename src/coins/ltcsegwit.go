package coins

import (
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ltcsuite/ltcd/btcec/v2"
	"github.com/ltcsuite/ltcd/ltcutil"
	"golang.org/x/crypto/ripemd160"
	"wallet-sdk/src/types"
)

const CurrencyLtcSegwit = "LTC_SEGWIT"

var coinLtcSegwit LtcSegwit

func init() {
	coinLtcSegwit = LtcSegwit{}
	RegisterUtxoCoin(coinLtcSegwit)

}

type LtcSegwit struct {
	Ltc
}

func (coin LtcSegwit) GetCurrency() string {
	return CurrencyLtcSegwit
}

func (coin LtcSegwit) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin LtcSegwit) GetBasePath(testNet bool) string {
	return "m/84'/2'/%d'/%d/%d"
}

func (coin LtcSegwit) ChainName() string {
	return CurrencyLtc
}

func (coin LtcSegwit) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := ltcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}

func (coin LtcSegwit) GenerateAddress(keyByte types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	privateKey, err := crypto.ToECDSA(keyByte)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	_, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	params := getLtcNetParams(testNet)

	shaHash := sha256.Sum256(pubKey.SerializeCompressed())
	ripeHasher := ripemd160.New()
	_, _ = ripeHasher.Write(shaHash[:])
	publicKeyHash := ripeHasher.Sum(nil)

	hash, err := ltcutil.NewAddressWitnessPubKeyHash(publicKeyHash, &params)
	if err != nil {
		return nil, err
	}
	var adddress = types.CoinAddress{}
	adddress.AddressStr = hash.EncodeAddress()
	return &adddress, nil
}
