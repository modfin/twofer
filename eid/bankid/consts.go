package bankid

import "errors"

var bankidStoppedError = errors.New("bankid client is Stopped")

const authURL = "/rp/v5/auth"
const signURL = "/rp/v5/sign"
const collectURL = "/rp/v5/collect"
const cancelURL = "/rp/v5/cancel"

func Name() string{
	return "BankID"
}