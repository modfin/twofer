package otpserver

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/net/context"
	"net/url"
	"strconv"
	"strings"
	"time"
	"twofer/twoferrpc"
)

func New() *Server {
	return &Server{}
}

type Server struct {
}

type wrapper struct {
	URI     string `json:"uri"`
	Counter uint64 `json:"counter,omitempty"`
}

func (s Server) Enroll(ctx context.Context, en *twoferrpc.OTPEnrollment) (resp *twoferrpc.OTPEnrollmentResponse, err error) {

	digits := otp.DigitsSix
	switch en.Digits {
	case twoferrpc.OTPDigits_SIX:
		digits = otp.DigitsSix
	case twoferrpc.OTPDigits_EIGHT:
		digits = otp.DigitsEight
	}

	if en.SecretSize < 20 {
		en.SecretSize = 20
	}

	var o wrapper
	switch en.Mode {
	case twoferrpc.OTPMode_TIME:
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      en.Issuer,
			AccountName: en.Account,
			Period:      uint(en.Period),
			SecretSize:  uint(en.SecretSize),
			Digits:      digits,
			Algorithm:   otp.Algorithm(en.Alg),
		})
		if err != nil {
			return nil, err
		}
		o.URI = key.URL()

	case twoferrpc.OTPMode_COUNTER:
		key, err := hotp.Generate(hotp.GenerateOpts{
			Issuer:      en.Issuer,
			AccountName: en.Account,
			SecretSize:  uint(en.SecretSize),
			Digits:      digits,
			Algorithm:   otp.Algorithm(en.Alg),
		})
		if err != nil {
			return nil, err
		}
		o.URI = key.URL()
		o.Counter = 1
	default:
		return nil, errors.New("mode must be time or counter")
	}

	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	return &twoferrpc.OTPEnrollmentResponse{
		Uri:    o.URI,
		Secret: base64.StdEncoding.EncodeToString(b),
	}, nil
}

func (s Server) Validate(ctx context.Context, va *twoferrpc.OTPValidate) (*twoferrpc.OTPValidateResponse, error) {

	sec, err := base64.StdEncoding.DecodeString(va.Secret)
	if err != nil {
		return nil, err
	}

	var v wrapper
	err = json.Unmarshal(sec, &v)
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(v.URI)
	if err != nil {
		return nil, err
	}

	var didgets otp.Digits
	switch uri.Query().Get("digits") {
	case "6":
		didgets = otp.DigitsSix
	case "8":
		didgets = otp.DigitsEight
	default:
		didgets = otp.DigitsSix
	}
	var period uint = 30
	p := uri.Query().Get("period")
	if len(p) > 0 {
		pp, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return nil, err
		}
		period = uint(pp)
	}

	var alg otp.Algorithm
	switch strings.ToUpper(uri.Query().Get("algorithm")) {
	case "SHA1":
		alg = otp.AlgorithmSHA1
	case "SHA256":
		alg = otp.AlgorithmSHA256
	case "SHA512":
		alg = otp.AlgorithmSHA512
	default:
		alg = otp.AlgorithmSHA1
	}

	var valid bool
	switch uri.Host {
	case "totp":
		valid, err = totp.ValidateCustom(va.Otp, uri.Query().Get("secret"), time.Now().UTC(), totp.ValidateOpts{
			Period:    period,
			Skew:      1,
			Digits:    didgets,
			Algorithm: alg,
		})
		if err != nil {
			return nil, err
		}
	case "hotp":
		for i := uint64(0); i < 5; i++ {
			valid, err = hotp.ValidateCustom(va.Otp, v.Counter+i, uri.Query().Get("secret"), hotp.ValidateOpts{
				Digits:    didgets,
				Algorithm: alg,
			})
			if err != nil {
				return nil, err
			}
			if !valid {
				continue
			}
			v.Counter += i + 1
			sec, err = json.Marshal(v)
			if err != nil {
				return nil, err
			}
			break
		}

	default:
		return nil, errors.New("otp scheme is not valid " + uri.Host)
	}

	return &twoferrpc.OTPValidateResponse{
		Valid:  valid,
		Secret: base64.StdEncoding.EncodeToString(sec),
	}, nil

}
