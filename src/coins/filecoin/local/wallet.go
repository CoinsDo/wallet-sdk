package local

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"wallet-sdk/src/coins/filecoin/sigs"
	_ "wallet-sdk/src/coins/filecoin/sigs/secp"
	"wallet-sdk/src/coins/filecoin/types"
)

// WalletSign signs the given bytes using the KeyType and private key.
func WalletSign(typ types.KeyType, pk []byte, data []byte) (*crypto.Signature, error) {
	return sigs.Sign(ActSigType(typ), pk, data)
}

// WalletVerify verify the signed message
func WalletVerify(sig *crypto.Signature, addr address.Address, msg []byte) error {
	return sigs.Verify(sig, addr, msg)
}

// WalletSignMessage signs the given message using the given private key.
func WalletSignMessage(typ types.KeyType, pk []byte, msg *types.Message) (*types.SignedMessage, error) {
	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, fmt.Errorf("serializing message: %w", err)
	}

	sig, err := WalletSign(typ, pk, mb.Cid().Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return &types.SignedMessage{
		Message:   msg,
		Signature: sig,
	}, nil
}

func WalletVerifyMessage(sm *types.SignedMessage) error {
	return WalletVerify(sm.Signature, sm.Message.From, sm.Message.Cid().Bytes())
}

func ActSigType(typ types.KeyType) crypto.SigType {
	switch typ {
	case types.KTBLS:
		return crypto.SigTypeBLS
	case types.KTSecp256k1:
		return crypto.SigTypeSecp256k1
	default:
		return crypto.SigTypeUnknown
	}
}
