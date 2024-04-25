package ndjson

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"
)

func NewReader[Response any](ctx context.Context, rc io.ReadCloser) <-chan Response {
	eventChan := make(chan Response)
	go func() {
		defer close(eventChan)
		dec := json.NewDecoder(rc)
		for {
			var data Response
			err := dec.Decode(&data)
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				// TODO: Log error
				return
			}

			select {
			case eventChan <- data:
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				// TODO: Log that we have skipped to send 'data' (channel timeout)
			}
		}
	}()
	return eventChan
}
