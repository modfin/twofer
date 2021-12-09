package servotp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"golang.org/x/net/context"
	"net/url"
	"strconv"
	"strings"
	"time"
	"github.com/modfin/twofer/grpc/gotp"
	"github.com/modfin/twofer/grpc/gqr"
	"github.com/modfin/twofer/internal/crypt"
	"github.com/modfin/twofer/internal/ratelimit"
)

type OTPConfig struct {
	SkewCounter uint
	SkewTime    uint
	RateLimit   uint
}

func New(conf OTPConfig, keys []string) (*Server, error) {
	s := &Server{
		conf: conf,
	}
	var err error
	if len(keys) > 0 {
		s.store, err = crypt.New(keys)
		if err != nil {
			return nil, err
		}
	}

	if s.store == nil {
		s.store = &crypt.NilStore{}
	}

	if s.conf.RateLimit == 0 {
		s.conf.RateLimit = 10
	}

	s.ratelimiter = ratelimit.New(s.conf.RateLimit)

	return s, nil
}

type rlItem struct {
	start time.Time
	count uint
}

type Server struct {
	store       crypt.Store
	conf        OTPConfig
	ratelimiter *ratelimit.Ratelimiter
}

func (s *Server) Upgrade(context.Context, *gotp.Blob) (*gotp.Blob, error) {
	panic("implement me")
}

type wrapper struct {
	URI     string `json:"uri"`
	Counter uint64 `json:"counter,omitempty"`
}

func (s *Server) Enroll(ctx context.Context, en *gotp.Enrollment) (resp *gotp.EnrollmentResponse, err error) {

	digits := otp.DigitsSix
	switch en.Digits {
	case gotp.Digits_SIX:
		digits = otp.DigitsSix
	case gotp.Digits_EIGHT:
		digits = otp.DigitsEight
	}

	if en.SecretSize < 20 {
		en.SecretSize = 20
	}

	var o wrapper
	switch en.Mode {
	case gotp.Mode_TIME:
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

	case gotp.Mode_COUNTER:
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

	b, err = s.store.Encrypt(b)
	if err != nil {
		return nil, err
	}

	return &gotp.EnrollmentResponse{
		Uri:      o.URI,
		UserBlob: base64.StdEncoding.EncodeToString(b),
	}, nil
}

func (s *Server) Auth(ctx context.Context, va *gotp.Credentials) (*gotp.AuthResponse, error) {

	sec, err := base64.StdEncoding.DecodeString(va.UserBlob)
	if err != nil {
		return nil, err
	}
	sec, err = s.store.Decrypt(sec)
	if err != nil {
		return nil, err
	}

	var v wrapper
	err = json.Unmarshal(sec, &v)
	if err != nil {
		return nil, err
	}

	// Checking ratelimit
	uri, err := url.Parse(v.URI)
	if err != nil {
		return nil, err
	}

	err = s.ratelimiter.Hit(uri.Host + uri.Path)
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
			Skew:      s.conf.SkewTime,
			Digits:    didgets,
			Algorithm: alg,
		})
		if err != nil {
			return nil, err
		}
	case "hotp":
		for i := uint64(0); i <= uint64(s.conf.SkewCounter); i++ {
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

	sec, err = s.store.Encrypt(sec)
	if err != nil {
		return nil, err
	}

	return &gotp.AuthResponse{
		Valid:    valid,
		UserBlob: base64.StdEncoding.EncodeToString(sec),
	}, nil

}

func (s *Server) GetQRImage(ctx context.Context, va *gotp.Credentials) (*gqr.Image, error) {
	sec, err := base64.StdEncoding.DecodeString(va.UserBlob)
	if err != nil {
		return nil, err
	}
	sec, err = s.store.Decrypt(sec)
	if err != nil {
		return nil, err
	}

	var v wrapper
	err = json.Unmarshal(sec, &v)
	if err != nil {
		return nil, err
	}

	data := gqr.Data{
		RecoveryLevel:        0,
		Size:                 0,
		Data:                 v.URI,
	}

	size := int(data.Size)

	if size < 10 {
		size = 256
	}

	level := qrcode.RecoveryLevel(data.RecoveryLevel)
	image, err := qrcode.Encode(data.Data, level, size)

	return &gqr.Image{
		Data:        image,
		ContentType: "image/png",
	}, err
}
