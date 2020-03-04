package webauthnserver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"time"
)

func toUserVerification(mode string) protocol.UserVerificationRequirement {
	switch mode {
	case "required":
		return protocol.VerificationRequired
	case "preferred":
		return protocol.VerificationPreferred
	case "discouraged":
		fallthrough
	default:
		return protocol.VerificationDiscouraged
	}

}

type User struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	Credentials []Credential `json:"credentials, omitempty"`
}

type Credential struct {
	webauthn.Credential
	RPID string
}

func (u User) WebAuthnID() []byte {
	return []byte(u.Id)
}

// User Name according to the Relying Party
func (u User) WebAuthnName() string {
	return u.Name
}

// Display Name of the user
func (u User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// User's icon url
func (u User) WebAuthnIcon() string {
	return ""
}

// Credentials owned by the user
func (u User) WebAuthnCredentials() []webauthn.Credential {
	var cred []webauthn.Credential
	for _, c := range u.Credentials {
		cred = append(cred, c.Credential)
	}
	return cred
}

type Session struct {
	Deadline time.Time             `json:"deadline"`
	Data     *webauthn.SessionData `json:"data"`
	User     User                  `json:"user"`
}

func (s Session) Marshal(key []byte) ([]byte, error) {
	j, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, key)
	_, err = mac.Write(j)
	if err != nil {
		return nil, err
	}
	sig := mac.Sum(nil)
	return []byte(base64.StdEncoding.EncodeToString(j) + "." + base64.StdEncoding.EncodeToString(sig)), nil
}

func (s *Session) Unmarshal(key []byte, token []byte) error {

	parts := bytes.Split(token, []byte("."))
	if len(parts) != 2 {
		return errors.New("not a correct formatted token")
	}

	payload := parts[0]
	sig1, err := base64.StdEncoding.DecodeString(string(parts[1]))
	if err != nil {
		return err
	}
	j, err := base64.StdEncoding.DecodeString(string(payload))

	mac := hmac.New(sha256.New, key)
	_, err = mac.Write(j)
	if err != nil {
		return err
	}
	sig2 := mac.Sum(nil)
	if !hmac.Equal(sig1, sig2) {
		return errors.New("signature does not match content")
	}

	return json.Unmarshal(j, &s)
}
