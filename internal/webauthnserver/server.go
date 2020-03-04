package webauthnserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"twofer/internal/config"
	"twofer/twoferrpc/gw6n"
)

func New(config config.WebAuthn) (*Server, error) {
	s := &Server{
		defaultConfig: &webauthn.Config{
			RPDisplayName: config.RPDisplayName,
			RPID:          config.RPID,
			RPOrigin:      config.RPOrigin,
			RPIcon:        "",
			AuthenticatorSelection: protocol.AuthenticatorSelection{
				UserVerification: toUserVerification(config.UserVerification),
			},
			Timeout: 0,
			Debug:   false,
		},
	}
	_, err := webauthn.New(s.defaultConfig)
	if err != nil {
		return nil, err
	}

	return s, err
}

type Server struct {
	defaultConfig *webauthn.Config
}

func (s *Server) create(config *webauthn.Config) (*webauthn.WebAuthn, error) {
	if config == nil {
		return webauthn.New(s.defaultConfig)
	}
	return webauthn.New(config)
}

func (s *Server) BeginRegister(_ context.Context, req *gw6n.BeginRegisterRequest) (res *gw6n.BeginRegisterResponse, err error) {

	service, err := s.create(nil)
	if err != nil {
		return nil, err
	}

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

	credentialCreation, sessionData, err := service.BeginRegistration(u)
	if err != nil {
		return
	}

	data, err := json.Marshal(credentialCreation)
	if err != nil {
		return
	}

	session, err := json.Marshal(Session{
		Data: sessionData,
		User: u,
	})

	response := &gw6n.BeginRegisterResponse{
		Session: session,
		Json:    data,
	}
	return response, nil
}

func (s *Server) FinishRegister(_ context.Context, req *gw6n.FinishRegisterRequest) (res *gw6n.FinishRegisterResponse, err error) {

	service, err := s.create(nil)
	if err != nil {
		return nil, err
	}

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

	credential, err := service.CreateCredential(session.User, *session.Data, credentialCreation)
	if err != nil {
		fmt.Printf("Something's bonkers err=%+v\n", err)
		return nil, err
	}

	session.User.Credentials = append(session.User.Credentials, *credential)

	res = &gw6n.FinishRegisterResponse{}
	res.UserBlob, err = json.Marshal(session.User)
	return res, err
}
func (s *Server) BeginLogin(_ context.Context, req *gw6n.BeginLoginRequest) (res *gw6n.BeginLoginResponse, err error) {

	service, err := s.create(nil)
	if err != nil {
		return nil, err
	}

	var u User
	err = json.Unmarshal(req.UserBlob, &u)
	if err != nil {
		return
	}

	credentialAssertion, sessionData, err := service.BeginLogin(u)
	if err != nil {
		return
	}

	session, err := json.Marshal(Session{
		User: u,
		Data: sessionData,
	})
	if err != nil {
		return
	}

	response, err := json.Marshal(credentialAssertion.Response)
	if err != nil {
		return
	}

	return &gw6n.BeginLoginResponse{
		Session: session,
		Json:    response,
	}, nil
}
func (s *Server) FinishLogin(_ context.Context, req *gw6n.FinishLoginRequest) (res *gw6n.FinishLoginResponse, err error) {

	service, err := s.create(nil)
	if err != nil {
		return nil, err
	}

	body, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Signature))

	var session Session
	err = json.Unmarshal(req.Session, &session)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, errors.New("session data must be provided")

	}

	_, err = service.ValidateLogin(session.User, *session.Data, body)
	if err != nil {
		return nil, err
	}

	return &gw6n.FinishLoginResponse{}, nil

}
