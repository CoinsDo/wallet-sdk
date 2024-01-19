package coins

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"strings"
	errors2 "wallet-sdk/src/errors"
	types2 "wallet-sdk/src/types"
)

func createTransaction(params types2.TxParams) (*types2.BaseTransaction, error) {
	txParams := params.(EthTxParams)

	err := validateAddr(txParams.ToAddress)
	if err != nil {
		return nil, err
	}
	wei := ToWei(txParams.Amount, 18)
	var nonce = txParams.Nonce
	var gasprice = txParams.GasPrice.Int
	var gaslimit = txParams.GasLimit.Int
	address := common.HexToAddress(txParams.ToAddress)
	legacyTx := types.LegacyTx{
		GasPrice: &gasprice,
		Nonce:    uint64(nonce),
		Gas:      gaslimit.Uint64(),
		Value:    wei,
		To:       &address,
	}
	if txParams.Data != "" {
		decode, err := hexutil.Decode(txParams.Data)
		if err != nil {
			return nil, err
		}
		legacyTx.Data = decode
	}

	tx := types.NewTx(&legacyTx)

	transaction := types2.BaseTransaction{}
	transaction.CoinTransaction = tx
	return &transaction, nil

}

func validateAddr(address string) error {
	toAddrBytes, err := hex.DecodeString(strings.TrimPrefix(address, "0x"))
	if err != nil || len(toAddrBytes) != common.AddressLength {
		return errors2.ErrorInvalidAddress
	}
	return nil
}

func createAddress(privateKey *ecdsa.PrivateKey, testNet bool) (*types2.CoinAddress, error) {
	pubKeyHex := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	var adddress = types2.CoinAddress{}
	adddress.AddressStr = strings.ToLower(pubKeyHex)
	return &adddress, nil
}

func signTx(chainId *big.Int, baseTransaction *types2.BaseTransaction, privateKey types2.PrivateKey) (*string, error) {
	toECDSA, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	tx := baseTransaction.CoinTransaction.(*types.Transaction)
	if err != nil {
		return nil, err
	}
	signer := types.NewEIP155Signer(chainId)
	signTx, err := types.SignTx(tx, signer, toECDSA)

	if err != nil {
		log.Fatal(err)
		return nil, nil
	}
	bytes, err := rlp.EncodeToBytes(signTx)

	if err != nil {
		return nil, nil
	}
	rawTxHex := common.Bytes2Hex(bytes)
	var txHex = "0x" + rawTxHex
	return &txHex, nil
}

func DecodeTx(rawTx string) (*types.Transaction, error) {
	trimedTx := strings.TrimPrefix(rawTx, "0x")
	raw, err := hex.DecodeString(trimedTx)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var tx *types.Transaction
	err = rlp.DecodeBytes(raw, &tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// ToWei decimals to wei
func ToWei(iamount interface{}, decimals int64) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromInt(10).Pow(decimal.NewFromInt(decimals))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)
	return wei
}

