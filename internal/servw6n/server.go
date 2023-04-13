package servw6n

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/modfin/twofer/grpc/gw6n"
	"github.com/modfin/twofer/internal/config"
	"github.com/modfin/twofer/internal/ratelimit"
	"time"
)

func New(config config.WebAuthn) (*Server, error) {
	s := &Server{
		ratelimiter: ratelimit.New(config.RateLimit),
		hmacKey:     []byte(config.HMACKey),
		timeout:     config.Timeout,
		defaultConfig: &webauthn.Config{
			RPDisplayName: config.RPDisplayName,
			RPID:          config.RPID,
			RPOrigin:      config.RPOrigin,
			RPIcon:        "",
			AuthenticatorSelection: protocol.AuthenticatorSelection{
				UserVerification: toUserVerification(config.UserVerification),
			},
			Timeout: int(config.Timeout.Milliseconds()),
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
	ratelimiter   *ratelimit.Ratelimiter
	defaultConfig *webauthn.Config
	hmacKey       []byte
	timeout       time.Duration
}

func (s *Server) create(c interface{ GetCfg() *gw6n.Config }) (*webauthn.WebAuthn, error) {
	if c == nil {
		return webauthn.New(s.defaultConfig)
	}
	cfg := c.GetCfg()
	if cfg == nil {
		return webauthn.New(s.defaultConfig)
	}

	return webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigin:      cfg.RPOrigin,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: toUserVerification(cfg.UserVerification),
		},
	})
}

func (s *Server) EnrollInit(_ context.Context, req *gw6n.EnrollInitReq) (res *gw6n.InitRes, err error) {

	service, err := s.create(req)
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

	err = s.ratelimiter.Hit(u.Id)
	if err != nil {
		return nil, err
	}

	credentialCreation, sessionData, err := service.BeginRegistration(u)
	if err != nil {
		return
	}

	data, err := json.Marshal(credentialCreation)
	if err != nil {
		return
	}

	session, err := Session{
		Deadline: time.Now().Add(s.timeout),
		Data:     sessionData,
		User:     u,
	}.Marshal(s.hmacKey)

	response := &gw6n.InitRes{
		Session: session,
		Json:    data,
	}
	return response, nil
}

func (s *Server) EnrollFinal(_ context.Context, req *gw6n.FinalReq) (res *gw6n.FinalRes, err error) {

	service, err := s.create(req)
	if err != nil {
		return nil, err
	}
	credentialCreation, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.Signature))
	if err != nil {
		return nil, errors.New("failed to parse authenticator response")
	}

	var session Session
	err = session.Unmarshal(s.hmacKey, req.Session)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, errors.New("session data was not provided")
	}

	err = s.ratelimiter.Hit(session.User.Id)
	if err != nil {
		return nil, err
	}

	credential, err := service.CreateCredential(session.User, *session.Data, credentialCreation)
	if err != nil {
		fmt.Printf("Something's bonkers err=%+v\n", err)
		return nil, err
	}

	session.User.Credentials = append(session.User.Credentials, Credential{Credential: *credential, RPID: service.Config.RPID})

	res = &gw6n.FinalRes{}
	res.UserBlob, err = json.Marshal(session.User)
	return res, err
}
func (s *Server) AuthInit(_ context.Context, req *gw6n.AuthInitReq) (res *gw6n.InitRes, err error) {

	service, err := s.create(req)
	if err != nil {
		return nil, err
	}

	var u User
	err = json.Unmarshal(req.UserBlob, &u)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.ratelimiter.Hit(u.Id)
	if err != nil {
		return nil, err
	}

	// Filtering out credentials valid for this RP
	var credentials []Credential
	for _, c := range u.Credentials {
		if c.RPID == service.Config.RPID {
			credentials = append(credentials, c)
		}
	}
	u.Credentials = credentials

	credentialAssertion, sessionData, err := service.BeginLogin(u)
	if err != nil {
		return
	}

	session, err := Session{
		Deadline: time.Now().Add(s.timeout),
		User:     u,
		Data:     sessionData,
	}.Marshal(s.hmacKey)
	if err != nil {
		return
	}

	response, err := json.Marshal(credentialAssertion.Response)
	if err != nil {
		return
	}

	return &gw6n.InitRes{
		Session: session,
		Json:    response,
	}, nil
}
func (s *Server) AuthFinal(_ context.Context, req *gw6n.FinalReq) (res *gw6n.FinalRes, err error) {

	service, err := s.create(req)
	if err != nil {
		return nil, err
	}

	body, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Signature))

	var session Session
	err = session.Unmarshal(s.hmacKey, req.Session)
	if err != nil {
		return nil, err
	}

	err = s.ratelimiter.Hit(session.User.Id)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, errors.New("session data must be provided")

	}
	_, err = service.ValidateLogin(session.User, *session.Data, body)
	if err != nil {
		return &gw6n.FinalRes{
			Valid: false,
		}, err
	}

	return &gw6n.FinalRes{
		Valid: true,
	}, nil

}
