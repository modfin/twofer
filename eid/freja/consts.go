package freja

const initAuthURL = "/authentication/1.0/initAuthentication"
const getAuthResultsURL = "/authentication/1.0/getResults"
const getOneAuthResultURL = "/authentication/1.0/getOneResult"
const cancelAuthURL = "/authentication/1.0/cancel"

const initSignURL = "/sign/1.0/initSignature"
const getSignResultsURL = "/sign/1.0/getResults"
const getOneSignResultURL = "/sign/1.0/getOneResult"
const cancelSignURL = "/sign/1.0/cancel"


func Name() string{
	return "FrejaID"
}