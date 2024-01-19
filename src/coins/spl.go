package coins

import (
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/shopspring/decimal"
	"sort"
	"time"
	"wallet-sdk/src/errors"
	"wallet-sdk/src/types"
)

const CurrencySpl = "SPL"

type SplTxParams struct {
	types.BaseTxParams
	FromAccountOwner string            `json:"solSourceTokenAccountOwner"`
	ToAccountOwner   string            `json:"solDestTokenAccountOwner"`
	ToAccount        string            `json:"toAddress"`
	Amount           decimal.Decimal   `json:"amount"`
	Fee              decimal.Decimal   `json:"fee"`
	Memo             string            `json:"memo"`
	RecentBlockHash  string            `json:"solRecentBlockHash"`
	ContractAddress  string            `json:"commonContractAddress"`
	FromAccounts     []SplTokenAccount `json:"fromAccounts"`
	TokenDecimal     int64             `json:"commonTokenDecimal"`
}

type SplTokenAccount struct {
	TokenAccount string          `json:"tokenAccount"`
	Amount       decimal.Decimal `json:"amount"`
	MintAddress  string          `json:"mintAddress"`
}

var coinSpl Spl

func init() {
	coinSpl = Spl{}
	RegisterCoin(coinSpl)
}

type Spl struct {
	Sol
}

func (coin Spl) GetCurrency() string {
	return CurrencySpl
}

func (coin Spl) GetEmptyTransactionParams() types.TxParams {
	return SolanaTxParams{}
}

func (coin Spl) GetTransactionParamsFromJson(paramsJson string) types.TxParams {
	params := SplTxParams{}
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return nil
	}
	return params
}

func (coin Spl) CreateTransaction(params types.TxParams, testNet bool) (*types.BaseTransaction, error) {
	extraParams := params.(SplTxParams)
	sourceOwnerKey, err := solana.PublicKeyFromBase58(extraParams.FromAccountOwner)
	if err != nil {
		return nil, errors.ErrorInvalidAddress
	}
	builder1 := solana.NewTransactionBuilder()
	if extraParams.ToAccount == "" && extraParams.ToAccountOwner == "" {
		return nil, errors.ErrorAccountNil
	}
	// When toAddress is empty, create a token account for the main account.
	if extraParams.ToAccount == "" {

		ownerKey, err := solana.PublicKeyFromBase58(extraParams.ToAccountOwner)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		mintAccount, err := solana.PublicKeyFromBase58(extraParams.ContractAddress)
		if err != nil {
			return nil, errors.ErrorInvalidContractAddress
		}
		addr, _, err := solana.FindAssociatedTokenAddress(ownerKey, mintAccount)
		if err != nil {
			return nil, err
		}
		extraParams.ToAccount = addr.String()
		builder := associatedtokenaccount.NewCreateInstructionBuilder()
		builder.SetMint(mintAccount)
		builder.SetPayer(sourceOwnerKey)
		meta := solana.NewAccountMeta(addr, true, false)
		builder.Append(meta)
		builder.SetWallet(ownerKey)
		builder1.AddInstruction(builder.Build())

	}
	// Create a timestamp memo to distinguish different transactions.
	meta := solana.NewAccountMeta(sourceOwnerKey, true, true)
	instruction := solana.NewInstruction(solana.MemoProgramID, solana.AccountMetaSlice{meta}, []byte(fmt.Sprint(time.Now().UnixMilli())))
	builder1.AddInstruction(instruction)
	//Create on-chain memo
	if extraParams.Memo != "" {
		builder1.AddInstruction(solana.NewInstruction(solana.MemoProgramID, solana.AccountMetaSlice{meta}, []byte(extraParams.Memo)))
	}
	if len(extraParams.FromAccounts) == 0 {
		return nil, errors.ErrorNoAvailableAccount
	}
	sort.SliceStable(extraParams.FromAccounts, func(i, j int) bool {
		return extraParams.FromAccounts[i].Amount.LessThan(extraParams.FromAccounts[j].Amount)
	})
	//Calculate balances for accounts with the same token type, and provide transfers if the balance is not zero.
	var transferAccounts []SplTokenAccount
	var amountSum = decimal.NewFromInt(0)
	var amountDrops = decimal.NewFromFloat(float64(10)).
		Pow(decimal.NewFromFloat(float64(extraParams.TokenDecimal))).
		Mul(extraParams.Amount)
	for _, account := range extraParams.FromAccounts {
		if decimal.NewFromInt(0).LessThan(account.Amount) {
			amountSum = amountSum.Add(account.Amount)
			if amountSum.GreaterThanOrEqual(extraParams.Amount) {
				sub := amountSum.Sub(amountDrops)
				d := account.Amount.Sub(sub)
				account.Amount = d
				transferAccounts = append(transferAccounts, account)
				break
			} else {
				transferAccounts = append(transferAccounts, account)
			}
		}
	}
	//transaction
	if amountSum.LessThan(extraParams.Amount) {
		return nil, errors.ErrorInsufficientFunds
	}
	for _, account := range transferAccounts {
		builder := token.NewTransferInstructionBuilder()
		builder.SetAmount(uint64(account.Amount.IntPart()))
		builder.SetOwnerAccount(sourceOwnerKey, sourceOwnerKey)
		sourceAccount, err := solana.PublicKeyFromBase58(account.TokenAccount)
		if err != nil {
			return nil, errors.ErrorInvalidSendAddress
		}
		builder.SetSourceAccount(sourceAccount)
		toAccount, err := solana.PublicKeyFromBase58(extraParams.ToAccount)
		if err != nil {
			return nil, errors.ErrorInvalidAddress
		}
		builder.SetDestinationAccount(toAccount)
		build, err := builder.ValidateAndBuild()
		builder1.AddInstruction(build)
	}

	builder1.SetFeePayer(sourceOwnerKey)
	fromBase58, err := solana.HashFromBase58(extraParams.RecentBlockHash)
	if err != nil {
		return nil, err
	}
	builder1.SetRecentBlockHash(fromBase58)
	transaction, err := builder1.Build()
	if err != nil {
		return nil, err
	}
	return &types.BaseTransaction{
		CoinTransaction: transaction,
	}, nil
}
