package sse

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Sender represents a http server-sent events sender used to send compatible SSE events.
// See https://html.spec.whatwg.org/multipage/server-sent-events.html
type Sender struct {
	response *echo.Response
	writer   http.ResponseWriter
	flush    func()
}

func NewSender(response *echo.Response) (*Sender, error) {
	flusher, ok := response.Writer.(http.Flusher)
	if !ok {
		return nil, errors.New("failed to instantiate a http.Flusher from the response writer")
	}

	return &Sender{
		response: response,
		writer:   response.Writer,
		flush:    flusher.Flush,
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

	s.flush()
	s.response.Committed = true // Needed to fix 'superfluous response.WriteHeader' console messages
	return nil
}
