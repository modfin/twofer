package ordertoken

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/modfin/twofer/internal/crypt"
)

type claims struct {
	jwt.RegisteredClaims
	Payload []byte `json:"payload"`
}

type Payload struct {
	OrderRef  string `json:"orderRef"`
	EndUserIp string `json:"endUserIp"`
}

type Manager struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	cryptStore crypt.Store
}

var ErrOrderIpMismatch = errors.New("order ip mismatch")

func NewManager(ec256 string, ec256pub string, encryptionKey []string) (*Manager, error) {
	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(ec256))
	if err != nil {
		return nil, fmt.Errorf("unable to parse ECDSA private key: %w", err)
	}
	publicKey, err := jwt.ParseECPublicKeyFromPEM([]byte(ec256pub))
	if err != nil {
		return nil, fmt.Errorf("unable to parse ECDSA public key: %w", err)
	}
	s, err := crypt.New(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create crypt store: %w", err)
	}
	return &Manager{
		privateKey: privateKey,
		publicKey:  publicKey,
		cryptStore: s,
	}, nil
}

// Parse parses and validates an order token and returns it's encrypted payload
func (m *Manager) Parse(orderToken string, endUserIp string) (Payload, error) {
	var claims claims
	_, err := jwt.ParseWithClaims(orderToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", token.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		return Payload{}, err
	}
	p, err := m.cryptStore.Decrypt(claims.Payload)
	if err != nil {
		return Payload{}, fmt.Errorf("failed to decrypt payload: %w", err)
	}
	var payload Payload
	err = json.Unmarshal(p, &payload)
	if err != nil {
		return Payload{}, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.EndUserIp != endUserIp {
		return Payload{}, ErrOrderIpMismatch
	}
	if payload.OrderRef == "" {
		return Payload{}, errors.New("order ref empty")
	}
	return payload, nil
}

func (m *Manager) Create(expire time.Duration, payload Payload) (string, error) {
	if expire <= 0 {
		return "", errors.New("order token expire must be positive")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	p, err := m.cryptStore.Encrypt(b)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt payload: %w", err)
	}
	t := time.Now()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  &jwt.NumericDate{Time: t},
			ExpiresAt: &jwt.NumericDate{Time: t.Add(expire)},
			Issuer:    "twofer",
		},
		Payload: p,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, c).SignedString(m.privateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}
