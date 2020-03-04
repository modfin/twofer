package webauthnserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"twofer/twoferrpc/gw6n"
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

func (s Server) BeginRegister(_ context.Context, req *gw6n.BeginRegisterRequest) (res *gw6n.BeginRegisterResponse, err error) {

	var u User
	if len(req.UserBlob) > 0 {
		err = json.Unmarshal(req.UserBlob, &u)
		if err != nil {
			return
		}
	}

	if req.User != nil {
		if u.Id == "" {
			u.Id = req.User.Id
		}
		if u.Name == "" {
			u.Name = req.User.Name
		}
	}

	if u.Id == "" {
		return nil, errors.New("an user id must be provided for registration")
	}

	if u.Name == "" {
		u.Name = u.Id
	}
	if u.DisplayName == "" {
		u.DisplayName = u.Name

	}

	credentialCreation, sessionData, err := s.webAuthn.BeginRegistration(u)
	if err != nil {
		return
	}

	data, err := json.Marshal(credentialCreation)
	if err != nil {
		return
	}

	session, err := json.Marshal(Session{
		Data: sessionData,
		User: User{
			Id:          u.Id,
			Name:        u.Name,
			DisplayName: u.DisplayName,
		},
	})

	response := &gw6n.BeginRegisterResponse{
		Session: session,
		Json:    data,
	}
	return response, nil
}

func (s Server) FinishRegister(_ context.Context, req *gw6n.FinishRegisterRequest) (res *gw6n.FinishRegisterResponse, err error) {

	credentialCreation, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.Signature))
	if err != nil {
		return nil, errors.New("failed to parse authenticator response")
	}

	var session Session
	err = json.Unmarshal(req.Session, &session)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, errors.New("session data was not provided")
	}

	var u User
	if len(req.UserBlob) > 0 {
		err = json.Unmarshal(req.UserBlob, &u)
		if err != nil {
			return
		}
	}
	if u.Id == "" {
		u.Id = session.User.Id
	}
	if u.Name == "" {
		u.Name = session.User.Name
	}
	if u.DisplayName == "" {
		u.DisplayName = session.User.DisplayName
	}

	credential, err := s.webAuthn.CreateCredential(u, *session.Data, credentialCreation)
	if err != nil {
		fmt.Printf("Something's bonkers err=%+v\n", err)
		return nil, err
	}

	u.Credentials = append(u.Credentials, *credential)

	res = &gw6n.FinishRegisterResponse{}
	res.UserBlob, err = json.Marshal(u)
	return res, err
}
func (s Server) BeginLogin(_ context.Context, req *gw6n.BeginLoginRequest) (res *gw6n.BeginLoginResponse, err error) {
	var u User
	err = json.Unmarshal(req.UserBlob, &u)
	if err != nil {
		return
	}

	credentialAssertion, sessionData, err := s.webAuthn.BeginLogin(u, webauthn.WithUserVerification(protocol.VerificationDiscouraged))
	if err != nil {
		return
	}

	session := Session{
		Data: sessionData,
	}

	ss, err := json.Marshal(session)
	if err != nil {
		return
	}

	response, err := json.Marshal(credentialAssertion.Response)
	if err != nil {
		return
	}

	return &gw6n.BeginLoginResponse{
		Session: ss,
		Json:    response,
	}, nil
}
func (s Server) FinishLogin(_ context.Context, req *gw6n.FinishLoginRequest) (res *gw6n.FinishLoginResponse, err error) {

	body, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Signature))
	var u User

	err = json.Unmarshal(req.UserBlob, &u)
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal(req.Session, &session)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, errors.New("session data must be provided")

	}

	_, err = s.webAuthn.ValidateLogin(u, *session.Data, body)
	if err != nil {
		return nil, err
	}

	return &gw6n.FinishLoginResponse{}, nil

}
