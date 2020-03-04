package webauthnserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"twofer/twoferrpc"
)

func New() (*Server, error) {
	//webAuthnConfig := config.Get().WebAuthn
	webAuthn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "localhost",             //webAuthnConfig.RPDisplayName,
		RPID:          "localhost",             //webAuthnConfig.RPID,
		RPOrigin:      "http://localhost:8080", //webAuthnConfig.RPOrigin,
	})
	s := &Server{
		webAuthn: webAuthn,
	}
	return s, err
}

type Server struct {
	webAuthn *webauthn.WebAuthn
}

func (s Server) BeginRegister(_ context.Context, req *twoferrpc.BeginRegisterRequest) (res *twoferrpc.BeginRegisterResponse, err error) {

	u := User{
		Id:          req.UserId,
		Name:        "Test",
		DisplayName: "Test",
	}

	if len(req.UserBlob) > 0 {
		err = json.Unmarshal(req.UserBlob, &u.Credentials)
		if err != nil {
			return
		}
	}

	credentialCreation, sessiondata, err := s.webAuthn.BeginRegistration(u)
	if err != nil {
		return
	}

	resp, err := json.Marshal(credentialCreation)
	if err != nil {
		return
	}
	session, err := json.Marshal(sessiondata)

	response := &twoferrpc.BeginRegisterResponse{
		Session:       session,
		Response2User: resp,
	}
	return response, nil
}

func (s Server) FinishRegister(_ context.Context, req *twoferrpc.FinishRegisterRequest) (res *twoferrpc.FinishRegisterResponse, err error) {

	credentialCreation, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.UserSignature))
	if err != nil {
		return nil, errors.New("failed to parse authenticator response")
	}

	var session webauthn.SessionData
	err = json.Unmarshal(req.Session, &session)
	if err != nil {
		return nil, err
	}

	u := User{
		Id:          string(session.UserID),
		Name:        "Test",
		DisplayName: "Test",
		Credentials: []webauthn.Credential{},
	}

	credential, err := s.webAuthn.CreateCredential(u, session, credentialCreation)
	if err != nil {
		fmt.Printf("Something's bonkers err=%+v\n", err)
		return nil, err
	}

	var credentials []*webauthn.Credential

	if req.UserBlob != nil {
		err = json.Unmarshal(req.UserBlob, &credentials)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}
	credentials = append(credentials, credential)

	res = &twoferrpc.FinishRegisterResponse{}
	res.UserBlob, err = json.Marshal(credentials)
	return res, err
}
func (s Server) BeginLogin(_ context.Context, req *twoferrpc.BeginLoginRequest) (res *twoferrpc.BeginLoginResponse, err error) {
	u := User{
		Id:          "test@example.com",
		Name:        "Test",
		DisplayName: "Test",
	}

	err = json.Unmarshal(req.UserBlob, &u.Credentials)
	if err != nil {
		return
	}

	credentialAssertion, sessionData, err := s.webAuthn.BeginLogin(u, webauthn.WithUserVerification(protocol.VerificationDiscouraged))
	if err != nil {
		return
	}

	session, err := json.Marshal(sessionData)

	response, err := json.Marshal(credentialAssertion.Response)
	if err != nil {
		return
	}

	return &twoferrpc.BeginLoginResponse{
		Session:       session,
		Response2User: response,
	}, nil
}
func (s Server) FinishLogin(_ context.Context, req *twoferrpc.FinishLoginRequest) (res *twoferrpc.FinishLoginResponse, err error) {

	body, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.UserSignature))
	u := User{
		Id:          "test@example.com",
		Name:        "Test",
		DisplayName: "Test",
	}

	err = json.Unmarshal(req.UserBlob, &u.Credentials)
	if err != nil {
		return nil, err
	}

	var session webauthn.SessionData
	err = json.Unmarshal(req.Session, &session)
	if err != nil {
		return nil, err
	}

	_, err = s.webAuthn.ValidateLogin(u, session, body)
	if err != nil {
		return nil, err
	}

	return &twoferrpc.FinishLoginResponse{}, nil

}

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
