package coins

import (
	"strings"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

var supportedCoins = make(map[string]Coin, 0)
var useUtxoCoins = make(map[string]Coin, 0)
var ethLikeCoins = make(map[string]Coin, 0)
var nftCoins = make(map[string]Coin, 0)

type Coin interface {
	// GetCurrency Get the currency
	GetCurrency() string
	// GetPath Get the bip44 path of the index node part
	GetPath(index int64, testNet bool) string
	// GetBasePath Get the bip44 test path index node part, mostly the same as PATH, mainly used for test chains
	GetBasePath(testNet bool) string
	// ChainName Get the chain name - the same as the main currency
	ChainName() string
	// GetDecimal Get the precision
	GetDecimal() int

	// PrivateKeyFromString string key to []byte
	PrivateKeyFromString(key string) (types.PrivateKey, error)

	// PrivateKeyToString key byte to string
	PrivateKeyToString(key types.PrivateKey) (string, error)

	// GenerateAddress Generate an address with a byte array private key
	GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error)

	// GetDeriver Get the deriver
	GetDeriver() deriver.Deriver
	// GetEmptyTransactionParams Get transaction building parameters
	GetEmptyTransactionParams() types.TxParams

	// CreateTransaction Create a raw transaction (generally unsigned)
	CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error)

	// SignTx Sign the transaction, returning the signed rawHex
	SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error)

	// GetTransactionParamsFromJson Parse JSON into TxParams
	GetTransactionParamsFromJson(paramsJson string) types.TxParams
}

type TransactionDecoder interface {
	DecodeTransaction(rawTx string, testnet bool) (interface{}, error)
}

type BitcoinMultiSenderSigner interface {
	SignMultipleSendAddressTx(baseTransaction *types.BaseTransaction, testNet bool, keys map[string]types.PrivateKey) (*string, error)
}

type BitcoinSizeEstimator interface {
	EstimateSize(inputCount int, outputAddrs []string, hasExtraChangeAddr bool, testNet bool) int
}

func GetSupportedCurrencies() []Coin {
	var coins []Coin
	for _, coin := range supportedCoins {
		coins = append(coins, coin)
	}
	return coins
}

// GetCoin Add a Coin object
func GetCoin(currency string) (Coin, error) {
	coin := supportedCoins[strings.ToUpper(currency)]
	if coin != nil {
		return coin, nil
	}
	return nil, errors.ErrorCurrencyNotSupported
}

// RegisterCoin Add a Coin object
func RegisterCoin(coin Coin) {
	supportedCoins[coin.GetCurrency()] = coin
}

// RegisterUtxoCoin Add a Utxo object
func RegisterUtxoCoin(coin Coin) {
	useUtxoCoins[coin.GetCurrency()] = coin
	RegisterCoin(coin)
}

// RegisterEthLikeCoin Add an Evm object
func RegisterEthLikeCoin(coin Coin) {
	ethLikeCoins[coin.GetCurrency()] = coin
	RegisterCoin(coin)
}

// RegisterNftToken Add an NftCoin object
func RegisterNftToken(coin Coin) {
	nftCoins[coin.GetCurrency()] = coin
	RegisterCoin(coin)
}

// GetTokenList Get the collection of all main coins and tokens under the same main chain
func GetTokenList(currency string) ([]Coin, error) {
	coin, err := GetCoin(currency)
	if err != nil {
		return nil, err
	}
	supportedCurrencies := GetSupportedCurrencies()

	var tokenList []Coin
	for i := range supportedCurrencies {
		if coin.ChainName() == supportedCurrencies[i].ChainName() {
			tokenList = append(tokenList, supportedCurrencies[i])
		}
	}
	return tokenList, nil
}

// IsToken Check if it's a TOKEN
func IsToken(currency string) (bool, error) {
	coin, err := GetCoin(currency)
	if err != nil {
		return false, err
	}
	if coin.GetCurrency() == CurrencyLtcSegwit {
		return false, nil
	}

	return coin.ChainName() != coin.GetCurrency(), nil
}

// IsNft Check if it's an NFT
func IsNft(currency string) (bool, error) {
	_, err := GetCoin(currency)
	if err != nil {
		return false, err
	}
	return nftCoins[currency] != nil, nil
}

// IsEthLike Check if it's an ETH-like chain
func IsEthLike(currency string) (bool, error) {
	_, err := GetCoin(currency)
	if err != nil {
		return false, err
	}
	return ethLikeCoins[currency] != nil, nil
}

// IsUsingUtxo Check if it's using UTXO transactions
func IsUsingUtxo(currency string) (bool, error) {
	_, err := GetCoin(currency)
	if err != nil {
		return false, err
	}
	return useUtxoCoins[currency] != nil, nil
}
