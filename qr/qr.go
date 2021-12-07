package qr

type QRData struct {
	Reference string `json:"reference"`
	Image     []byte `json:"image"`
}
