package sse

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

func NewReader(ctx context.Context, rc io.ReadCloser) <-chan Event {
	eventChan := make(chan Event)
	go process(ctx, rc, eventChan)
	return eventChan
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Handle messages that use 'CRLF' or 'LF' line endings.
	advance, token, err = bufio.ScanLines(data, atEOF)
	if advance > 0 && len(token) > 0 {
		// Check if returned token contain CR
		if i := bytes.IndexByte(token, '\r'); i >= 0 {
			// We have a full CR-terminated line.
			return i + 1, token[0:i], nil
		}
	}
	if advance != 0 || len(data) == 0 {
		return
	}

	// Check if data contain CR
	// (this can potentially break things if CRLF is used but `data` ends with the CR because we haven't read further from the reader yet)
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full CR-terminated line.
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func readField(s *bufio.Scanner) (string, string, bool, error) {
	if !s.Scan() {
		if s.Err() == nil {
			return "", "", false, io.EOF
		}
		return "", "", false, s.Err()
	}

	line := s.Text()
	if line == "" {
		return "", "", true, nil // End of event
	}

	colonPos := strings.Index(line, ":")
	if colonPos <= 1 || (len(line) >= 1 && line[0] == '#') {
		return "", "", false, nil // Not a field, ignore
	}

	return line[0:colonPos], strings.TrimSpace(line[colonPos+1:]), false, nil
}

func process(ctx context.Context, rc io.ReadCloser, ec chan<- Event) {
	defer rc.Close()
	defer close(ec)
	buf := bufio.NewScanner(rc)
	buf.Split(scanLines)
	var event Event
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Continue to read data
		}
		fieldName, fieldData, eom, err := readField(buf)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			fmt.Printf("failed to read field, error: %v\n", err)
			return
		}

		if eom {
			if event.Event == "" {
				event.Event = "message"
			}
			ec <- event
			event = Event{}
			continue
		}

		switch fieldName {
		case "event":
			if event.Event != "" {
				event.Event += "\n"
			}
			event.Event += fieldData
		case "data":
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += fieldData
		case "id":
			if event.ID != "" {
				event.ID += "\n"
			}
			event.ID += fieldData
		case "retry":
			if event.Retry != "" {
				event.Retry += "\n"
			}
			event.Retry += fieldData
		default:
			fmt.Printf("ignored unknown field %s: %s\n", fieldName, fieldData)
		}
	}
}
