package ndjson

import (
	"context"
	"io"
	"testing"
	"time"
)

func TestNewReader(t *testing.T) {
	type testReadObj struct {
		Data string `json:"data"`
	}
	type send struct {
		buffer     string
		wantEvents []testReadObj
	}
	tests := []struct {
		name  string
		sends []send
	}{
		{
			name: "single message",
			sends: []send{
				{buffer: "{\"data\":\"test1-1\"}\n", wantEvents: []testReadObj{{"test1-1"}}},
				{buffer: "{\"data\":\"test1-2\"}\n", wantEvents: []testReadObj{{"test1-2"}}},
				{buffer: "{\"data\":\"test1-3\"}\n", wantEvents: []testReadObj{{"test1-3"}}},
			},
		},
		{
			name: "concatenated messages",
			sends: []send{
				{buffer: "{\"data\":\"test2-1\"}\n{\"data\":\"test2-2\"}\n{\"data\":\"test2-3\"}\n", wantEvents: []testReadObj{{"test2-1"}, {"test2-2"}, {"test2-3"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr, pw := io.Pipe()
			chn := NewReader[testReadObj](context.Background(), pr)

			for i, s := range tt.sends {
				if _, err := pw.Write([]byte(s.buffer)); err != nil {
					t.Errorf("write buffer #%d ('%s') failed with error: %v", i, s.buffer, err)
				}

				var gotEvents []testReadObj
			done:
				for {
					select {
					case e := <-chn:
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
		})
	}
}
