package sse

import (
	"bufio"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func Test_scanLines(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		eof         bool
		wantAdvance int
		wantToken   []byte
		wantErr     bool
	}{
		{
			name:        "lf_line",
			data:        []byte("token 1\ntoken 2\n"),
			wantAdvance: 8,
			wantToken:   []byte("token 1"),
		},
		{
			name:        "cr_line",
			data:        []byte("token 2\rtoken 3\r"),
			wantAdvance: 8,
			wantToken:   []byte("token 2"),
		},
		{
			name:        "crlf_line",
			data:        []byte("token 3\r\ntoken 4\r\n"),
			wantAdvance: 9,
			wantToken:   []byte("token 3"),
		},
		{
			name:        "mixed_line_line_endings_1",
			data:        []byte("token 4\ntoken 5\r\ntoken 6\r"),
			wantAdvance: 8,
			wantToken:   []byte("token 4"),
		},
		{
			name:        "mixed_line_line_endings_2",
			data:        []byte("token 5\r\ntoken 6\rtoken 7\n"),
			wantAdvance: 9,
			wantToken:   []byte("token 5"),
		},
		{
			name:        "mixed_line_line_endings_3",
			data:        []byte("token 6\rtoken 7\ntoken 8\r\n"),
			wantAdvance: 8,
			wantToken:   []byte("token 6"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdvance, gotToken, err := scanLines(tt.data, tt.eof)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error: %v, want: %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("got advance: %v, want: %v", gotAdvance, tt.wantAdvance)
			}
			if string(gotToken) != string(tt.wantToken) {
				t.Errorf("got token: %s, want %s", gotToken, tt.wantToken)
			}
		})
	}
}

func Test_readField(t *testing.T) {
	tests := []struct {
		name     string
		buffer   string
		wantType string
		wantData string
		wantEOM  bool
		wantErr  error
	}{
		{
			name:     "event",
			buffer:   "event: test\n",
			wantType: "event",
			wantData: "test",
			wantEOM:  false,
		},
		{
			name:     "comment",
			buffer:   "# comment\n",
			wantType: "",
			wantData: "",
			wantEOM:  false,
		},
		{
			name:     "empty_line",
			buffer:   "\n",
			wantType: "",
			wantData: "",
			wantEOM:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bufio.NewScanner(strings.NewReader(tt.buffer))
			buf.Split(scanLines)
			gotType, gotData, gotEOM, gotErr := readField(buf)
			if gotErr != tt.wantErr {
				t.Fatalf("got error: %v, want: %v", gotErr, tt.wantErr)
			}
			if gotType != tt.wantType {
				t.Errorf("got type: %s, want %s", gotType, tt.wantType)
			}
			if gotData != tt.wantData {
				t.Errorf("got data: %s, want %s", gotData, tt.wantData)
			}
			if gotEOM != tt.wantEOM {
				t.Errorf("got End-of-Message: %v, want %v", gotEOM, tt.wantEOM)
			}
		})
	}
}

func Test_process(t *testing.T) {
	type send struct {
		buffer     string
		wantEvents []Event
	}
	tests := []struct {
		name  string
		sends []send
	}{
		{
			name: "single event",
			sends: []send{
				{
					buffer:     "event:1\n\n",
					wantEvents: []Event{{Event: "1"}},
				},
			},
		},
		{
			name: "event_properties",
			sends: []send{
				{
					buffer:     "event:2\n\ndata:2\n\nid:2\n\nretry:2\n\n",
					wantEvents: []Event{{Event: "2"}, {Event: "message", Data: "2"}, {Event: "message", ID: "2"}, {Event: "message", Retry: "2"}},
				},
			},
		},
		{
			name: "multi_send",
			sends: []send{
				{
					buffer:     "event:31\n\n",
					wantEvents: []Event{{Event: "31"}},
				},
				{
					buffer:     "event:32\n\ndata:32\n\n",
					wantEvents: []Event{{Event: "32"}, {Event: "message", Data: "32"}},
				},
				{
					buffer:     "event:33\n\ndata:33\n\nid:33\n\n",
					wantEvents: []Event{{Event: "33"}, {Event: "message", Data: "33"}, {Event: "message", ID: "33"}},
				},
				{
					buffer:     "event:34\n\ndata:34\n\nid:34\n\nretry:34\n\n",
					wantEvents: []Event{{Event: "34"}, {Event: "message", Data: "34"}, {Event: "message", ID: "34"}, {Event: "message", Retry: "34"}},
				},
			},
		},
		{
			name: "cr_line_breaks",
			sends: []send{
				{
					buffer:     "event:41\r\revent:42\r\r",
					wantEvents: []Event{{Event: "41"}, {Event: "42"}},
				},
			},
		},
		{
			name: "crlf_line_breaks",
			sends: []send{
				{
					buffer:     "event:51\r\n\r\nevent:52\r\n\r\n",
					wantEvents: []Event{{Event: "51"}, {Event: "52"}},
				},
			},
		},
		{
			name: "incomplete_message",
			sends: []send{
				{
					buffer:     "no-line-break",
					wantEvents: []Event{},
				},
			},
		},
		{
			name: "message_without_supported_lines",
			sends: []send{
				{
					buffer:     "# Comment\nignored_ data\n\n",
					wantEvents: []Event{{Event: "message"}},
				},
			},
		},
		{
			name: "multi_line_message",
			sends: []send{
				{
					buffer:     "event: first line\nevent: second line\n\n",
					wantEvents: []Event{{Event: "first line\nsecond line"}},
				},
				{
					buffer:     "data: first line\ndata: second line\n\n",
					wantEvents: []Event{{Event: "message", Data: "first line\nsecond line"}},
				},
				{
					buffer:     "id: first line\nid: second line\n\n",
					wantEvents: []Event{{Event: "message", ID: "first line\nsecond line"}},
				},
				{
					buffer:     "retry: first line\nretry: second line\n\n",
					wantEvents: []Event{{Event: "message", Retry: "first line\nsecond line"}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr, pw := io.Pipe()
			eventChan := make(chan Event)
			var terminated bool
			go func() {
				t.Log("start async process...")
				process(context.Background(), pr, eventChan)
				terminated = true
				t.Log("async process have terminated...")
			}()
			for i, s := range tt.sends {
				_, err := pw.Write([]byte(s.buffer))
				if err != nil {
					t.Fatalf("failed to write buffer #%d, error: %v", i, err)
				}
				var gotEvents []Event
			done:
				for {
					select {
					case e := <-eventChan:
						gotEvents = append(gotEvents, e)
						t.Logf("got event: %#v", e)
					case <-time.After(time.Millisecond * 50):
						break done
					}
				}

				if len(gotEvents) != len(s.wantEvents) {
					t.Errorf("got %d events, want %d\nreceived events: %#v", len(gotEvents), len(s.wantEvents), gotEvents)
				} else {
					for i, we := range s.wantEvents {
						ge := gotEvents[i]
						if ge != we {
							t.Errorf("got event #%d: %#v, want: %#v", i, ge, we)
						}
					}
				}
			}

			// Close pipe writer to indicate that there won't be any more data to read for the async process
			err := pw.Close()
			if err != nil {
				t.Fatalf("failed to close pipe, error: %v", err)
			}

			// wait a short time so that the async process have time to shutdown after we've closed the reader
			time.Sleep(time.Millisecond * 100)

			// Check that the async process ended so that it's go-routine can exit cleanly
			if !terminated {
				t.Errorf("the async process didn't terminate after reader had been closed")
			}
		})
	}
}
