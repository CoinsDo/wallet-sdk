package coins

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

func UtilCreatePayloadSimpleSend(propertyID uint, amount float64, divisible bool) (string, error) {
	var intPart int64

	if divisible {
		amt, err := btcutil.NewAmount(amount)
		if err != nil {
			return "", err
		}
		intPart = int64(amt)
	} else {
		intPart = int64(amount)
	}
	return fmt.Sprintf("%016x%016x", propertyID, intPart), nil
}

func GetClassCOpreturnDataScript(propertyID uint, amount float64, divisible bool) ([]byte, error) {
	payload, err := UtilCreatePayloadSimpleSend(propertyID, amount, divisible)
	if err != nil {
		return nil, err
	}

	fmt.Println(omniHex + payload)
	b, err := hex.DecodeString(omniHex + payload)

	if err != nil {
		return nil, fmt.Errorf("could not decode payload, %v", err)
	}
	opreturnScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(b).Script()
	if err != nil {
		return nil, fmt.Errorf("failed to create opreturn data, %v", err)
	}
	return opreturnScript, nil
}

type PreviousDependentTxOutputAmount struct {
	TxID   string
	Vout   uint32
	Amount float64
}
