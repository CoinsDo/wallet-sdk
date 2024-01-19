package coins

import (
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/btcsuite/btcd/btcec"
	"github.com/eoscanada/eos-go/btcsuite/btcutil"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"github.com/eoscanada/eos-go/token"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wallet-sdk/src/deriver"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

type EOSAction int

const (
	EOSAction_Transaction EOSAction = 0
	EOSAction_Stake       EOSAction = 1
	EOSAction_BuyRam      EOSAction = 2
	EOSAction_Unstake     EOSAction = 3
	EOSAction_SellRam     EOSAction = 4
	CurrencyEos                     = "EOS"
)

type SignParams struct {
	ChainId     string
	HeadBlockId string
	//LastIrreversibleBlockNum string
	RefBlockPrefix uint32
	Exp            uint64
}

type EosTxParams struct {
	types.BaseTxParams

	FromAddress    string          `json:"fromAddress"`
	ToAddress      string          `json:"toAddress"`
	Amount         decimal.Decimal `json:"amount"`
	TokenDecimal   int64           `json:"commonTokenDecimal"`
	PrivateKey     string          `json:"commonPrivateKey"`
	Action         EOSAction       `json:"eosAction"`
	Memo           string          `json:"memo"`
	StakeCpuAmount float32         `json:"eosStakeCpuAmount"`
	StakeNetAmount float32         `json:"eosStakeNetAmount"`
	BuyRamAmount   float32         `json:"eosBuyRamAmount"`
	SellRamBytes   uint64          `json:"eosSellRamBytes"`
	Token          string          `json:"eosToken"` // Currently only supports EOS, use "EOS"
	SignParams     SignParams      `json:"signParams"`
	ExpireDuration int64           `json:"expireDuration"` // EOS timeout duration, set as a duration. The calculation involves adding ExpireDuration to the current time to determine the timeout time. Unit is in milliseconds.

}

type EOSTransaction struct {
	Compression eos.CompressionType `json:"compression"`
	Signatures  []ecc.Signature     `json:"signatures"`
	Transaction eos.Transaction     `json:"transaction"`
	ChainID     eos.Checksum256     `json:"-"`
}

var coinEos Eos

func init() {
	coinEos = Eos{}
	RegisterCoin(coinEos)
}

type Eos struct {
}

func (coin Eos) GetCurrency() string {

	return CurrencyEos
}

func (coin Eos) GetPath(index int64, testNet bool) string {
	return fmt.Sprintf(coin.GetBasePath(testNet), 0, 0, index)
}

func (coin Eos) GetBasePath(testNet bool) string {
	return "m/44'/194'/%d'/%d/%d"
}

func (coin Eos) ChainName() string {
	return CurrencyEos
}

func (coin Eos) GetDecimal() int {
	return 4
}

func (coin Eos) PrivateKeyFromString(key string) (types.PrivateKey, error) {
	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		return nil, err
	}
	return wif.PrivKey.Serialize(), nil
}

func (coin Eos) PrivateKeyToString(key types.PrivateKey) (string, error) {
	ecdsaPrivateKey, err := crypto.ToECDSA(key)
	if err != nil {
		return "", err
	}
	privateKeyBytes := crypto.FromECDSA(ecdsaPrivateKey)

	publicKey := ecdsaPrivateKey.PublicKey

	btcPrivKey, _ := btcec.PrivKeyFromBytes(publicKey.Curve, privateKeyBytes)
	wif, err := btcutil.NewWIF(btcPrivKey, 0x80, true)
	return wif.String(), nil
}

func (coin Eos) GenerateAddress(privateKey types.PrivateKey, testNet bool) (*types.CoinAddress, error) {
	ecdsaPrivateKey, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(ecdsaPrivateKey)

	publicKey := ecdsaPrivateKey.PublicKey

	btcPrivKey, _ := btcec.PrivKeyFromBytes(publicKey.Curve, privateKeyBytes)
	wif, err := btcutil.NewWIF(btcPrivKey, 0x80, true)

	if err != nil {
		return nil, err
	}
	key, err := ecc.NewPrivateKey(wif.String())

	if err != nil {
		return nil, err
	}

	var adddress = types.CoinAddress{}
	adddress.AddressStr = key.PublicKey().String()
	return &adddress, nil
}

func (coin Eos) GetDeriver() deriver.Deriver {
	return &deriver.Bip39Deriver{}
}

func (coin Eos) GetEmptyTransactionParams() types.TxParams {
	return EosTxParams{}
}

