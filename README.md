# About CoinsDo Wallet SDK
CoinsDo Wallet SDK is a powerful open-source, cross-platform library. It encompasses the core wallet functionalities, spanning from wallet creation and key pair generation to transaction execution. 

This SDK is at the heart of various CoinsDo wallet products, including [CoinSend](https://www.coinsdo.com/wallet_dispatching.html), [CoinGet](https://www.coinsdo.com/wallet_collecting.html), and [CoinWallet](https://www.coinsdo.com/wallet_coinsdo.html). It is primarily written in Golang, offering seamless integration and support across multiple platforms, including iOS, Android, and the web extension.

## Supported Blockchains
CoinsDo Wallet SDK supports a wide array of blockchains, including but not limited to:

- Bitcoin (BTC)
- Dogecoin (DOGE)
- Dash (DASH)
- Litecoin (LTC)
- Bitcoin SV (BSV)
- Bitcoin Cash (BCH)
- Ethereum (ETH)
- TRON (TRX)
- Binance Smart Chain (BSC)
- Solana (SOL)
- Polkadot (DOT)
- Cardano (ADA)
- Arbitrum (ARB1)
- Avalanche C-Chain (AVAXC)
- EOS
- Ethereum Classic (ETC)
- Filecoin (FILECOIN)
- Fantom (FTM)
- xDai (XDAI)
- NEAR Protocol (NEAR)
- Optimism (OPT)
- XRP
- Zcash (ZEC)

## EVM-compatible blockchains
You can dynamically add EVM-compatible blockchains according to your requirements. 

## Functions

### Create a new wallet
```sh
createdWallet, err := wallet.New()
```

### Import wallet from mnemonic
```sh
mnemonicWallet, err := wallet.NewFromMnemonic(mnemonic)
```
### Import wallet from keystore
```sh
keystoreWallet, err := wallet.NewFromKeystore(keystore, password)
```

### Export keystore
```sh
keystoreString, err := mnemonicWallet.ExportKeyStore(password)
```

### Generate address
```sh
key, path, err := wallet.DerivePrivateKey(coins.CurrencyTrx, int64(i), testnet)
```

### Custom derivation path
```sh
var customPath = `m/44'/195'/0'/0`
key, path, err := wallet.DerivePrivateKeyByPath(coins.CurrencyTrx, customPath, testnet)
```

### Generate address from a private key
```sh
coin, err := coins.GetCoin(coins.CurrencyTrx)
address, err := coin.GenerateAddress(key, testnet)
```

### Generate Bitcoin SegWit address
```sh
coin, err := coins.GetCoin(coins.CurrencyBtc)
address, err := coin.GenerateNestedSegitAddress(key, testnet)
```

### Create transaction
```sh
coin, err := coins.GetCoin(coins.CurrencyTrx)
var trxTransactionparams = coins.TrxTxParams{
    	ToAddress:             address,
		Amount:                 decimal.NewFromFloat(2000),
		FromAddress:            address,
		Fee:                    decimal.NewFromInt(1),
		ContractType:           core.Transaction_Contract_FreezeBalanceV2Contract,
		TronFreezeResourceCode: core.ResourceCode_ENERGY,
		ExpireTime:             time.Now().UnixMilli() + 60*60*1000*10,
		BlockData: Block{
			BlockId:        "000000000277adfabcc24c7b065635f3f125c364238ddcfa58ba59368379097c",
			BlockNumber:    41397754,
			BlockTimeStamp: time.Now().UnixMilli(),
		},

}

createTransaction, err := coin.CreateTransaction(params, testNet)
```

### Sign transaction
```sh
tx, err := coin.SignTx(createTransaction, testNet, key)
```

### Sign a transaction with multiple Bitcoin addresses
```sh
coin, err := coins.GetCoin(coins.CurrencyBtc)
var privateKeys = map[string]types.PrivateKey{}  // [address] privateKey
privateKeys[addr1] = key1
privateKeys[addr2] = key2
if multiSenderSigner, ok := coin.(coins.BitcoinMultiSenderSigner); ok {
    tx, err := multiSenderSigner.SignMultipleSendAddressTx(createTransaction, testNet, privateKeyBytesMap)
} 
```

## Working offline
CoinsDo Wallet SDK able to operates entirely offline. Hence, you can build your own customized hot or cold wallet tailored to your specific needs. 

## License Information
CoinsDo Wallet SDK is released under the Apache License 2.0. This license provides you with the freedom to use, modify, and distribute the software, either for personal or commercial purposes. However, it is crucial to note that this SDK incorporates third-party extensions, each governed by its own license.

See [LICENSE](LICENSE) file for more info

Before utilizing this SDK for commercial purposes, it is imperative to conduct a thorough check of the licenses associated with the third-party extensions included. Each extension may have its licensing terms, and please do your due diligence to ensure compliance with those terms.

To facilitate this process, we have included a comprehensive list of third-party licenses in the repository. Please refer to the [3rd-party-licenses](3rd-party-licenses.md) file for detailed information on the licenses of the included extensions.

## Software Security 
CoinsDo takes the security of our software seriously. While we strive to minimize risks and vulnerabilities in CoinsDo Wallet SDK, it is essential to acknowledge that no system is entirely free from potential security threats.

### Continuous Security Improvement
CoinsDo’s offerings leverage SDK versions that are several iterations ahead, ensuring the code is in optimal condition for release as open source.We are committed to continuously improving the security of CoinsDo Wallet SDK. If you discover any security vulnerabilities or issues, please responsibly disclose them by contacting our team at [cs@coinsdo.com](mailto:cs@coinsdo.com). Your cooperation helps us maintain the integrity and security of the CoinsDo Wallet SDK for the entire community.

### No Reported Security Incidents
As of our knowledge cutoff date in January 2022, CoinsDo Wallet SDK has not experienced any reported security incidents. We will continue to monitor and address security concerns proactively to provide a secure development environment for our users.

## Accountability Disclaimer
We make every effort to implement robust security measures and regularly update the SDK to address potential vulnerabilities. However, CoinsDo, the maintainers, and contributors to the CoinsDo Wallet SDK are not accountable for any security breaches or incidents that may occur during the use of this SDK. 

## Contributing to CoinsDo Wallet SDK
CoinsDo Wallet SDK, designed for sharing and educational purposes, is an open-source project that welcomes contributions from the community. Whether you are a developer, product designer, tester, or enthusiast, your involvement can play a crucial role in enhancing the features and reliability of this SDK.

## Issues and Feature Requests
If you encounter bugs or have ideas for new features, feel free to open an issue on our GitHub repository. We appreciate your feedback and are responsive to community input.

We appreciate your collaboration in CoinsDo Wallet SDK, and together we create a secure and reliable platform for blockchain development.

## Get in Touch
Join the discussion on our Github Forum or engage with us on [Twitter](https://twitter.com/CoinsDogroup) or email at [cs@coinsdo.com](mailto:cs@coinsdo.com). Your questions, suggestions, and feedback are invaluable to the growth and improvement of CoinsDo Wallet SDK.

Thank you for considering contributing to CoinsDo Wallet SDK. Together, we can build a robust and feature-rich platform for the blockchain development community!


