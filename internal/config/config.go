package config

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	HTTPPort int `env:"HTTP_PORT" envDefault:"8080"`

	BankID BankID

	QREnabled bool `env:"QR_ENABLE" envDefault:"TRUE"`
	OTP       OTP
	WebAuthn  WebAuthn
	PWD       PWD
}

func (c Config) EIDEnabled() bool {
	return c.BankID.Enabled
}

type OTP struct {
	Enabled       bool     `env:"OTP_ENABLE" envDefault:"FALSE"`
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

type WebAuthn struct {
	Enabled          bool   `env:"WEBAUTHN_ENABLED" envDefault:"FALSE"`
	RPDisplayName    string `env:"WEBAUTHN_RP_DISPLAYNAME"`
	RPID             string `env:"WEBAUTHN_RP_ID"`
	RPOrigin         string `env:"WEBAUTHN_RP_ORIGIN"`
	HMACKey          string `env:"WEBAUTHN_HMAC_KEY"`
	UserVerification string `env:"WEBAUTHN_USER_VERIFICATION" envDefault:"discouraged"`

	RateLimit uint          `env:"WEBAUTHN_RATE_LIMIT" envDefault:"10"`
	Timeout   time.Duration `env:"WEBAUTHN_TIMEOUT" envDefault:"60s"`
}

type PWD struct {
	Enabled       bool     `env:"PWD_ENABLE" envDefault:"FALSE"`
	EncryptionKey []string `env:"PWD_ENCRYPTION_KEY" envSeparator:" "`
	DefaultAlg    int32    `env:"PWD_ALG" envDefault:"0"`

	DefaultHashCount int `env:"PWD_HASH_COUNT" envDefault:"1"`

	DefaultBCryptCost int `env:"PWD_BCRYPT_COST" envDefault:"10"`

	DefaultSCryptN      int `env:"PWD_SCRYPT_N" envDefault:"32768"`
	DefaultSCryptR      int `env:"PWD_SCRYPT_R" envDefault:"8"`
	DefaultSCryptP      int `env:"PWD_SCRYPT_P" envDefault:"1"`
	DefaultSCryptKeyLen int `env:"PWD_SCRYPT_KEY_LEN" envDefault:"32"`
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
