package sse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/modfin/twofer/stream"
)

type Writer struct {
	mu sync.Mutex
	http.ResponseWriter
	flush func()
}

var (
	ErrInvalidWriter = errors.New("invalid writer")
	ErrWrite         = errors.New("write error")
)

var _ stream.Writer = (*Writer)(nil) // Compile-time check that we implement stream.Writer interface

func NewWriter(w http.ResponseWriter) (*Writer, error) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("%w: the provided writer don't support the http.Flusher interface", ErrInvalidWriter)
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	return &Writer{
		ResponseWriter: w,
		flush:          f.Flush,
	}, nil
}

func NewEncoder(w http.ResponseWriter) (stream.Encoder, error) {
	enc, err := NewWriter(w)
	if err != nil {
		return nil, err
	}

	return enc.SendJSON, nil
}

func writeField(buf *bytes.Buffer, field, data string) {
	// Ignore empty and default values
	if data == "" || (field == "event" && data == "message") {
		return
	}

	// If data has line breaks, the `field` must be prefixed to each line
	hasLineBreaks := strings.ContainsAny(data, "\n\r")
	if !hasLineBreaks {
		buf.WriteString(field)
		buf.WriteString(": ")
		buf.WriteString(data)
		buf.WriteString("\n")
		return
	}

	// Split data and add each line as a separate field
	start, dl := 0, len(data)
	for i, r := range data {
		if r == 13 || r == 10 {
			writeField(buf, field, data[start:i])
			start = i + 1
		}
		if i+1 == dl {
			writeField(buf, field, data[start:i+1])
		}
	}
}

func (w *Writer) SendEvent(event, data, id, retry string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var buf bytes.Buffer
	writeField(&buf, "event", event)
	writeField(&buf, "data", data)
	writeField(&buf, "id", id)
	writeField(&buf, "retry", retry)

	if buf.Len() == 0 {
		return nil
	}

	buf.WriteString("\n")
	bw, err := w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if bw != buf.Len() {
		return fmt.Errorf("%w: wrote %d bytes, expected %d", ErrWrite, bw, buf.Len())
	}

	w.flush()
	return nil
}

func (w *Writer) SendJSON(id, event string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return w.SendEvent(event, string(b), id, "")
}
