package servqr

type Data_Recovery int32

const (
	// Level L: 7% error recovery.
	Data_LOW Data_Recovery = 0
	// Level M: 15% error recovery. Good default choice.
	Data_MEDIUM Data_Recovery = 1
	// Level Q: 25% error recovery.
	Data_HIGH Data_Recovery = 2
	// Level H: 30% error recovery.
	Data_HIGHEST Data_Recovery = 3
)

var Data_Recovery_name = map[int32]string{
	0: "LOW",
	1: "MEDIUM",
	2: "HIGH",
	3: "HIGHEST",
}
var Data_Recovery_value = map[string]int32{
	"LOW":     0,
	"MEDIUM":  1,
	"HIGH":    2,
	"HIGHEST": 3,
}

type Data struct {
	RecoveryLevel Data_Recovery `json:"RecoveryLevel,omitempty"`
	Size          int32         `json:"size,omitempty"`
	Data          string        `json:"data,omitempty"`
}

func (m *Data) GetRecoveryLevel() Data_Recovery {
	if m != nil {
		return m.RecoveryLevel
	}
	return Data_LOW
}

func (m *Data) GetSize() int32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *Data) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

type Image struct {
	ContentType string `json:"contentType,omitempty"`
	Data        []byte `json:"data,omitempty"`
}

type QRData struct {
	Reference string `json:"reference"`
	Image     []byte `json:"image"`
}
