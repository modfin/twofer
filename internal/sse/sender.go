package sse

import (
	"errors"
	"fmt"
	"net/http"
)

// Sender represents a http server-sent events sender used to send compatible SSE events.
// See https://html.spec.whatwg.org/multipage/server-sent-events.html
type Sender struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

func NewSender(writer http.ResponseWriter) (*Sender, error) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return nil, errors.New("failed to instantiate a http.Flusher from the response writer")
	}

	return &Sender{
		writer:  writer,
		flusher: flusher,
	}, nil
}

func (s *Sender) Prepare() {
	s.writer.Header().Set("Content-Type", "text/event-stream")
	s.writer.Header().Set("Cache-Control", "no-cache")
	s.writer.Header().Set("Connection", "keep-alive")
	s.writer.Header().Set("Transfer-Encoding", "chunked")
}

type Event struct {
	Id    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
}

func (s *Sender) Send(event Event) error {
	_, err := fmt.Fprintf(s.writer, "id: %s\n", event.Id)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "event: %s\n", event.Event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "data: %s\n", event.Data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "\n")
	if err != nil {
		return err
	}

	s.flusher.Flush()

	return nil
}
