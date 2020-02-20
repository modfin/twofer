package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

func CreateHTTPClient(pemRootCA []byte, pemClientCert []byte, pemClientKey []byte) (c *http.Client, err error) {

	// using there your own root cert and client cert (mTLS)
	// https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go

	rootPool := x509.NewCertPool()
	rootPool.AppendCertsFromPEM(pemRootCA)

	cert, err := tls.X509KeyPair(pemClientCert, pemClientKey)
	if err != nil {
		return
	}

	c = &http.Client{}

	trans := http.DefaultTransport.(*http.Transport).Clone()
	trans.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootPool,
	}
	c.Transport = trans

	return c, nil
}
