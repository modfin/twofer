package stream

type (
	Writer interface {
		SendJSON(string, string, any) error
	}
	Encoder func(string, string, any) error
)
