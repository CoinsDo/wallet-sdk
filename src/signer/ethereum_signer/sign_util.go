package ethereum_signer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"wallet-sdk/src/types"
)

func SignPersonalMessage(message string, privateKey types.PrivateKey) (string, error) {
	var signFormat = "\x19Ethereum Signed Message:\n%d%s"
	newUnsignData := fmt.Sprintf(signFormat, len([]byte(message)), message)
	unsignDataHash := crypto.Keccak256([]byte(newUnsignData))
	return signMessage(unsignDataHash, privateKey)
}

//func SignMessage(message string, privateKey types.PrivateKey) (string, error) {
//	sum256 := sha256.Sum256([]byte(message))
//	return signMessage(sum256[:], privateKey)
//}

func signMessage(message []byte, privateKey types.PrivateKey) (string, error) {
	fmt.Println(hexutil.Encode(message))

	key, err := crypto.ToECDSA(privateKey)
	if err != nil {
		log.Println("sign ToECDSA err:", err.Error())
		return "", err
	}
	signatureByte, err := crypto.Sign(message, key)
	if err != nil {
		log.Println("sign Sign err:", err.Error())
		return "", err
	}
	signatureByte[64] += 27
	return hexutil.Encode(signatureByte), nil
}

func SignTypedData(message []byte, privateKey types.PrivateKey) (string, error) {

	key, err := crypto.ToECDSA(privateKey)
	if err != nil {
		log.Println("sign ToECDSA err:", err.Error())
		return "", err
	}
	signatureByte, err := crypto.Sign(message, key)
	if err != nil {
		log.Println("sign Sign err:", err.Error())
		return "", err
	}
	signatureByte[64] += 27
	return hexutil.Encode(signatureByte), nil
}
