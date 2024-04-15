package sse

type Event struct {
	Event string
	Data  string
	ID    string
	Retry string
}
