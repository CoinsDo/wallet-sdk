package errors

import "errors"

var ErrorCurrencyNotSupported = errors.New("currency not supported")

var ErrorInsufficientFunds = errors.New("insufficient funds available to construct transaction")
var ErrorFeeAddressError = errors.New("fee estimation requires change scripts no larger than P2PKH output scripts")
var ErrorLessThanMinimum = errors.New("less than the minimum quantity")

var ErrorAccountNil = errors.New("toAddress and extraParams.SOLTokenDestAccountOwner all null")

var ErrorNoAvailableAccount = errors.New("none validate token account")

var ErrorDecodeNotSupported = errors.New("decode not supported")

var ErrorInvalidAddress = errors.New("invalid address")

var ErrorInvalidSendAddress = errors.New("invalid send address")

var ErrorInvalidAmount = errors.New("invalid amount")

var ErrorInvalidInput = errors.New("invalid input")

var ErrorInvalidContractAddress = errors.New("invalid contract address")

var ErrorKeyNotFound = errors.New("key not found")
