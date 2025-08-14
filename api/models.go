package api

import "time"

type (
	// BankIdV6Response used as a catch all response for V2
	// Deprecated: stick to V1 or move to V3
	BankIdV6Response struct {
		OrderRef       string                  `json:"orderRef"`
		ErrorCode      string                  `json:"errorCode,omitempty"`
		ErrorText      string                  `json:"errorText,omitempty"`
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

	// V3 request / response messages

	// BankIdv6AuthSignRequestV3 is used to start either an auth or sign request against BankID
	BankIdv6AuthSignRequestV3 struct {
		// EndUserIp The user IP address as it is seen by your service. Required.
		EndUserIp string `json:"endUserIp"`

		// ReturnUrl Orders started on the same device as where the user's BankID is stored (started with autostart
		// token) will call this URL when the order is completed
		ReturnUrl string `json:"returnUrl,omitempty"`

		// UserNonVisibleData Data that you wish to include but not display to the user
		UserNonVisibleData string `json:"userNonVisibleData,omitempty"`

		// UserVisibleData Text displayed to the user during the order
		UserVisibleData string `json:"userVisibleData,omitempty"`

		// UserVisibleDataFormat, 'plaintext' or 'simpleMarkdownV1'
		UserVisibleDataFormat string `json:"userVisibleDataFormat,omitempty"`

		// PersonalNumber The personal identity number allowed to confirm the identification
		PersonalNumber string `json:"personalNumber,omitempty"`

		// PinCode User is required to confirm the order with their security code even if they have biometrics activated
		PinCode bool `json:"pinCode,omitempty"`

		// Once if true, will start an auth/sign and just return a single QR code, if false, auth/sign endpoint return
		// an SSE / NDJSON stream and send a new QR-code each second, for 30 seconds before returning
		Once bool `json:"once,omitempty"`

		// OrderTokenExpire if order tokens are enabled, sets token expire time.
		OrderTokenExpire time.Duration `json:"order_token_expire,omitempty"`
	}

	// BankIdV6AuthSignResponseV3 is sent as a successful reply to an auth or sign request. If SSE / NDJSON is used, a
	// new BankIdV6AuthSignResponseV3 is sent each second (for 30 seconds)
	BankIdV6AuthSignResponseV3 struct {
		// OrderRef The reference ID for an order
		OrderRef string `json:"orderRef"`

		// URI Start URL, for "BankID on this device"
		URI string `json:"uri"`

		// QR contain the data for the QR-code
		QR string `json:"qr"`

		// OrderToken is returned if order token support is enabled.
		OrderToken string `json:"orderToken,omitempty"`
	}

	// BankIdv6CollectRequestV3 is used to collect status on a started auth / sign request
	BankIdv6CollectRequestV3 struct {
		// OrderRef A reference ID for an order
		OrderRef string `json:"orderRef"`

		// WaitForChange allows the request to wait until a change is detected
		WaitForChange bool `json:"waitForChange"`

		// WaitUntilFinished allows the request to wait until the referenced request has either completed or failed,
		// and will not return on state changes during the ongoing process.
		WaitUntilFinished bool `json:"waitUntilFinished"`

		// OrderToken optionally pass an order token. This is an alternative to OrderRef.
		// If you use this the end user IP of the user who triggered collect is required for verification
		OrderToken string `json:"orderToken"`
		EndUserIp  string `json:"endUserIp"`
	}

	// BankIdV6CollectResponseV3 is sent for a successful collect, if WaitUntilFinished is set in the request, it will
	// only return once the order have either completed or failed. If WaitUntilFinished isn't set, it will return once
	// a change is detected.
	BankIdV6CollectResponseV3 struct {
		// OrderRef The reference ID for an order
		OrderRef       string                  `json:"orderRef"`
		Status         string                  `json:"status,omitempty"`
		HintCode       string                  `json:"hintCode,omitempty"`
		CompletionData *BankIdV6CompletionData `json:"completionData,omitempty"`
	}

	// BankIdv6CancelRequestV3 request the cancellation of a pending auth / sign request
	BankIdv6CancelRequestV3 struct {
		// OrderRef A reference ID for an order
		OrderRef string `json:"orderRef"`
	}

	BankIdv6CancelResponseV3 struct {
		Status string `json:"status"`
	}

	// BankIdv6ErrorResponseV3 is sent when an endpoint return an error (4xx, 5xx) http status code
	BankIdv6ErrorResponseV3 struct {
		// Origin contains the origin of the error, currently either 'Twofer' or 'BankIDv6'
		Origin string `json:"origin"` // Twofer / BankIDv6

		// StatusCode contain the original HTTP status code, if the error originates from BankID
		StatusCode int `json:"statusCode,omitempty"`

		// ErrorCode contain the original error code that we may get from BankID when they return an http 400
		Code string `json:"code,omitempty"`

		// Detail may contain the original error detail that we may get from BankID when they return an http 400, or it
		// can be an error message generated in twofer, for twofer errors
		Detail string `json:"detail"`
	}
)

// Status codes
const (
	StatusPending  = "pending"
	StatusComplete = "complete"
	StatusFailed   = "failed"
	StatusError    = "error"
	StatusQrCode   = "qrcode"
)

// Error origin codes
const (
	ErrorOriginTwofer   = "Twofer"
	ErrorOriginBankIDv6 = "BankIDv6"
)
