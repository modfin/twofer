package ndjson

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/modfin/twofer/stream"
)

// Writer implement a NDJSON stream writer used to send streamed JSON objects over a HTTP stream.
type Writer struct {
	mu      sync.Mutex
	encoder *json.Encoder
	flush   func()
}

var _ stream.Writer = (*Writer)(nil) // Compile-time check that we implement stream.Writer interface

func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flush := func() {} // NOP flusher, used if the provided writer don't support the http.Flusher interface
	f, ok := w.(http.Flusher)
	if ok {
		// Replace NOP flusher since provided writer support the http.Flusher interface
		flush = f.Flush
	}

	w.Header().Set("Content-Type", "application/x-json-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	return &Writer{encoder: json.NewEncoder(w), flush: flush}, nil
}

func NewEncoder(w http.ResponseWriter) (stream.Encoder, error) {
	enc, err := NewWriter(w)
	if err != nil {
		return nil, err
	}
	return enc.SendJSON, nil
}

func (w *Writer) SendJSON(_, _ string, data any) error {
	if data == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	defer w.flush()
	return w.encoder.Encode(data)
}
