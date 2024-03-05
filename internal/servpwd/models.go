package servpwd

type Alg int32

const (
	Alg_SHA_256 Alg = 0
	Alg_SHA_512 Alg = 1
	Alg_SCrypt  Alg = 2
	Alg_BCrypt  Alg = 3
)

var Alg_name = map[int32]string{
	0: "SHA_256",
	1: "SHA_512",
	2: "SCrypt",
	3: "BCrypt",
}
var Alg_value = map[string]int32{
	"SHA_256": 0,
	"SHA_512": 1,
	"SCrypt":  2,
	"BCrypt":  3,
}

type EnrollReq struct {
	Password string `json:"password,omitempty"`
}

type AuthReq struct {
	Password string `json:"password,omitempty"`
	UserBlob string `json:"userBlob,omitempty"`
}

type Res struct {
	Valid   bool   `json:"valid,omitempty"`
	Message string `json:"message,omitempty"`
}

type Blob struct {
	UserBlob string `json:"userBlob,omitempty"`
}
