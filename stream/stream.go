package stream

type (
	Writer interface {
		SendJSON(string, any) error
	}
	Encoder func(string, any) error
)
