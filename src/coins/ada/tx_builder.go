package ada

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/echovl/cardano-go/crypto"
	"golang.org/x/crypto/blake2b"
	"wallet-sdk/src/errors"
)

const maxUint64 uint64 = 1<<64 - 1

type txBuilderInput struct {
	input  transactionInput
	amount uint64
}

type txBuilderOutput struct {
	address string
	amount  uint64
}

type txBuilder struct {
	tx       Transaction
	protocol ProtocolParams
	inputs   []txBuilderInput
	outputs  []txBuilderOutput
	ttl      uint64
	fee      uint64
	vkeys    map[string]crypto.ExtendedVerificationKey
	pkeys    map[string]crypto.ExtendedSigningKey
}

func NewTxBuilder(protocol ProtocolParams) *txBuilder {
	return &txBuilder{
		protocol: protocol,
		vkeys:    map[string]crypto.ExtendedVerificationKey{},
		pkeys:    map[string]crypto.ExtendedSigningKey{},
	}
}

func (builder *txBuilder) AddInput(xvk []byte, txId TransactionID, index, amount uint64) {
	input := txBuilderInput{input: transactionInput{ID: txId.Bytes(), Index: index}, amount: amount}
	builder.inputs = append(builder.inputs, input)

	vkeyHashBytes := blake2b.Sum256(xvk)
	vkeyHashString := hex.EncodeToString(vkeyHashBytes[:])
	builder.vkeys[vkeyHashString] = xvk
}

func (builder *txBuilder) AddOutput(address string, amount uint64) {
	output := txBuilderOutput{address: address, amount: amount}
	builder.outputs = append(builder.outputs, output)
}

func (builder *txBuilder) SetTtl(ttl uint64) {
	builder.ttl = ttl
}

func (builder *txBuilder) SetFee(fee uint64) {
	builder.fee = fee
}

func (builder *txBuilder) AddFee(address string) error {
	builder.SetFee(maxUint64)
	minFee := builder.CalculateMinFee()
	inputAmount := uint64(0)
	for _, txIn := range builder.inputs {
		inputAmount += txIn.amount
	}

	outputAmount := uint64(0)
	for _, txOut := range builder.outputs {
		outputAmount += txOut.amount
	}
	outputWithFeeAmount := outputAmount + minFee

	if inputAmount > outputWithFeeAmount {
		minAda := builder.protocol.MinimumUtxoValue
		change := inputAmount - outputWithFeeAmount
		if change > minAda {
			feeChange := builder.feeForOuput(address, change)
			newFee := minFee + feeChange
			change = inputAmount - (outputAmount + newFee)

			if change > minAda {
				if address != "" && &address != nil {
					builder.AddOutput(address, change)
				}
				builder.SetFee(newFee)
			} else {
				builder.SetFee(minFee + change)
			}
		} else {
			builder.SetFee(minFee + change)
		}

	} else if inputAmount == outputWithFeeAmount {
		builder.SetFee(minFee)
	} else {
		builder.SetFee(minFee)
		return errors.ErrorInsufficientFunds
	}

	return nil
}

func (builder *txBuilder) CalculateMinFee() uint64 {
	fakeXSigningKey := crypto.NewExtendedSigningKey([]byte{
		0x0c, 0xcb, 0x74, 0xf3, 0x6b, 0x7d, 0xa1, 0x64, 0x9a, 0x81, 0x44, 0x67, 0x55, 0x22, 0xd4, 0xd8, 0x09, 0x7c, 0x64, 0x12,
	}, "")

	body, err := builder.buildBody()
	if err != nil {
		fmt.Println(err)
	}

	witnessSet := transactionWitnessSet{}
	for range builder.inputs {
		witness := vkeyWitness{VKey: fakeXSigningKey.ExtendedVerificationKey()[:32], Signature: fakeXSigningKey.Sign(fakeXSigningKey.ExtendedVerificationKey()[:32])}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	tx := Transaction{
		Body:       *body,
		WitnessSet: witnessSet,
		Metadata:   nil,
	}

	return builder.CalculateFee(&tx)
}

func (builder *txBuilder) feeForOuput(address string, amount uint64) uint64 {
	builderCpy := *builder

	builderCpy.SetFee(0)

	feeBefore := builderCpy.CalculateMinFee()
	builderCpy.AddOutput(address, amount)
	feeAfter := builderCpy.CalculateMinFee()

	return feeAfter - feeBefore
}

func (builder *txBuilder) CalculateFee(tx *Transaction) uint64 {

	var size = uint64(69*len(tx.Body.Outputs) + 87 + 137*len(tx.Body.Inputs))

	return builder.protocol.MinFeeA*size + builder.protocol.MinFeeB
}

func (builder *txBuilder) Build() Transaction {
	body, _ := builder.buildBody()
	return Transaction{Body: *body, Metadata: nil}
}

func (builder *txBuilder) buildBody() (*transactionBody, error) {
	inputs := make([]transactionInput, len(builder.inputs))
	for i, txInput := range builder.inputs {
		inputs[i] = transactionInput{
			ID:    txInput.input.ID,
			Index: txInput.input.Index,
		}
	}

	outputs := make([]transactionOutput, len(builder.outputs))
	for i, txOutput := range builder.outputs {
		addressBytes, err := GetAddressBytes(txOutput.address)
		if err != nil {
			return nil, err
		}
		outputs[i] = transactionOutput{
			Address: addressBytes,
			Amount:  txOutput.amount,
		}
	}
	return &transactionBody{
		Inputs:  inputs,
		Outputs: outputs,
		Fee:     builder.fee,
		Ttl:     builder.ttl,
	}, nil
}

func pretty(v interface{}) string {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes)
}
