package sse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"
)

// Sender represents a http server-sent events sender used to send compatible SSE events.
// See https://html.spec.whatwg.org/multipage/server-sent-events.html
type Sender struct {
	eventID  atomic.Uint32
	response *echo.Response
	writer   http.ResponseWriter
	flusher  http.Flusher
}

func NewSender(response *echo.Response) (*Sender, error) {
	flusher, ok := response.Writer.(http.Flusher)
	if !ok {
		return nil, errors.New("failed to instantiate a http.Flusher from the response writer")
	}

	return &Sender{
		response: response,
		writer:   response.Writer,
		flusher:  flusher,
	}, nil
}

func (s *Sender) Prepare() {
	s.writer.Header().Set("Content-Type", "text/event-stream")
	s.writer.Header().Set("Cache-Control", "no-cache")
	s.writer.Header().Set("Connection", "keep-alive")
	s.writer.Header().Set("Transfer-Encoding", "chunked")
}

// type Event struct {
// 	Id    string `json:"id"`
// 	Event string `json:"event"`
// 	Data  string `json:"data"`
// }

func (s *Sender) Send(event string, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("ERR: failed to marshal event message: %v\n", err)
		return err
	}

	_, err = fmt.Fprintf(s.writer, "id: %d\n", s.eventID.Add(1)-1)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "event: %s\n", event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "data: %s\n", bytes)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "\n")
	if err != nil {
		return err
	}

	s.flusher.Flush()
	s.response.Committed = true
	return nil
}
