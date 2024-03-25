package servotp

type Alg int32

const (
	Alg_SHA_1   Alg = 0
	Alg_SHA_256 Alg = 1
	Alg_SHA_512 Alg = 2
)

var Alg_name = map[int32]string{
	0: "SHA_1",
	1: "SHA_256",
	2: "SHA_512",
}
var Alg_value = map[string]int32{
	"SHA_1":   0,
	"SHA_256": 1,
	"SHA_512": 2,
}

type Mode int32

const (
	Mode_TIME    Mode = 0
	Mode_COUNTER Mode = 1
)

var Mode_name = map[int32]string{
	0: "TIME",
	1: "COUNTER",
}
var Mode_value = map[string]int32{
	"TIME":    0,
	"COUNTER": 1,
}

type Digits int32

const (
	Digits_SIX   Digits = 0
	Digits_EIGHT Digits = 1
)

var Digits_name = map[int32]string{
	0: "SIX",
	1: "EIGHT",
}
var Digits_value = map[string]int32{
	"SIX":   0,
	"EIGHT": 1,
}

type Enrollment struct {
	Issuer     string `json:"issuer,omitempty"`
	Account    string `json:"account,omitempty"`
	Alg        Alg    `json:"alg,omitempty"`
	Mode       Mode   `json:"mode,omitempty"`
	Digits     Digits `json:"digits,omitempty"`
	Period     uint32 `json:"period,omitempty"`
	SecretSize uint32 `json:"secretSize,omitempty"`
}

func (m *Enrollment) GetIssuer() string {
	if m != nil {
		return m.Issuer
	}
	return ""
}

func (m *Enrollment) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *Enrollment) GetAlg() Alg {
	if m != nil {
		return m.Alg
	}
	return Alg_SHA_1
}

func (m *Enrollment) GetMode() Mode {
	if m != nil {
		return m.Mode
	}
	return Mode_TIME
}

func (m *Enrollment) GetDigits() Digits {
	if m != nil {
		return m.Digits
	}
	return Digits_SIX
}

func (m *Enrollment) GetPeriod() uint32 {
	if m != nil {
		return m.Period
	}
	return 0
}

func (m *Enrollment) GetSecretSize() uint32 {
	if m != nil {
		return m.SecretSize
	}
	return 0
}

type EnrollmentResponse struct {
	Uri      string `json:"uri,omitempty"`
	UserBlob string `json:"userBlob,omitempty"`
}

func (m *EnrollmentResponse) GetUri() string {
	if m != nil {
		return m.Uri
	}
	return ""
}

func (m *EnrollmentResponse) GetUserBlob() string {
	if m != nil {
		return m.UserBlob
	}
	return ""
}

type Credentials struct {
	Otp      string `json:"otp,omitempty"`
	UserBlob string `json:"userBlob,omitempty"`
}

func (m *Credentials) GetOtp() string {
	if m != nil {
		return m.Otp
	}
	return ""
}

func (m *Credentials) GetUserBlob() string {
	if m != nil {
		return m.UserBlob
	}
	return ""
}

type AuthResponse struct {
	Valid    bool   `json:"valid,omitempty"`
	UserBlob string `json:"userBlob,omitempty"`
}

func (m *AuthResponse) GetValid() bool {
	if m != nil {
		return m.Valid
	}
	return false
}

func (m *AuthResponse) GetUserBlob() string {
	if m != nil {
		return m.UserBlob
	}
	return ""
}

type Blob struct {
	UserBlob string `json:"userBlob,omitempty"`
}

func (m *Blob) GetUserBlob() string {
	if m != nil {
		return m.UserBlob
	}
	return ""
}
