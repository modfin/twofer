package sse

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

type reader struct {
	//	buf *bufio.Scanner
}

func NewReader(ctx context.Context, rc io.ReadCloser) <-chan Event {
	eventChan := make(chan Event)
	//r := reader{
	//	buf: bufio.NewScanner(rc),
	//}
	go process(ctx, rc, eventChan)
	return eventChan
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Handle messages that use 'CRLF' or 'LF' line endings.
	advance, token, err = bufio.ScanLines(data, atEOF)
	if advance != 0 || len(data) == 0 {
		return
	}

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
		if err != nil {
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
			event.Event += fieldData
		case "data":
			event.Data += fieldData
		case "id":
			event.ID += fieldData
		case "retry":
			event.Retry += fieldData
		default:
			fmt.Printf("ignored unknoen field %s: %s\n", fieldName, fieldData)
		}
	}
}

func readField(s *bufio.Scanner) (string, string, bool, error) {
	if !s.Scan() {
		return "", "", false, s.Err()
	}

	line := s.Text()
	if line == "" {
		return "", "", true, nil // End of event
	}

	colonPos := strings.Index(line, ":")
	if colonPos <= 1 {
		return "", "", false, nil // Not a field, ignore
	}

	return line[0:colonPos], strings.TrimSpace(line[colonPos+1:]), false, nil
}
