package api

type (
	BankIdV6Response struct {
		OrderRef       string                  `json:"orderRef"`
		CollectError   string                  `json:"error,omitempty"`
		ErrorCode      string                  `json:"errorCode,omitempty"`
		URI            string                  `json:"uri,omitempty"`
		QR             string                  `json:"qr,omitempty"`
		Status         string                  `json:"status,omitempty"`
		HintCode       string                  `json:"hintCode,omitempty"`
		CompletionData *BankIdV6CompletionData `json:"completionData,omitempty"`
	}
	BankIdV6CompletionData struct {
		User            BankIdV6User   `json:"user,omitempty"`
		Device          BankIdV6Device `json:"device,omitempty"`
		BankIdIssueDate string         `json:"bankIdIssueDate,omitempty"`
		StepUp          BankIdV6StepUp `json:"stepUp,omitempty"`
		Signature       string         `json:"signature,omitempty"`
		OcspResponse    string         `json:"ocspResponse,omitempty"`
	}
	BankIdV6User struct {
		PersonalNumber string `json:"personalNumber,omitempty"`
		Name           string `json:"name,omitempty"`
		GivenName      string `json:"givenName,omitempty"`
		SurName        string `json:"surName,omitempty"`
	}
	BankIdV6Device struct {
		IpAddress string `json:"ipAddress,omitempty"`
		UHI       string `json:"uhi,omitempty"`
	}
	BankIdV6StepUp struct {
		MRTD bool `json:"mrtd,omitempty"`
	}
)

const (
	StatusPending  = "pending"
	StatusComplete = "complete"
	StatusFailed   = "failed"
	StatusError    = "error"
)
