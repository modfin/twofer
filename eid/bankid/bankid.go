package bankid

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"twofer/eid/bankid/bankidm"
	"twofer/eid"
	"twofer/mtls"
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
}



func (c *Client) EID() eid.Client {
	return &eeid{parent: c}
}

func (c *Client) Ping() (ok bool) {
	_, err := c.client.Get(c.baseURL)
	if err != nil {
		fmt.Println(err)
	}
	return err == nil
}

func (c *Client) API() *API {
	return c.api
}

func (c *Client) AuthInit(ctx context.Context, authReq bankidm.AuthRequest) (*bankidm.AuthResponse, error) {
	return c.API().Auth(authReq)
}
// Auth
// canceling the context will clean up and cancel send a cancel request to freja
func (c *Client) Auth(ctx context.Context, authReq bankidm.AuthRequest) (*bankidm.CollectResponse, error) {
	res, err := c.AuthInit(ctx, authReq)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.Collect(ctx, res.OrderRef.OrderRef, true)
}

func (c *Client) SignInit(ctx context.Context, authReq bankidm.SignRequest) (*bankidm.SignResponse, error) {
	return c.API().Sign(authReq)
}


func (c *Client) Sign(ctx context.Context, authReq bankidm.SignRequest) (*bankidm.CollectResponse, error) {
	res, err := c.SignInit(ctx, authReq)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.Collect(ctx, res.OrderRef.OrderRef, true)
}


func (c *Client) Collect(ctx context.Context, orderRef string, cancelOnErr bool) (resp *bankidm.CollectResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr{
			go func() {
				fmt.Println("Canceling order,", err)
				err := c.api.Cancel(orderRef)
				if err != nil{
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

		resp, err = c.api.Collect(orderRef)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case bankidm.STATUS_FAILED, bankidm.STATUS_COMPLETE:
			return resp, nil
		}
	}
}
