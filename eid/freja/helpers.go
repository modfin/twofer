package freja

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func extractKeyFromCertPEM(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return cert.PublicKey.(*rsa.PublicKey), nil
	}
	return nil, errors.New("the certificate does not contain a rsa public key")
}
