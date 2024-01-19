package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	guuid "github.com/google/uuid"
	"github.com/tyler-smith/go-bip39"
	"log"
	"reflect"
	"wallet-sdk/src/coins"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/types"
)

type Wallet struct {
	mnemonic string
	derivers map[string]deriver.Deriver
}

// New   Create a new wallet.
func New() (*Wallet, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}
	return NewFromEntropy(entropy)
}

// NewFromEntropy Create a wallet from existing root private key.
func NewFromEntropy(entropy []byte) (*Wallet, error) {
	fromEntropy, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	return NewFromMnemonic(fromEntropy)
}

// NewFromMnemonic Create a wallet from an existing mnemonic.
func NewFromMnemonic(mnemonic string) (*Wallet, error) {
	_, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	wallet := Wallet{
		mnemonic: mnemonic,
		derivers: map[string]deriver.Deriver{},
	}
	return &wallet, nil
}

// NewFromKeystore Create a wallet from an existing KeyStore.
func NewFromKeystore(keystoreStr string, password string) (*Wallet, error) {
	mnemonic, err := getMnemonicFromKeystore(keystoreStr, password)
	if err != nil {
		return nil, err
	}
	return NewFromMnemonic(*mnemonic)
}

// ExportKeyStore Export KeyStore.
func (wallet Wallet) ExportKeyStore(password string) (*string, error) {
	return createKeyStore(wallet.mnemonic, password)
}

// GetMnemonic Get the mnemonic.
func (wallet Wallet) GetMnemonic() string {
	return wallet.mnemonic
}

// DerivePrivateKey Derive a private key.
func (wallet Wallet) DerivePrivateKey(currency string, index int64, testNet bool) (types.PrivateKey, types.Path, error) {
	coin, err := coins.GetCoin(currency)
	if err != nil {
		return nil, "", err
	}
	path := coin.GetPath(index, testNet)
	return wallet.DerivePrivateKeyByPath(currency, types.Path(path))
}

// DerivePrivateKeyByPath Derive a private key using a specific path.
func (wallet Wallet) DerivePrivateKeyByPath(currency string, path types.Path) (types.PrivateKey, types.Path, error) {
	coin, err := coins.GetCoin(currency)
	if err != nil {
		return nil, "", err
	}
	coinDeriver, err := wallet.getDeriver(coin)
	if err != nil {
		return nil, "", err
	}
	key, err := coinDeriver.Derive(string(path))
	if err != nil {
		return nil, "", err
	}
	return key, path, nil
}

func (wallet Wallet) getDeriver(coin coins.Coin) (deriver.Deriver, error) {
	coinDeriver := coin.GetDeriver()
	deriverName := reflect.TypeOf(coinDeriver).String()
	walletCoinDeriver := wallet.derivers[deriverName]
	if walletCoinDeriver == nil {
		err := coinDeriver.Initialize(wallet.mnemonic)
		if err != nil {
			return nil, err
		}
		wallet.derivers[deriverName] = coinDeriver
	} else {
		coinDeriver = wallet.derivers[deriverName]
	}
	return coinDeriver, nil
}

// ChangePassword Change the password of KeyStore.
func ChangePassword(keyStore, oldPassword, newPassword string) (*string, error) {
	privateKey, err := keystore.DecryptKey([]byte(keyStore), oldPassword)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(privateKey.PrivateKey.D.Bytes())
	store, err := createKeyStore(mnemonic, newPassword)
	if err != nil {
		return nil, err
	}
	return store, nil
}

// GetMnemonicFromKeystore Get mnemonic from KeyStore.
func getMnemonicFromKeystore(keystoreStr string, password string) (*string, error) {
	privateKey, err := keystore.DecryptKey([]byte(keystoreStr), password)
	if err != nil {
		log.Printf("error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", keystoreStr)
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	fromEntropy, err := bip39.NewMnemonic(privateKey.PrivateKey.D.Bytes())
	return &fromEntropy, nil
}

func createKeyStore(mnemonic, password string) (*string, error) {
	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	ecdsa := crypto.ToECDSAUnsafe(entropy)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}
	key := keystore.Key{}
	key.Address = crypto.PubkeyToAddress(ecdsa.PublicKey)
	key.Id = guuid.New()
	key.PrivateKey = ecdsa
	encryptKeyStore, err := keystore.EncryptKey(&key, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}
	s := string(encryptKeyStore)
	return &s, nil
}
