package bankid

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/modfin/twofer/eid"
	"github.com/modfin/twofer/eid/bankid/bankidm"
	"github.com/modfin/twofer/internal/mtls"
	"net/http"
	"time"
)

type ClientConfig struct {
	BaseURL string

	PemRootCA     []byte
	PemClientCert []byte
	PemClientKey  []byte

	Timeout time.Duration
}

func New(config ClientConfig) (client *Client, err error) {
	client = &Client{
		baseURL:       config.BaseURL,
		pemClientKey:  config.PemClientKey,
		pemClientCert: config.PemClientCert,
		pemRootCA:     config.PemRootCA,
		timeout:       config.Timeout,

		stop:       make(chan struct{}),
		infromAuth: make(chan string, 1),
		infromSign: make(chan string, 1),
	}

	p, _ := pem.Decode(client.pemClientCert)
	if p == nil {
		return nil, errors.New("couldn't decode client cert pem")
	}
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, err
	}
	client.parsedClientCert = *cert

	if client.timeout == 0 {
		client.timeout = time.Minute * 2
	}
	client.client, err = mtls.CreateHTTPClient(client.pemRootCA, client.pemClientCert, client.pemClientKey)
	client.api = &API{parent: client}

	if err != nil {
		return
	}
	return
}

type Client struct {
	baseURL string

	client *http.Client
	api    *API

	stop       chan struct{}
	infromAuth chan string
	infromSign chan string

	pemRootCA     []byte
	pemClientCert []byte
	pemClientKey  []byte

	timeout time.Duration

	parsedClientCert x509.Certificate
}

func (c *Client) EID() eid.Client {
	return &eeid{parent: c}
}

func (c *Client) API() *API {
	return c.api
}

func (c *Client) Change(ctx context.Context, orderRef string, cancelOnErr bool) (resp *bankidm.CollectResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr {
			go func() {
				fmt.Println("Canceling order,", err)
				err := c.api.Cancel(ctx, orderRef)
				if err != nil {
					fmt.Println("could not cancel order", err)
				}
			}()
		}
	}()

	startState, err := c.api.Collect(ctx, orderRef)
	if err != nil {
		return nil, err
	}
	switch startState.Status {
	case bankidm.STATUS_FAILED, bankidm.STATUS_COMPLETE:
		return resp, nil
	}
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(time.Second):
		}

		resp, err = c.api.Collect(ctx, orderRef)
		if err != nil {
			return nil, err
		}

		if resp.HintCode != startState.HintCode {
			return resp, nil
		}
	}
}

func (c *Client) Collect(ctx context.Context, orderRef string, cancelOnErr bool) (resp *bankidm.CollectResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr {
			go func() {
				fmt.Println("Canceling order,", err)
				err := c.api.Cancel(ctx, orderRef)
				if err != nil {
					fmt.Println("could not cancel order", err)
				}
			}()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(time.Second):
		}

		resp, err = c.api.Collect(ctx, orderRef)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case bankidm.STATUS_FAILED, bankidm.STATUS_COMPLETE:
			return resp, nil
		}
	}
}

func (c *Client) ParsedClientCert() x509.Certificate {
	return c.parsedClientCert
}
