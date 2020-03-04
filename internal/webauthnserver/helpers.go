package webauthnserver

import (
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
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
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"display_name"`
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

type Session struct {
	Data *webauthn.SessionData `json:"data"`
	User User                  `json:"user"`
}
