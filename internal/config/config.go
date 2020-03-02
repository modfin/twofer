package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"net/url"
	"strings"
	"sync"
)

type Config struct {
	Port int `env:"PORT" envDefault:"43210"`

	FrejaID FrejaID
	BankID  BankID

	QREnabled bool `env:"QR_ENABLE" envDefault:"TRUE"`
	OTP       OTP
	WebAuthn  WebAuthn
}

func (c Config) EIDEnabled() bool {
	return c.BankID.Enabled
}

type OTP struct {
	Enabled       bool     `env:"OTP_ENABLE" envDefault:"TRUE"`
	EncryptionKey []string `env:"OTP_ENCRYPTION_KEY" envSeparator:" "`
	RateLimit     uint     `env:"OTP_RATE_LIMIT" envDefault:"10"`
	SkewCounter   uint     `env:"OTP_SKEW_COUNTER" envDefault:"5"`
	SkewTime      uint     `env:"OTP_SKEW_TIME" envDefault:"1"`
}

type BankID struct {
	Enabled        bool     `env:"EID_BANKID_ENABLE" envDefault:"FALSE"`
	URL            *url.URL `env:"EID_BANKID_URL"`
	RootCA         string   `env:"EID_BANKID_ROOT_CA_PEM"`
	RootCAFile     string   `env:"EID_BANKID_ROOT_CA_PEM_FILE,file"`
	ClientCert     string   `env:"EID_BANKID_CLIENT_CERT"`
	ClientCertFile string   `env:"EID_BANKID_CLIENT_CERT_FILE,file"`
	ClientKey      string   `env:"EID_BANKID_CLIENT_KEY"`
	ClientKeyFile  string   `env:"EID_BANKID_CLIENT_KEY_FILE,file"`
}

func (b BankID) GetRootCA() []byte {
	if b.RootCAFile != "" {
		return []byte(b.RootCAFile)
	}
	return []byte(b.RootCA)
}
func (b BankID) GetClientCert() []byte {
	if b.ClientCertFile != "" {
		return []byte(b.ClientCertFile)
	}
	return []byte(b.ClientCert)
}
func (b BankID) GetClientKey() []byte {
	if b.ClientKeyFile != "" {
		return []byte(b.ClientKeyFile)
	}
	return []byte(b.ClientKey)
}

type FrejaID struct {
	Enabled        bool     `env:"EID_FREJA_ENABLE" envDefault:"FALSE"`
	URL            *url.URL `env:"EID_FREJA_URL"`
	RootCA         string   `env:"EID_FREJA_ROOT_CA_PEM"`
	RootCAFile     string   `env:"EID_FREJA_ROOT_CA_PEM_FILE,file"`
	ClientCert     string   `env:"EID_FREJA_CLIENT_CERT"`
	ClientCertFile string   `env:"EID_FREJA_CLIENT_CERT_FILE,file"`
	ClientKey      string   `env:"EID_FREJA_CLIENT_KEY"`
	ClientKeyFile  string   `env:"EID_FREJA_CLIENT_KEY_FILE,file"`
	JWSCert        string   `env:"EID_FREJA_JWS_CERT"`
	JWSCertFile    string   `env:"EID_FREJA_JWS_CERT_FILE,file"`
}

func (b FrejaID) GetRootCA() []byte {
	if b.RootCAFile != "" {
		return []byte(b.RootCAFile)
	}
	return []byte(b.RootCA)
}
func (b FrejaID) GetClientCert() []byte {
	if b.ClientCertFile != "" {
		return []byte(b.ClientCertFile)
	}
	return []byte(b.ClientCert)
}
func (b FrejaID) GetClientKey() []byte {
	if b.ClientKeyFile != "" {
		return []byte(b.ClientKeyFile)
	}
	return []byte(b.ClientKey)
}
func (b FrejaID) GetJWSCert() []byte {
	if b.JWSCertFile != "" {
		return []byte(b.JWSCertFile)
	}
	return []byte(b.JWSCert)
}

type WebAuthn struct {
	Enabled       bool   `env:"WEBAUTHN_ENABLED" envDefault:"FALSE"`
	RPDisplayName string `env:"RELYING_PARTY_DISPLAY_NAME" envDefault:"localhost"`
	RPID          string `env:"RELYING_PARTY_ID" envDefault:"localhost"`
	RPOrigin      string `env:"RELYING_PARTY_ORIGIN" envDefault:"localhost"`
	SigningKey    string `env:"SigningKey" envDefault:""`
}

var once sync.Once
var config Config

func Get() Config {
	once.Do(func() {
		err := env.Parse(&config)

		if err != nil {

			if strings.HasPrefix(err.Error(), "env: could not load content of file \"\"") {
				fmt.Printf("%+v\n\n", config)
			}

			//TODO something smart if things fail
			panic(err)
		}
	})

	return config
}