func createTokenTransaction(params types2.TxParams) (*types2.BaseTransaction, error) {
	extraParams := params.(Erc20TxParams)
	err := validateAddr(extraParams.ToAddress)
	if err != nil {
		return nil, err
	}
	var nonce = extraParams.Nonce
	var gasprice = extraParams.GasPrice.Int

	var gaslimit = extraParams.GasLimit.Int

	var contractAddress = extraParams.ContractAddress
	var tokenDecimal = extraParams.TokenDecimal

	wei := ToWei(extraParams.Amount, tokenDecimal)
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(common.HexToAddress(extraParams.ToAddress).Bytes(), 32)

	paddedAmount := common.LeftPadBytes(wei.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	tx := types.NewTransaction(
		big.NewInt(nonce).Uint64(),
		common.HexToAddress(contractAddress),
		big.NewInt(0), gaslimit.Uint64(), &gasprice,
		data,
	)

	transaction := types2.BaseTransaction{}
	transaction.CoinTransaction = tx
	return &transaction, nil

}

func createErc721TokenTransaction(params types2.TxParams) (*types2.BaseTransaction, error) {
	extraParams := params.(Erc721TxParams)

	err := validateAddr(extraParams.ToAddress)
	if err != nil {
		return nil, err
	}
	var nonce = extraParams.Nonce
	var gasprice = extraParams.GasPrice.Int

	var gaslimit = extraParams.GasLimit.Int

	var contractAddress = extraParams.ContractAddress
	fmt.Println("contractAddress:", contractAddress)

	transferFnSignature := []byte("safeTransferFrom(address,address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(common.HexToAddress(extraParams.ToAddress).Bytes(), 32)

	paddedFromAddress := common.LeftPadBytes(common.HexToAddress(extraParams.FromAddress).Bytes(), 32)

	paddedTokenId := common.LeftPadBytes(extraParams.TokenId.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedFromAddress...)
	data = append(data, paddedAddress...)
	data = append(data, paddedTokenId...)

	fmt.Println("contractAddress-to:", common.HexToAddress(contractAddress).String())
	tx := types.NewTransaction(
		big.NewInt(nonce).Uint64(),
		common.HexToAddress(contractAddress),
		big.NewInt(0), gaslimit.Uint64(), &gasprice,
		data,
	)
	fmt.Println("contractAddress-to:", tx.To().String())
	transaction := types2.BaseTransaction{}
	transaction.CoinTransaction = tx
	return &transaction, nil

}

func createErc1155TokenTransaction(params types2.TxParams) (*types2.BaseTransaction, error) {
	extraParams := params.(Erc1155TxParams)
	err := validateAddr(extraParams.ToAddress)
	if err != nil {
		return nil, err
	}
	if len(extraParams.BatchData) == 0 {
		return nil, errors.New("no token data")
	}

	if len(extraParams.BatchData) > 1 {
		return createErc1155BatchTransaction(params)
	}

	tokenData := extraParams.BatchData[0]

	var nonce = extraParams.Nonce
	var gasprice = extraParams.GasPrice.Int

	var gaslimit = extraParams.GasLimit.Int

	var contractAddress = extraParams.ContractAddress

	var method = `[	{ "type" : "function", "name" : "safeTransferFrom", "inputs" : [  { "name" : "from", "type" : "address" },{ "name" : "_to", "type" : "address" },{ "name" : "_tokenId", "type" : "uint256" },{ "name" : "amount", "type" : "uint256" },{ "name" : "data", "type" : "bytes" } ] }]`
	json, err := abi.JSON(strings.NewReader(method))
	if err != nil {
		return nil, err
	}

	bytes, err := json.Pack("safeTransferFrom", common.HexToAddress(extraParams.FromAddress),
		common.HexToAddress(extraParams.ToAddress), &tokenData.TokenId.Int, &tokenData.Amount.Int, []byte(""))
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		big.NewInt(nonce).Uint64(),
		common.HexToAddress(contractAddress),
		big.NewInt(0), gaslimit.Uint64(), &gasprice,
		bytes,
	)

	transaction := types2.BaseTransaction{}
	transaction.CoinTransaction = tx
	return &transaction, nil

}

func createErc1155BatchTransaction(params types2.TxParams) (*types2.BaseTransaction, error) {
	extraParams := params.(Erc1155TxParams)
	err := validateAddr(extraParams.ToAddress)
	if err != nil {
		return nil, err
	}
	var nonce = extraParams.Nonce
	var gasprice = extraParams.GasPrice.Int

	var gaslimit = extraParams.GasLimit.Int

	var contractAddress = extraParams.ContractAddress

	var method = `[	{ "type" : "function", "name" : "safeBatchTransferFrom", "inputs" : [  { "name" : "from", "type" : "address" },{ "name" : "_to", "type" : "address" },{ "name" : "_tokenId", "type" : "uint256[]" },{ "name" : "amount", "type" : "uint256[]" },{ "name" : "data", "type" : "bytes" } ] }]`
	json, err := abi.JSON(strings.NewReader(method))
	if err != nil {
		return nil, err
	}
	var tokenIds []*big.Int
	var amounts []*big.Int
	for index := range extraParams.BatchData {
		var data = extraParams.BatchData[index]
		tokenIds = append(tokenIds, &data.TokenId.Int)
		amounts = append(amounts, &data.Amount.Int)
	}

	bytes, err := json.Pack("safeBatchTransferFrom", common.HexToAddress(extraParams.FromAddress),
		common.HexToAddress(extraParams.ToAddress), tokenIds, amounts, []byte(""))
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		big.NewInt(nonce).Uint64(),
		common.HexToAddress(contractAddress),
		big.NewInt(0), gaslimit.Uint64(), &gasprice,
		bytes,
	)

	transaction := types2.BaseTransaction{}
	transaction.CoinTransaction = tx
	return &transaction, nil

}
