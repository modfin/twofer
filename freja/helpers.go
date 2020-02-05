package freja

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
)

func createHTTPClient(pemRootCA []byte, pemClientCert []byte, pemClientKey []byte) (c *http.Client, err error){

	// using there your own root cert and client cert (mTLS)
	// https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go

	rootPool := x509.NewCertPool()
	rootPool.AppendCertsFromPEM(pemRootCA)


	cert, err := tls.X509KeyPair(pemClientCert, pemClientKey)
	if err != nil {
		return
	}

	c = &http.Client{}
	trans := &(*http.DefaultTransport.(*http.Transport))
	trans.TLSClientConfig = &tls.Config{
		Certificates:                []tls.Certificate{cert},
		RootCAs:                     rootPool,
	}
	c.Transport = trans

	return c, nil
}

func extractKeyFromCertPEM(pubPEM []byte) (*rsa.PublicKey, error){
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch cert.PublicKey.(type) {
	case  *rsa.PublicKey:
		return cert.PublicKey.( *rsa.PublicKey), nil
	}
	return nil, errors.New("the certificate does not contain a rsa public key")
}


func authtopic(authRef string) string{
		return fmt.Sprintf("%s%s", pubSubAuthPrefix, authRef)
}

func signtopic(signRef string) string{
	return fmt.Sprintf("%s%s", pubSubSignPrefix, signRef)
}


