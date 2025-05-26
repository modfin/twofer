package ndjson

import (
	"net/http"
	"testing"
)

type (
	testWriter struct {
		flushed bool
		header  http.Header
		status  int
		written []byte
	}
	testWriteObj struct {
		Data string `json:"data,omitempty"`
	}
)

func (w *testWriter) Flush() {
	w.flushed = true
}

func (w *testWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *testWriter) Write(data []byte) (int, error) {
	w.written = append(w.written, data...)
	return len(data), nil
}

func (w *testWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func TestWriter_SendEvent(t *testing.T) {
	tests := []struct {
		name      string
		data      any
		wantErr   bool
		wantFlush bool
		wantData  string
	}{
		{
			name:      "all fields",
			data:      testWriteObj{Data: "test 1"},
			wantFlush: true,
			wantData:  `{"data":"test 1"}` + "\n",
		},
		{
			name:      "empty object",
			data:      testWriteObj{},
			wantFlush: true,
			wantData:  `{}` + "\n",
		},
		{
			name:      "nil object", // This should not send anything
			wantFlush: false,
			wantData:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := &testWriter{}

			w, err := NewWriter(tw)
			if err != nil {
				t.Fatalf("Failed to create new writer: %v", err)
			}

			err = w.SendJSON("", "", tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SendJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tw.flushed != tt.wantFlush {
				t.Errorf("got flush: %v, want: %v", tw.flushed, tt.wantFlush)
			}

			if string(tw.written) != tt.wantData {
				t.Errorf("got data: '%s', want: '%s'", tw.written, tt.wantData)
			}
		})
	}
}
