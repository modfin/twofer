package stream

type (
	Writer interface {
		SendJSON(string, string, any) error
	}
	Encoder func(id string, event string, data any) error
)
