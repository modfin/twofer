package freja

import (
	"context"
	"crypto/rsa"
	"fmt"
	"golang.org/x/oauth2/jws"
	"io/ioutil"
	"net/http"
	"time"
	"twofer/eid"
	"twofer/eid/freja/frejam"
	"twofer/internal/mtls"
)

type Client struct {
	baseURL      string
	pollInterval time.Duration

	client *http.Client
	api    *API

	pemRootCA     []byte
	pemClientCert []byte
	pemClientKey  []byte

	jwsPubKey *rsa.PublicKey

	defaultRegistrationLevel frejam.RegistrationLevel

	timeout time.Duration
}

type ClientConfig struct {
	BaseURL string

	PemRootCA     []byte
	PemClientCert []byte
	PemClientKey  []byte

	// If present the JWS tokens are validated, otherwise everything is let through
	PemJWSCert []byte

	Timeout      time.Duration
	PollInterval time.Duration

	DefaultRegistrationLevel frejam.RegistrationLevel
}

func New(config ClientConfig) (client *Client, err error) {
	client = &Client{
		baseURL:                  config.BaseURL,
		pemClientKey:             config.PemClientKey,
		pemClientCert:            config.PemClientCert,
		pemRootCA:                config.PemRootCA,
		timeout:                  config.Timeout,
		pollInterval:             config.PollInterval,
		defaultRegistrationLevel: config.DefaultRegistrationLevel,
	}

	if client.defaultRegistrationLevel == "" {
		client.defaultRegistrationLevel = frejam.RL_EXTENDED
	}

	if client.timeout == 0 {
		client.timeout = time.Minute * 2
	}
	if client.pollInterval < time.Second {
		client.pollInterval = time.Second
	}

	client.client, err = mtls.CreateHTTPClient(client.pemRootCA, client.pemClientCert, client.pemClientKey)
	client.api = &API{parent: client}

	if len(config.PemJWSCert) > 0 {
		client.jwsPubKey, err = extractKeyFromCertPEM(config.PemJWSCert)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return
	}
	return
}

func (c *Client) EID() eid.Client {
	return &eeid{parent: c}
}

func (c *Client) VerifyJWS(v frejam.Verifiable) error {
	return jws.Verify(v.JWSToken(), c.jwsPubKey)
}

func (c *Client) Ping() error {

	res, err := c.client.Get(c.baseURL)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(res.Body)

	return err
}

func (c *Client) API() *API {
	return c.api
}

func (c *Client) AuthInit(ctx context.Context, authReq frejam.AuthRequest) (authRef string, err error) {
	if authReq.MinRegistrationLevel == "" {
		authReq.MinRegistrationLevel = c.defaultRegistrationLevel
	}

	return c.api.AuthInitRequest(ctx, authReq)
}

func (c *Client) AuthCollect(ctx context.Context, authRef string, cancelOnErr bool) (resp *frejam.AuthResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr {
			go func() {
				fmt.Println("Canceling order,", err)
				err := c.api.AuthCancelRequest(ctx, authRef)
				if err != nil {
					fmt.Println("could not cancel auth", err)
				}
			}()
		}
	}()

	for {
		select {

		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(c.pollInterval):
		}

		resp, err = c.api.AuthGetOneResult(ctx, authRef)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case frejam.STATUS_APPROVED, frejam.STATUS_CANCELED, frejam.STATUS_RP_CANCELED, frejam.STATUS_EXPIRED, frejam.STATUS_REJECTED:
			return resp, nil
		}
	}
}

func (c *Client) SignInit(ctx context.Context, signReq frejam.SignRequest) (signRef string, err error) {
	if signReq.MinRegistrationLevel == "" {
		signReq.MinRegistrationLevel = c.defaultRegistrationLevel
	}
	return c.api.SignInitRequest(ctx, signReq)
}

func (c *Client) SignCollect(ctx context.Context, signRef string, cancelOnErr bool) (resp *frejam.SignResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr {
			go func() {
				err := c.api.SignCancelRequest(ctx, signRef)
				if err != nil {
					fmt.Println("could not cancel auth", err)
				}
			}()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(c.pollInterval):
		}

		resp, err = c.api.SignGetOneResult(ctx, signRef)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case frejam.STATUS_APPROVED, frejam.STATUS_CANCELED, frejam.STATUS_RP_CANCELED, frejam.STATUS_EXPIRED, frejam.STATUS_REJECTED:
			return resp, nil
		}
	}
}
