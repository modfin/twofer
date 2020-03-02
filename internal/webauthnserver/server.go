package webauthnserver

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"hash"
	"twofer/twoferrpc"
)

func New(macKey []byte) *Server {
	//webAuthnConfig := config.Get().WebAuthn
	webAuthn, _ := webauthn.New(&webauthn.Config{
		RPDisplayName: "localhost",              //webAuthnConfig.RPDisplayName,
		RPID:          "localhost",              //webAuthnConfig.RPID,
		RPOrigin:      "http://localhost:63343", //webAuthnConfig.RPOrigin,
	})
	s := &Server{
		webAuthn: webAuthn,
		macer: func() hash.Hash {
			return hmac.New(sha256.New, macKey)
		},
	}
	return s
}

type Server struct {
	webAuthn *webauthn.WebAuthn
	macer    func() hash.Hash
}

func (s Server) BeginRegister(_ context.Context, req *twoferrpc.BeginRegisterRequest) (res *twoferrpc.BeginRegisterResponse, err error) {
	allowedCredentials := twoferCredentialsToWebAuthnCredentials(req.User.AllowedCredentials)
	u := User{
		Id:          req.User.Id,
		Name:        req.User.Name,
		DisplayName: req.User.DisplayName,
		Credentials: allowedCredentials,
	}

	options, sessiondata, err := s.webAuthn.BeginRegistration(u)
	if err != nil {
		return
	}

	pkBlob, err := json.Marshal(options.Response)
	twoferSession := twoferrpc.SessionData{
		Challenge: sessiondata.Challenge,
		UserId:    sessiondata.UserID,
	}

	sign := createSignature(&twoferSession, s.macer)
	twoferSession.Signature = sign
	response := &twoferrpc.BeginRegisterResponse{
		PublicKey:   string(pkBlob),
		SessionData: &twoferSession,
		User:        req.User,
	}

	return response, nil
}

func (s Server) FinishRegister(_ context.Context, req *twoferrpc.FinishRegisterRequest) (res *twoferrpc.FinishRegisterResponse, err error) {
	signature := createSignature(req.SessionData, s.macer)
	if signature != req.SessionData.Signature {
		return nil, errors.New("the signatures do not match")
	}

	message := json.RawMessage(req.Blob)
	body, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(message))
	if err != nil {
		return nil, errors.New("failed to parse authenticator response")
	}

	sessionData := webauthn.SessionData{
		Challenge:            req.SessionData.Challenge,
		UserID:               req.SessionData.UserId,
		AllowedCredentialIDs: req.SessionData.AllowedCredentials,
		UserVerification:     protocol.UserVerificationRequirement(req.SessionData.UserVerification),
	}

	allowedCredentials := twoferCredentialsToWebAuthnCredentials(req.User.AllowedCredentials)

	u := User{
		Id:          req.User.Id,
		Name:        req.User.Name,
		DisplayName: req.User.DisplayName,
		Credentials: allowedCredentials,
	}
	credential, err := s.webAuthn.CreateCredential(u, sessionData, body)
	if err != nil {
		fmt.Printf("Something's bonkers\n err=%+v", err)
		return nil, err
	}
	var exists bool
	for _, ac := range req.User.AllowedCredentials {
		exists = exists || bytes.Equal(ac.Authenticator.AAGUID, credential.Authenticator.AAGUID)
		if exists {
			break
		}
	}
	if !exists {
		req.User.AllowedCredentials = append(req.User.AllowedCredentials, &twoferrpc.AllowedCredential{
			ID:              credential.ID,
			PublicKey:       credential.PublicKey,
			AttestationType: credential.AttestationType,
			Authenticator: &twoferrpc.Authenticator{
				AAGUID:       credential.Authenticator.AAGUID,
				SignCount:    credential.Authenticator.SignCount,
				CloneWarning: credential.Authenticator.CloneWarning,
			},
		})
	}
	toReturn := twoferrpc.FinishRegisterResponse{
		User: req.User,
	}
	return &toReturn, nil
}
func (s Server) BeginLogin(_ context.Context, req *twoferrpc.BeginLoginRequest) (res *twoferrpc.BeginLoginResponse, err error) {
	allowedCredentials := twoferCredentialsToWebAuthnCredentials(req.User.AllowedCredentials)

	u := User{
		Id:          req.User.Id,
		Name:        req.User.Name,
		DisplayName: req.User.DisplayName,
		Credentials: allowedCredentials,
	}
	credentialAssertion, sessionData, err := s.webAuthn.BeginLogin(u)
	if err != nil {
		return
	}

	credentialAssertionOptions, err := json.Marshal(credentialAssertion.Response)
	if err != nil {
		fmt.Printf("%+v", credentialAssertionOptions)
	}
	twoferSession := &twoferrpc.SessionData{
		Challenge:          sessionData.Challenge,
		UserId:             sessionData.UserID,
		AllowedCredentials: sessionData.AllowedCredentialIDs,
	}
	signature := createSignature(twoferSession, s.macer)
	twoferSession.Signature = signature

	loginResponse := twoferrpc.BeginLoginResponse{
		PublicKey:   string(credentialAssertionOptions),
		SessionData: twoferSession,
		User:        req.User,
	}
	return &loginResponse, nil
}
func (s Server) FinishLogin(_ context.Context, req *twoferrpc.FinishLoginRequest) (res *twoferrpc.FinishLoginResponse, err error) {
	signature := createSignature(req.Session, s.macer)
	if signature != req.Session.Signature {
		return nil, errors.New("signatures do not match")
	}
	rawMessage := json.RawMessage(req.Blob)
	body, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(rawMessage))
	if err != nil {
		return nil, err
	}
	sessionData := webauthn.SessionData{
		Challenge:            req.Session.Challenge,
		UserID:               req.Session.UserId,
		AllowedCredentialIDs: req.Session.AllowedCredentials,
		UserVerification:     protocol.UserVerificationRequirement(req.Session.UserVerification),
	}

	allowedCredentials := twoferCredentialsToWebAuthnCredentials(req.User.AllowedCredentials)

	u := User{
		Id:          req.User.Id,
		Name:        req.User.Name,
		DisplayName: req.User.DisplayName,
		Credentials: allowedCredentials,
	}

	_, err = s.webAuthn.ValidateLogin(u, sessionData, body)
	if err != nil {
		fmt.Printf("Something's bonkers")
		return nil, err
	}
	response := twoferrpc.FinishLoginResponse{
		User: req.User,
	}
	return &response, nil
}

type User struct {
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"displayName"`
	Credentials []webauthn.Credential `json:"credentials, omitempty"`
}

func createSignature(session *twoferrpc.SessionData, macer func() hash.Hash) string {
	mac := macer()
	toSign := session.Challenge + string(session.UserId)
	mac.Write([]byte(toSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return signature
}

func twoferCredentialsToWebAuthnCredentials(ac []*twoferrpc.AllowedCredential) []webauthn.Credential {
	var allowedCredentials []webauthn.Credential
	for _, c := range ac {
		allowedCredentials = append(allowedCredentials, webauthn.Credential{
			ID:              c.ID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       c.Authenticator.AAGUID,
				SignCount:    c.Authenticator.SignCount,
				CloneWarning: c.Authenticator.CloneWarning,
			},
		})
	}
	return allowedCredentials
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
