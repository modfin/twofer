package bankid

import "errors"

type AuthSignAPIResponse struct {
	OrderRef string `json:"orderRef"`
	URI      string `json:"uri"`
	QR       string `json:"qr"`
}

type GenericResponse struct {
	Message string `json:"message"`
}

type ChangeRequest struct {
	OrderRef string `json:"orderRef"`

	// WaitUntilFinished allows the request to wait until the referenced request has either completed or failed,
	// and will not return on state changes during the ongoing process.
	WaitUntilFinished bool `json:"onFinished"`
}

func (r *ChangeRequest) Validate() error {
	if r.OrderRef == "" {
		return errors.New("missing order ref")
	}

	return nil
}
