package bankid

import "errors"

var bankidStoppedError = errors.New("bankid client is Stopped")

const authURL = "/rp/v5.1/auth"
const signURL = "/rp/v5.1/sign"
const collectURL = "/rp/v5.1/collect"
const cancelURL = "/rp/v5.1/cancel"

func Name() string {
	return "BankID"
}
