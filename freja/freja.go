package freja

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/jws"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"twofer"
	"twofer/freja/mfreja"
)

type Client struct {
	production bool

	pubsub twofer.PubSub
	client *http.Client
	api    *API

	stop       chan struct{}
	infromAuth chan string
	infromSign chan string

	pemRootCA     []byte
	pemClientCert []byte
	pemClientKey  []byte

	jwsPubKey *rsa.PublicKey

	timeout time.Duration
}

type ClientConfig struct {
	Production    bool
	PemRootCA     []byte
	PemClientCert []byte
	PemClientKey  []byte

	// If present the JWS tokens are validated, otherwise everything is let through
	PemJWSCert []byte

	Timeout time.Duration
}

func New(config ClientConfig, pubsub twofer.PubSub) (client *Client, err error) {
	client = &Client{
		production:    config.Production,
		pubsub:        pubsub,
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
	client.client, err = createHTTPClient(client.pemRootCA, client.pemClientCert, client.pemClientKey)
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
	go client.authPoller()
	go client.signPoller()
	return
}

func (c *Client) baseURL() string {
	if c.production {
		return ProdURL
	}
	return TestURL
}
func (c *Client) baseResourceURL() string {
	if c.production {
		return ProdResourceURL
	}
	return TestResourceURL
}

func (c *Client) VerifyJWS(v mfreja.Verifiable) error {
	return jws.Verify(v.JWSToken(), c.jwsPubKey)
}

func (c *Client) Stop() {
	close(c.stop)
}

func (c *Client) Stopped() bool {
	select {
	case <-c.stop:
		return true
	default:
	}
	return false
}

func (c *Client) Ping() (ok bool) {
	if c.Stopped() {
		return false
	}

	res, err := c.client.Get(c.baseURL() + "ping")
	if err != nil {
		fmt.Println(err)
		return false
	}
	d, err := ioutil.ReadAll(res.Body)

	fmt.Println(err)
	fmt.Println(string(d))
	return err == nil
}

func (c *Client) API() *API {
	return c.api
}

func (c *Client) qrLink(authRef string) string {
	ref := fmt.Sprintf("frejaeid://bindUserToTransaction?transactionReference=%s", url.QueryEscape(authRef))
	u := fmt.Sprintf("%sqrcode/generate?qrcodedata=%s", c.baseResourceURL(), ref)
	return u
}

// Auth
// canceling the context will clean up and cancel send a cancel request to freja
func (c *Client) Auth(ctx context.Context, authReq mfreja.AuthRequest) (*mfreja.AuthResponse, error) {
	if c.Stopped() {
		return nil, frejaStoppedError
	}

	authRef, err := c.api.AuthInitRequest(ctx, authReq)
	if err != nil {
		return nil, err
	}

	c.infromAuth <- authRef

	fmt.Println(c.qrLink(authRef))



	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	data, err := c.pubsub.Next(ctx, authtopic(authRef))

	if ctx.Err() != nil { // Context was canceled. Lets cancel the request to freja
		go c.api.AuthCancelRequest(authRef)
	}
	if err != nil {
		return nil, err
	}

	var resp mfreja.AuthResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if c.jwsPubKey != nil {
		err = c.VerifyJWS(resp)
	}

	return &resp, err
}

type AsyncAuthResponse struct {
	Response *mfreja.AuthResponse
	Err      error
}

func (c *Client) AuthAsync(ctx context.Context, authReq mfreja.AuthRequest) <-chan *AsyncAuthResponse {
	r := make(chan *AsyncAuthResponse, 1)
	go func() {
		a := AsyncAuthResponse{}
		a.Response, a.Err = c.Auth(ctx, authReq)
		r <- &a
		close(r)
	}()
	return r
}

// Sign
// canceling the context will clean up and cancel send a cancel request to freja
func (c *Client) Sign(ctx context.Context, signReq mfreja.SignRequest) (*mfreja.AuthResponse, error) {
	if c.Stopped() {
		return nil, frejaStoppedError
	}

	signRef, err := c.api.SignInitRequest(ctx, signReq)
	if err != nil {
		return nil, err
	}

	c.infromSign <- signRef

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	data, err := c.pubsub.Next(ctx, signtopic(signRef))

	if ctx.Err() != nil { // Context was canceled. Lets cancel the request to freja
		go c.api.SignCancelRequest(signRef)
	}
	if err != nil {
		return nil, err
	}

	var resp mfreja.AuthResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if c.jwsPubKey != nil {
		err = c.VerifyJWS(resp)
	}

	return &resp, nil
}

func (c *Client) authPoller() {
	// TODO defer on panics and start authPoller again ...

	var zeroTime time.Time

	// TODO clean seen
	seen := map[string]time.Time{}
	waitingFor := map[string]struct{}{}

	process := func(resp *mfreja.AuthResponse) {
		if seen[resp.AuthRef.AuthRef] == zeroTime {
			switch resp.Status {
			case mfreja.STATUS_APPROVED, mfreja.STATUS_CANCELED, mfreja.STATUS_RP_CANCELED, mfreja.STATUS_EXPIRED, mfreja.STATUS_REJECTED:

				fmt.Printf("Auth %s: %v\n", resp.Status, resp.AuthRef.AuthRef)

				data, err := json.Marshal(resp)
				if err != nil {
					fmt.Printf("could not marchal resp: %s\n", resp.AuthRef.AuthRef)
					return
				}

				err = c.pubsub.Notify(authtopic(resp.AuthRef.AuthRef), data)
				if err != nil {
					fmt.Printf("could notify on resp: %s\n", resp.AuthRef.AuthRef)
					return
				}

				delete(waitingFor, resp.AuthRef.AuthRef)
				seen[resp.AuthRef.AuthRef] = time.Now()
			}
		}
	}

	poll := func() {
		if 0 < len(waitingFor) && len(waitingFor) < 5 {
			for key, _ := range waitingFor {
				resp, err := c.api.AuthGetOneResult(key)
				if err != nil {
					fmt.Println("single poll err", err)
					continue
				}
				process(resp)
			}
			return
		}

		resps, err := c.api.AuthGetResults()
		if err != nil {
			fmt.Println("poll err", err)
			return
		}
		for _, resp := range resps {
			process(&resp)
		}
	}

	const stdSleep = time.Second * 30
	sleep := stdSleep
	resetTime := time.Now()
	for {
		select {
		case <-time.After(sleep):
			fmt.Println("Resetting auth poll")
			poll()
		case ref := <-c.infromAuth:
			fmt.Println("Informed auth poll")
			waitingFor[ref] = struct{}{}
			sleep = time.Second
			resetTime = time.Now()
			poll()
		case <-c.stop:
			fmt.Println("turning of auth poller")
			return
		}

		if sleep < stdSleep {
			dur := time.Since(resetTime)
			if len(waitingFor) == 0 {
				sleep = stdSleep
			} else if dur > 90*time.Second {
				sleep = stdSleep
			} else if dur > 60*time.Second {
				sleep = 3 * time.Second
			} else if dur > 30*time.Second {
				sleep = 2 * time.Second
			}
		}

	}
}

func (c *Client) signPoller() {
	// TODO defer on panics and start authPoller again ...

	var zeroTime time.Time

	// TODO clean seen
	seen := map[string]time.Time{}
	waitingFor := map[string]struct{}{}

	process := func(resp *mfreja.SignResponse) {
		if seen[resp.SignRef.SignRef] == zeroTime {
			switch resp.Status {
			case mfreja.STATUS_APPROVED, mfreja.STATUS_CANCELED, mfreja.STATUS_RP_CANCELED, mfreja.STATUS_EXPIRED, mfreja.STATUS_REJECTED:
				fmt.Printf("Sign: %s: %v\n", resp.Status, resp.SignRef.SignRef)
				data, err := json.Marshal(resp)
				if err != nil {
					fmt.Printf("could not marchal resp: %s\n", resp.SignRef.SignRef)
					return
				}
				err = c.pubsub.Notify(signtopic(resp.SignRef.SignRef), data)
				if err != nil {
					fmt.Printf("could notify on resp: %s\n", resp.SignRef.SignRef)
					return
				}

				delete(waitingFor, resp.SignRef.SignRef)
				seen[resp.SignRef.SignRef] = time.Now()
			}
		}
	}

	poll := func() {

		if 0 < len(waitingFor) && len(waitingFor) < 5 {
			for key, _ := range waitingFor {
				resp, err := c.api.SignGetOneResult(key)
				if err != nil {
					fmt.Println("single poll err", err)
					continue
				}
				process(resp)
			}
			return
		}

		fmt.Println("Doing sign multi poll")
		// Do this if waitingFor > 5 or 0, switching to multi mode to lessen request load
		resps, err := c.api.SignGetResults()
		if err != nil {
			fmt.Println("multi poll err", err)
			return
		}
		for _, resp := range resps {
			process(&resp)
		}
	}

	const stdSleep = time.Second * 30
	sleep := stdSleep
	resetTime := time.Now()
	for {
		select {
		case <-time.After(sleep):
			fmt.Println("Resetting sign poll")
			poll()
		case ref := <-c.infromSign:
			fmt.Println("Informed sign poll")
			waitingFor[ref] = struct{}{}
			sleep = time.Second
			resetTime = time.Now()
			poll()
		case <-c.stop:
			fmt.Println("turning of sing poller")
			return
		}

		if sleep < stdSleep {
			dur := time.Since(resetTime)
			if len(waitingFor) == 0 {
				sleep = stdSleep
			} else if dur > 90*time.Second {
				sleep = stdSleep
			} else if dur > 60*time.Second {
				sleep = 3 * time.Second
			} else if dur > 30*time.Second {
				sleep = 2 * time.Second
			}
		}

	}
}