func (coin Eos) CreateTransaction(baseTxParams types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := baseTxParams.(EosTxParams)
	signParams := extraParams.SignParams
	err := validateEOSAccount(extraParams.FromAddress)
	if err != nil {
		return nil, errors.ErrorInvalidSendAddress
	}
	err = validateEOSAccount(extraParams.ToAddress)
	if err != nil {
		return nil, err
	}
	from := eos.AccountName(extraParams.FromAddress)
	to := eos.AccountName(extraParams.ToAddress)

	var parser = "%." + strconv.Itoa(int(extraParams.TokenDecimal)) + "f"
	f, _ := extraParams.Amount.Float64()
	input := fmt.Sprintf(parser, f) + " " + extraParams.Token
	fromString, err := eos.NewEOSAssetFromString(input)
	if err != nil {
		return nil, err
	}

	s := fromString.String()
	fmt.Println(s)
	var chainIdCheckSum eos.Checksum256
	if !strings.HasPrefix(signParams.ChainId, "0x") {
		signParams.ChainId = "0x" + signParams.ChainId
	}
	decode, err := hexutil.Decode(signParams.ChainId)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(signParams.HeadBlockId, "0x") {
		signParams.HeadBlockId = "0x" + signParams.HeadBlockId
	}
	var headBlockIdSum eos.Checksum256
	decodeHeadBlockIdBytes, err := hexutil.Decode(signParams.HeadBlockId)
	if err != nil {
		return nil, err
	}

	txOpts := eos.TxOptions{
		ChainID:          append(chainIdCheckSum, decode...),
		DelaySecs:        0,
		MaxCPUUsageMS:    0,
		MaxNetUsageWords: 0,
		Compress:         eos.CompressionNone,
		HeadBlockID:      append(headBlockIdSum, decodeHeadBlockIdBytes...),
	}

	var actions []*eos.Action
	if extraParams.Action == EOSAction_Transaction {
		actions = append(actions, token.NewTransfer(from, to, fromString, extraParams.Memo))
	} else if extraParams.Action == EOSAction_Stake {

		stakeCpuAmountAsset, err := eos.NewEOSAssetFromString(fmt.Sprintf(parser, extraParams.StakeCpuAmount) + " " + "EOS")
		if err != nil {
			return nil, err
		}
		stakeNetAmountAsset, err := eos.NewEOSAssetFromString(fmt.Sprintf(parser, extraParams.StakeNetAmount) + " " + "EOS")
		if err != nil {
			return nil, err
		}
		bw := system.NewDelegateBW(from, to, stakeCpuAmountAsset, stakeNetAmountAsset, false)

		actions = append(actions, bw)
	} else if extraParams.Action == EOSAction_Unstake {

		unstakeCpuAmountAsset, err := eos.NewEOSAssetFromString(fmt.Sprintf(parser, extraParams.StakeCpuAmount) + " " + "EOS")
		if err != nil {
			return nil, err
		}
		unstakeNetAmountAsset, err := eos.NewEOSAssetFromString(fmt.Sprintf(parser, extraParams.StakeNetAmount) + " " + "EOS")
		if err != nil {
			return nil, err
		}
		bw := system.NewUndelegateBW(from, to, unstakeCpuAmountAsset, unstakeNetAmountAsset)

		actions = append(actions, bw)
	} else if extraParams.Action == EOSAction_BuyRam {
		buyRamAmountAsset, err := eos.NewEOSAssetFromString(fmt.Sprintf(parser, extraParams.BuyRamAmount) + " " + "EOS")
		if err != nil {
			return nil, err
		}
		ram := system.NewBuyRAM(from, to, uint64(buyRamAmountAsset.Amount))
		actions = append(actions, ram)
	} else if extraParams.Action == EOSAction_SellRam {
		ram := system.NewSellRAM(from, extraParams.SellRamBytes)
		actions = append(actions, ram)
	}
	tx := eos.NewTransaction(actions, &txOpts)
	i := time.Duration(extraParams.ExpireDuration) * time.Millisecond
	tx.SetExpiration(i)
	transaction := EOSTransaction{
		ChainID:     txOpts.ChainID,
		Transaction: *tx,
	}

	return &types.BaseTransaction{CoinTransaction: transaction}, nil
}

func (coin Eos) DecodeTransaction(rawTx string, testnet bool) (interface{}, error) {

	return nil, errors.ErrorDecodeNotSupported
}

func (coin Eos) SignTx(baseTransaction *types.BaseTransaction, testNet bool, privateKey types.PrivateKey) (*string, error) {
	eosTransaction := baseTransaction.CoinTransaction.(EOSTransaction)
	ecdsaPrivateKey, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := crypto.FromECDSA(ecdsaPrivateKey)

	publicKey := ecdsaPrivateKey.PublicKey

	btcPrivKey, _ := btcec.PrivKeyFromBytes(publicKey.Curve, privateKeyBytes)
	wif, err := btcutil.NewWIF(btcPrivKey, 0x80, true)

	transaction, err := coin.signTransaction(&eosTransaction.Transaction, wif.String(), eosTransaction.ChainID)
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(transaction)
	s := string(marshal)
	return &s, err
}

func (coin Eos) signTransaction(tx *eos.Transaction, privateKey string, chainId eos.Checksum256) (*EOSTransaction, error) {
	signedTx := eos.NewSignedTransaction(tx)
	bag := eos.NewKeyBag()

	err := bag.Add(privateKey)
	key, err := ecc.NewPrivateKey(privateKey)

	_, _ = bag.Sign(signedTx, chainId, key.PublicKey())

	pack, err := signedTx.Pack(eos.CompressionNone)
	if err != nil {
		return nil, err
	}

	transaction := EOSTransaction{
		Compression: pack.Compression,
		Signatures:  pack.Signatures,
		Transaction: *tx,
	}

	return &transaction, nil
}

func (coin Eos) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := EosTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func validateEOSAccount(account string) error {
	eosAccountRegex := "^[a-z1-5]{12}$"
	match, _ := regexp.MatchString(eosAccountRegex, account)
	if !match {
		return errors.ErrorInvalidAddress
	}
	return nil
}
