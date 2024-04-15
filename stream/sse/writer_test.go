package sse

import (
	"bytes"
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

func Test_writeField(t *testing.T) {
	tests := []struct {
		name  string
		field string
		data  string
		want  string
	}{
		{
			name:  "vanilla",
			field: "data",
			data:  "vanilla",
			want:  "data:vanilla\n",
		},
		{
			name:  "multi-line",
			field: "data",
			data:  "multi\nline",
			want:  "data:multi\ndata:line\n",
		},
		{
			name:  "empty_data",
			field: "data",
			data:  "",
			want:  "",
		},
		{
			name:  "default_event_type",
			field: "event",
			data:  "message",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeField(&buf, tt.field, tt.data)
			got := string(buf.Bytes())
			if got != tt.want {
				t.Errorf("got: '%s', want: '%s", got, tt.want)
			}
		})
	}
}

func TestWriter_SendEvent(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		data      string
		id        string
		retry     string
		wantErr   bool
		wantFlush bool
		wantData  string
	}{
		{
			name:      "all fields",
			event:     "e1",
			data:      "d1",
			id:        "1",
			retry:     "1",
			wantFlush: true,
			wantData:  "event:e1\ndata:d1\nid:1\nretry:1\n\n",
		},
		{
			name:      "multiline_data",
			data:      "multi\nline\ndata",
			wantFlush: true,
			wantData:  "data:multi\ndata:line\ndata:data\n\n",
		},
		{
			name:      "only_default_event_type", // This should not send anything
			event:     "message",
			wantFlush: false,
			wantData:  "",
		},
		{
			name:      "only_non_default_event_type",
			event:     "ping",
			wantFlush: true,
			wantData:  "event:ping\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := &testWriter{}

			w, err := NewWriter(tw)
			if err != nil {
				t.Fatalf("Failed to create new writer: %v", err)
			}

			err = w.SendEvent(tt.event, tt.data, tt.id, tt.retry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SendEvent() error = %v, wantErr %v", err, tt.wantErr)
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
