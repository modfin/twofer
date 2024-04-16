package bankid

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type AuthSignRequest struct {
	EndUserIp             string      `json:"endUserIp"`
	Requirement           Requirement `json:"requirement,omitempty"`
	UserVisibleData       string      `json:"userVisibleData,omitempty"`
	UserNonVisibleData    string      `json:"userNonVisibleData,omitempty"`
	UserVisibleDataFormat string      `json:"userVisibleDataFormat,omitempty"`
}

func (r *AuthSignRequest) ValidateAuthRequest() error {
	if r.EndUserIp == "" {
		return errors.New("missing ip address")
	}

	if r.UserVisibleData != "" {
		data := base64.StdEncoding.EncodeToString([]byte(r.UserVisibleData))
		if len(data) > 1500 {
			return errors.New("userVisibleData is more than 1500 characters long")
		}
	}

	if r.UserNonVisibleData != "" {
		data := base64.StdEncoding.EncodeToString([]byte(r.UserNonVisibleData))
		if len(data) > 1500 {
			return errors.New("userNonVisibleData is more than 1500 characters long")
		}
	}

	if r.UserVisibleDataFormat != "" && r.UserVisibleDataFormat != "simpleMarkdownV1" {
		return errors.New("invalid userVisibleDataFormat")
	}

	return nil
}

func (r *AuthSignRequest) ValidateSignRequest() error {
	if r.EndUserIp == "" {
		return errors.New("missing ip address")
	}

	if r.UserVisibleData == "" {
		return errors.New("missing userVisibleData")
	}

	visibleData := base64.StdEncoding.EncodeToString([]byte(r.UserVisibleData))
	if len(visibleData) > 40000 {
		return errors.New("userVisibleData is more than 40 000 characters long")
	}

	if r.UserNonVisibleData != "" {
		data := base64.StdEncoding.EncodeToString([]byte(r.UserNonVisibleData))
		if len(data) > 200000 {
			return errors.New("userNonVisibleData is more than 200 000 characters long")
		}
	}

	if r.UserVisibleDataFormat != "" && r.UserVisibleDataFormat != "simpleMarkdownV1" {
		return errors.New("invalid userVisibleDataFormat")
	}

	return nil
}

func (r *AuthSignRequest) MarshalJSON() ([]byte, error) {
	type alias AuthSignRequest // Needed to avoid recursion
	asr := alias(*r)

	if asr.UserVisibleData != "" {
		asr.UserVisibleData = base64.StdEncoding.EncodeToString([]byte(asr.UserVisibleData))
	}
	if asr.UserNonVisibleData != "" {
		asr.UserNonVisibleData = base64.StdEncoding.EncodeToString([]byte(asr.UserNonVisibleData))
	}
	return json.Marshal(asr)
}

type AuthSignResponse struct {
	OrderRef       string `json:"orderRef"`
	AutoStartToken string `json:"autoStartToken"`
	QrStartToken   string `json:"qrStartToken"`
	QrStartSecret  string `json:"qrStartSecret"`
}

// BuildQrCode builds a BankID compatible QR code
// See https://www.bankid.com/utvecklare/guider/teknisk-integrationsguide/qrkoder
func (r *AuthSignResponse) BuildQrCode(time int) string {
	sb := strings.Builder{}
	sb.WriteString("bankid.")
	sb.WriteString(r.QrStartToken)
	sb.WriteString(".")
	sb.WriteString(strconv.Itoa(time))
	sb.WriteString(".")

	mac := hmac.New(sha256.New, []byte(r.QrStartSecret))
	mac.Write([]byte(strconv.Itoa(time)))
	qrAuthCode := mac.Sum(nil)

	sb.WriteString(hex.EncodeToString(qrAuthCode))

	return sb.String()
}

type Requirement struct {
	PinCode             bool     `json:"pinCode,omitempty"`
	MRTD                bool     `json:"mrtd,omitempty"`
	CardReader          string   `json:"cardReader,omitempty"`
	CertificatePolicies []string `json:"certificatePolicies,omitempty"`
	PersonalNumber      string   `json:"personalNumber,omitempty"`
}

type CollectRequest struct {
	OrderRef string `json:"orderRef"`
}

func (c *CollectRequest) Validate() error {
	if c.OrderRef == "" {
		return errors.New("missing order ref")
	}

	return nil
}

type CollectResponse struct {
	OrderRef       string         `json:"orderRef"`
	Status         Status         `json:"status"`
	HintCode       HintCode       `json:"hintCode"`
	CompletionData CompletionData `json:"completionData"`
}

type Status string

const (
	Pending  Status = "pending"
	Complete Status = "complete"
	Failed   Status = "failed"
)

type HintCode string

const (
	OutstandingTransaction HintCode = "outstandingTransaction"
	NoClient               HintCode = "noClient"
	Started                HintCode = "started"
	UserMrtd               HintCode = "userMrtd"
	UserCallConfirm        HintCode = "userCallConfirm"
	UserSign               HintCode = "userSign"
)

type CompletionData struct {
	User            User   `json:"user"`
	Device          Device `json:"device"`
	BankIdIssueDate string `json:"bankIdIssueDate"`
	StepUp          StepUp `json:"stepUp"`
	Signature       string `json:"signature"`
	OcspResponse    string `json:"ocspResponse"`
}
type User struct {
	PersonalNumber string `json:"personalNumber"`
	Name           string `json:"name"`
	GivenName      string `json:"givenName"`
	SurName        string `json:"surName"`
}
type Device struct {
	IpAddress string `json:"ipAddress"`
	UHI       string `json:"uhi"`
}
type StepUp struct {
	MRTD bool `json:"mrtd"`
}

type WatchResponse struct {
	Cancelled bool
	Status    string
}

type CancelRequest struct {
	OrderRef string `json:"orderRef"`
}

func (c *CancelRequest) Validate() error {
	if c.OrderRef == "" {
		return errors.New("missing order ref")
	}

	return nil
}

type BankIdError struct {
	ErrorCode string `json:"errorCode"`
	Details   string `json:"details"`
}

func (e BankIdError) Error() string {
	return fmt.Sprintf("bankid: %s, %s", e.ErrorCode, e.Details)
}

type Empty struct{}
