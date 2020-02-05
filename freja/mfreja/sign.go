package mfreja

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)


type SignatureType string
const (
	ST_SIMPLE SignatureType = "SIMPLE"
	ST_EXTENDED SignatureType = "EXTENDED"
)

type SignRef struct {
	SignRef string `json:"signRef"` // mandatory. The authentication reference of the authentication.
}

type PushNotification struct {
	Title string `json:"title"`
	Text string `json:"text"`
}


type DataToSign struct {

	//string, mandatory, 4096 plain text characters maximum prior to Base64 encoding. The text that will be shown in
	// the mobile application and signed by the end user. The content of the Base64 string are bytes representing a UTF-8
	// encoding of the text to be displayed to and signed by the user.
	Text string `json:"text"`

	// binaryData: string, mandatory, 5 MB maximum. This is not shown to the user in the mobile application but is,
	// nonetheless included in the signature.
	BinaryData []byte `json:"binaryData"`
}

func (a DataToSign) MarshalJSON() ([]byte, error) {

	var res []byte
	res = append(res, []byte(`{"text":`)...)
	d, err := json.Marshal(base64.StdEncoding.EncodeToString([]byte(a.Text)))
	if err != nil{
		return nil, err
	}
	res = append(res, d...)

	if len(a.BinaryData) > 0{
		data := base64.StdEncoding.EncodeToString(a.BinaryData)
		res = append(res, []byte(`,"binaryData":"`)...)
		res = append(res, data...)
		res = append(res, []byte(`"`)...)
	}

	return append(res, []byte(`}`)...), nil
}


type SignRequest struct {
	UserInfoType UserInfoType `json:"userInfoType"`



	// 256 characters maximum
	//EMAIL or PHONE, interpreted as a string value
	//SSN, then it must be a Base64 encoding of the ssnuserinfo JSON structure described
	//PHONE MUST be in the form of: "+4673*******"; the leading plus '+' is present whereas the leading zero from the mobile phone operator code '0' is not.
	//INFERRED userInfo N/A is used when the user is being authenticated by scanning a QR code
	UserInfo string `json:"userInfo"`

	// optional.
	// Minimum required registration level of a user in order to approve/decline authentication. Can be BASIC, EXTENDED or PLUS. If not present, default level will be BASIC.
	// default level will be BASIC.
	MinRegistrationLevel RegistrationLevel `json:"minRegistrationLevel"`

	// optional
	// 128 characters maximum. The title to display in the transaction list if presented to the user on the mobile device.
	// The title will be presented regardless of the confidentiality setting (see below). If not present, a system default
	// text will be presented.
	Title string `json:"title"`

	// optional.
	// The title and the text of the notification sent to the mobile device to alert the user of a signature request.
	// The character limit for the push notification title and text is 256 characters for each. If not present,
	// a system default title and text will be presented.
	PushNotification *PushNotification `json:"pushNotification"`

	// optional. Describes the time until which the Relying Party is ready to wait for the user to confirm the signature
	// request. Expressed in milliseconds since January 1, 1970, 00:00 UTC. Min value is current time +2 minutes,
	// max value is current time +30 days. If not present, defaults to current time +2 minutes.
	Expiry int64 `json:"expiry,omitempty"`

	// Dont set this. The marshaling figure it out from the Data you want to sign
	SignatureType SignatureType `json:"signatureType"`
	// Dont set this. The marshaling figure it out from the Data you want to sign
	DataToSignType string `json:"dataToSignType"`

	// mandatory
	DataToSign DataToSign `json:"dataToSign"`

	//optional
	// If the requested attribute is BASIC_USER_INFO, DATE_OF_BIRTH or SSN the minRegistrationLevel must be set to EXTENDED or PLUS.
	AttributesToReturn []Attribute `json:"attributesToReturn"`

}

func (i *SignRequest) SetExpiry(t time.Time){
	i.Expiry = t.UnixNano() / int64(time.Millisecond)
}

func (i *SignRequest) UseEmail(email string) {
	i.UserInfo = email
	i.UserInfoType = UIT_EMAIL
}
func (i *SignRequest) UsePhone(phone string) {
	i.UserInfo = phone
	i.UserInfoType = UIT_PHONE
}
func (i *SignRequest) UseSSN(country string, ssn string) error {
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

func (i SignRequest) Marshal() (s string, err error) {
	if i.UserInfoType == UIT_SSN {
		i.UserInfo = base64.StdEncoding.EncodeToString([]byte(i.UserInfo))
	}
	if i.UserInfoType == UIT_INFERRED {
		return "", errors.New("inferred User Info Type is not allowed for signing")
	}

	if i.SignatureType == "" {
		i.SignatureType = ST_SIMPLE
	}
	i.DataToSignType = "SIMPLE_UTF8_TEXT"
	if i.SignatureType == ST_EXTENDED || len(i.DataToSign.BinaryData) > 0 {
		i.SignatureType = ST_EXTENDED
		i.DataToSignType = "EXTENDED_UTF8_TEXT"
	}


	b, err := json.Marshal(i)
	if err != nil {
		return
	}

	fmt.Println(string(b))

	return fmt.Sprintf("initSignRequest=%s", base64.StdEncoding.EncodeToString(b)), nil
}


type SignResponse struct {
	SignRef
	Status              Status              `json:"status"`

	RequestedAttributes RequestedAttributes `json:"requestedAttributes"` // optional
	Details             string              `json:"details"`             // JWS signed data, see below
}

func (a SignResponse) JWSToken() string{
	return a.Details
}
