package frejam

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)



type Attribute struct {
	Attribute AttributeName `json:"attribute"`
}

type AuthRequest struct {
	UserInfoType UserInfoType `json:"userInfoType"`

	// 256 characters maximum
	//EMAIL or PHONE, interpreted as a string value
	//SSN, then it must be a Base64 encoding of the ssnuserinfo JSON structure described
	//PHONE MUST be in the form of: "+4673*******"; the leading plus '+' is present whereas the leading zero from the mobile phone operator code '0' is not.
	//INFERRED userInfo N/A is used when the user is being authenticated by scanning a QR code
	UserInfo string `json:"userInfo"`

	//optional
	// If the requested attribute is BASIC_USER_INFO, DATE_OF_BIRTH or SSN the minRegistrationLevel must be set to EXTENDED or PLUS.
	AttributesToReturn []Attribute `json:"attributesToReturn"`

	// optional.
	// Minimum required registration level of a user in order to approve/decline authentication. Can be BASIC, EXTENDED or PLUS. If not present, default level will be BASIC.
	// default level will be BASIC.
	MinRegistrationLevel RegistrationLevel `json:"minRegistrationLevel"`
}

func (i *AuthRequest) UseInferred() {
	i.UserInfo = "N/A"
	i.UserInfoType = UIT_INFERRED
}
func (i *AuthRequest) UseEmail(email string) {
	i.UserInfo = email
	i.UserInfoType = UIT_EMAIL
}
func (i *AuthRequest) UsePhone(phone string) {
	i.UserInfo = phone
	i.UserInfoType = UIT_PHONE
}
func (i *AuthRequest) UseSSN(country string, ssn string) error {
	b, err := json.Marshal(SSN{
		Country: strings.ToUpper(country),
		SSN:     ssn,
	})
	if err != nil {
		return err
	}
	i.UserInfo = string(b)
	i.UserInfoType = UIT_SSN
	return nil
}

func (i AuthRequest) Marshal() (s string, err error) {
	if i.UserInfoType == UIT_SSN {
		i.UserInfo = base64.StdEncoding.EncodeToString([]byte(i.UserInfo))
	}

	if i.UserInfoType == UIT_INFERRED {
		i.UserInfo = "N/A"
	}

	b, err := json.Marshal(i)
	if err != nil {
		return
	}

	return fmt.Sprintf("initAuthRequest=%s", base64.StdEncoding.EncodeToString(b)), nil
}


type AuthRef struct {
	AuthRef string `json:"authRef"` // mandatory. The authentication reference of the authentication.
}


type AuthResponse struct {
	AuthRef
	Status              Status              `json:"status"`
	RequestedAttributes RequestedAttributes `json:"requestedAttributes"` // optional
	Details             string              `json:"details"`             // JWS signed data, see below
}

func (a AuthResponse) JWSToken() string{
	return a.Details
}

type RequestedAttributes struct {
	BasicUserInfo            BasicUserInfo `json:"basicUserInfo,omitempty"`
	EmailAddress             string        `json:"emailAddress,omitempty"`
	DateOfBirth              string        `json:"dateOfBirth,omitempty"`
	CustomIdentifier         string        `json:"customIdentifier,omitempty"`
	SSN                      SSN           `json:"ssn,omitempty"`
	RelyingPartyUserID       string        `json:"relyingPartyUserId,omitempty"`
	IntegratorSpecificUserID string        `json:"integratorSpecificUserId,omitempty"`
}
