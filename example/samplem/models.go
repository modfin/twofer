package samplem

import (
	"encoding/json"
	"github.com/duo-labs/webauthn/webauthn"
)

type User struct {
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"displayName"`
	Credentials []webauthn.Credential `json:"credentials, omitempty"`
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
	return u.Credentials
}

type RegisterResponse struct {
	PublicKey json.RawMessage `json:"publicKey"`
}

type LoginResponse struct {
	PublicKey json.RawMessage `json:"publicKey"`
}
