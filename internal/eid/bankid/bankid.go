package bankid

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/modfin/twofer/internal/bankid"
	bankid_v51 "github.com/modfin/twofer/internal/eid/bankid/v5.1"
	"github.com/modfin/twofer/internal/mtls"
	"net/http"
)

type ClientConfig struct {
	BaseURL string

	PemRootCA     []byte
	PemClientCert []byte
	PemClientKey  []byte
}

func New(config ClientConfig) (client *BankID, err error) {
	client = &BankID{
		baseURL:       config.BaseURL,
		pemClientKey:  config.PemClientKey,
		pemClientCert: config.PemClientCert,
		pemRootCA:     config.PemRootCA,

		//stop:       make(chan struct{}),
		//infromAuth: make(chan string, 1),
		//infromSign: make(chan string, 1),
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

	client.httpClient, err = mtls.CreateHTTPClient(client.pemRootCA, client.pemClientCert, client.pemClientKey)
	if err != nil {
		return nil, err
	}

	client.APIv51 = bankid_v51.NewEid(client.httpClient, config.BaseURL)

	client.APIv60 = bankid.NewAPI(client.httpClient, client.baseURL)

	return
}

type BankID struct {
	APIv51 *bankid_v51.Eeid
	APIv60 *bankid.API

	baseURL string

	httpClient *http.Client

	//stop       chan struct{}
	//infromAuth chan string
	//infromSign chan string

	pemRootCA     []byte
	pemClientCert []byte
	pemClientKey  []byte

	parsedClientCert x509.Certificate
}

func (c *BankID) ParsedClientCert() x509.Certificate {
	return c.parsedClientCert
}
