package samplem

import (
	"encoding/json"
)

type RegisterResponse struct {
	PublicKey json.RawMessage `json:"publicKey"`
}

type LoginResponse struct {
	PublicKey json.RawMessage `json:"publicKey"`
}
