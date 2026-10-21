package ordertoken

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
	"time"
)

func ecKeys(t *testing.T, curve elliptic.Curve) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	priv, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: priv})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub}))
}

func aesKey(t *testing.T) []string {
	t.Helper()
	k := make([]byte, 16)
	_, err := rand.Read(k)
	if err != nil {
		t.Fatal(err)
	}
	return []string{"1:aes:" + base64.StdEncoding.EncodeToString(k)}
}

// Keys that each parse can still fail together; started orders would then get
// no token, so the manager refuses them up front.
func TestNewManager_refuses_keys_that_cannot_make_tokens(t *testing.T) {
	priv, pub := ecKeys(t, elliptic.P256())
	_, otherPub := ecKeys(t, elliptic.P256())
	p384, p384Pub := ecKeys(t, elliptic.P384())

	tests := []struct {
		name      string
		priv, pub string
		keys      []string
	}{
		{"no encryption key", priv, pub, nil},
		{"public key from another pair", priv, otherPub, aesKey(t)},
		{"not a P-256 key", p384, p384Pub, aesKey(t)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewManager(tt.priv, tt.pub, tt.keys); err == nil {
				t.Error("manager accepted keys that cannot make a token")
			}
		})
	}

	if _, err := NewManager(priv, pub, aesKey(t)); err != nil {
		t.Errorf("working keys refused: %v", err)
	}
}

func TestValidateExpire(t *testing.T) {
	// time.Duration is nanoseconds on the wire, so 300 meant as seconds is 300ns
	for _, d := range []time.Duration{-time.Minute, 0, 300, 500 * time.Millisecond, MinExpire - 1} {
		if ValidateExpire(d) == nil {
			t.Errorf("%v accepted, the token would not outlive its order", d)
		}
	}
	for _, d := range []time.Duration{MinExpire, 5 * time.Minute} {
		if err := ValidateExpire(d); err != nil {
			t.Errorf("%v refused: %v", d, err)
		}
	}
}
