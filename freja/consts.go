package freja

import "errors"

var frejaStoppedError = errors.New("freja client is Stopped")

const pubSubAuthPrefix = "freja:eid:authref:"
const pubSubSignPrefix = "freja:eid:signref:"

const TestURL = "https://services.test.frejaeid.com/"
const ProdURL = "https://services.prod.frejaeid.com/"

const ProdResourceURL = "https://resources.prod.frejaeid.com/"
const TestResourceURL = "https://resources.test.frejaeid.com/"



const initAuthURL = "authentication/1.0/initAuthentication"
const getAuthResultsURL = "authentication/1.0/getResults"
const getOneAuthResultURL = "authentication/1.0/getOneResult"
const cancelAuthURL = "authentication/1.0/cancel"

const initSignURL = "sign/1.0/initSignature"
const getSignResultsURL = "sign/1.0/getResults"
const getOneSignResultURL = "sign/1.0/getOneResult"
const cancelSignURL = "sign/1.0/cancel"
