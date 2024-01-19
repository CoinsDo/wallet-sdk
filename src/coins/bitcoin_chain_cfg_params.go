package coins

import (
	"github.com/btcsuite/btcd/chaincfg"
)

func DogeTestNet3Params() chaincfg.Params {

	var params = chaincfg.Params{}
	params.Name = "org.dogecoin.test"
	params.Net = 0xfcc1b7dc
	params.PubKeyHashAddrID = 0x71
	params.ScriptHashAddrID = 0xc4
	params.PrivateKeyID = 0xf1
	params.Bech32HRPSegwit = "doget"
	params.HDPrivateKeyID = [4]byte{0x04, 0x35, 0x87, 0xcf}
	params.HDPublicKeyID = [4]byte{0x04, 0x35, 0x83, 0x94}
	return params
}

func DogeMainNetParams() chaincfg.Params {

	var params = chaincfg.TestNet3Params
	params.Name = "doge"
	params.Net = 0xc0c0c0c0
	params.PubKeyHashAddrID = 0x1e
	params.ScriptHashAddrID = 0x16
	params.PrivateKeyID = 0x9e
	params.HDPrivateKeyID = [4]byte{0x02, 0xfa, 0xc3, 0x98}
	params.HDPublicKeyID = [4]byte{0x02, 0xfa, 0xca, 0xfd}
	params.Bech32HRPSegwit = "doge"
	return params
}

func DashTestNet3Params() chaincfg.Params {

	var params = chaincfg.Params{}
	params.Name = "org.litecoin.test"
	params.Net = 0xcee2caff
	params.PubKeyHashAddrID = 0x8c
	params.ScriptHashAddrID = 0x13
	params.PrivateKeyID = 0xef
	params.Bech32HRPSegwit = "dasht"
	params.HDPrivateKeyID = [4]byte{0x04, 0x35, 0x87, 0xcf}
	params.HDPublicKeyID = [4]byte{0x04, 0x35, 0x83, 0x94}
	return params
}

func DashMainNetParams() chaincfg.Params {

	var params = chaincfg.TestNet3Params
	params.Name = "dash"
	params.Net = 0xbf0c6bbd
	params.PubKeyHashAddrID = 0x4c
	params.ScriptHashAddrID = 0x10
	params.PrivateKeyID = 0xcc
	params.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	params.HDPublicKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	params.Bech32HRPSegwit = "dash"
	return params
}

func BsvMainNetParams() chaincfg.Params {

	var params = chaincfg.TestNet3Params
	params.Name = "mainnet"
	params.Net = 0xe8f3e1e3
	params.PubKeyHashAddrID = 0x00
	params.ScriptHashAddrID = 0x05
	params.PrivateKeyID = 0x80
	params.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	params.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
	params.Bech32HRPSegwit = "dash"
	return params
}

func BsvTestNet3Params() chaincfg.Params {

	var params = chaincfg.TestNet3Params
	params.Name = "testnet3"
	params.Net = 0xf4f3e5f4
	params.PubKeyHashAddrID = 0x6f
	params.ScriptHashAddrID = 0xc4
	params.PrivateKeyID = 0xef
	params.HDPrivateKeyID = [4]byte{0x04, 0x35, 0x83, 0x94}
	params.HDPublicKeyID = [4]byte{0x04, 0x35, 0x87, 0xcf}
	params.Bech32HRPSegwit = "dash"
	return params
}
