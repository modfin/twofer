package bankidm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type AuthRequest struct {
	EndUserIP   string      `json:"endUserIp"`
	SSN         string      `json:"personalNumber,omitempty"`
	Requirement Requirement `json:"requirement,omitempty"`
}

func (a AuthRequest) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

type SignRequest struct {
	EndUserIP          string      `json:"endUserIp"`
	SSN                string      `json:"personalNumber,omitempty"`
	Requirement        Requirement `json:"requirement,omitempty"`
	UserVisibleData    string      `json:"userVisibleData"`
	UserNonVisibleData []byte      `json:"userNonVisibleData,omitempty"`
}

func (a SignRequest) Marshal() ([]byte, error) {
	if len(a.UserVisibleData) > 0 {
		a.UserVisibleData = base64.StdEncoding.EncodeToString([]byte(a.UserVisibleData))
	}
	return json.Marshal(a)
}

type Requirement struct {
}

type OrderRef struct {
	OrderRef string `json:"orderRef"`
}

func (a OrderRef) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

type SignResponse AuthResponse
type AuthResponse struct {
	OrderRef
	AutoStartToken string `json:"autoStartToken"`
}

type Status string

const (
	STATUS_PENDING  Status = "pending"
	STATUS_FAILED   Status = "failed"
	STATUS_COMPLETE Status = "complete"
)

type HintCode string

const (
	// Pending hints
	HINT_OUTSTANDING HintCode = "outstandingTransaction"
	HINT_NO_CLIENT   HintCode = "noClient"
	HINT_STARTED     HintCode = "started"
	HINT_USER_SIGN   HintCode = "userSign"

	// Failed hints
	HINT_EXPIRED      HintCode = "expiredTransaction"
	HINT_CERT_ERR     HintCode = "certificateErr"
	HINT_USER_CANCEL  HintCode = "userCancel"
	HINT_CANCELED     HintCode = "cancelled"
	HINT_START_FAILED HintCode = "startFailed"
)

type CollectResponse struct {
	OrderRef
	Status         Status         `json:"status"`
	HintCode       HintCode       `json:"hintCode"`
	CompletionData CompletionData `json:"completionData"`
}

type CompletionData struct {
	User         User   `json:"user"`
	Device       Device `json:"device"`
	Cert         Cert   `json:"cert"`
	Signature    string `json:"signature"`
	OCSPResponse string `json:"ocspResponse"`
}
type User struct {
	PersonalNumber string `json:"personalNumber"`
	Name           string `json:"name"`
	GivenName      string `json:"givenName"`
	Surname        string `json:"surname"`
}

type Device struct {
	IPAddress string `json:"ipAddress"`
}

type Cert struct {
	NotBefore string `json:"notBefore"`
	NotAfter  string `json:"notAfter"`
}

type BankIdError struct {
	ErrorCode string `json:"errorCode"`
	Details   string `json:"details"`
}

func (e BankIdError) Error() string {
	return fmt.Sprintf("bankid: %s, %s", e.ErrorCode, e.Details)
}
