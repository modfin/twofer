package bankidclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/modfin/twofer/api"
)

// maxErrorBody caps the raw error reply kept on BankIdError, it ends up in logs.
const maxErrorBody = 4096

// BankIdV3Client calls twofer's BankID V3 API: /bankid/v6/authv3, collectV3 and
// cancelV3.
type BankIdV3Client struct {
	c       *http.Client
	baseUrl string
}

func NewBankIdV3Client(baseurl string) *BankIdV3Client {
	return &BankIdV3Client{
		// No Timeout: authv3 streams for ~30s and collectV3 long-polls.
		// Callers bound each request with ctx instead.
		c:       &http.Client{},
		baseUrl: strings.TrimRight(baseurl, "/"),
	}
}

// BankIdError is a non-2xx reply from twofer. Code carries BankID's errorCode
// when Origin is api.ErrorOriginBankIDv6. A reply that is not twofer's error
// JSON leaves the embedded fields empty and only Body set.
type BankIdError struct {
	api.BankIdv6ErrorResponseV3
	Endpoint   string
	HTTPStatus int
	// Body is the raw reply, capped at 4096 bytes
	Body string
}

func (e *BankIdError) Error() string {
	return fmt.Sprintf("twofer %s returned status %d: %s", e.Endpoint, e.HTTPStatus, e.Body)
}

func (c *BankIdV3Client) post(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("err marshal twofer %s request: %w", endpoint, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/bankid/v6/"+endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("err create twofer %s request: %w", endpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("err post twofer %s: %w", endpoint, err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return resp, nil
	}

	defer resp.Body.Close()
	errBody, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	if err != nil {
		return nil, fmt.Errorf("err read twofer %s status %d reply: %w", endpoint, resp.StatusCode, err)
	}
	be := &BankIdError{Endpoint: endpoint, HTTPStatus: resp.StatusCode, Body: string(errBody)}
	_ = json.Unmarshal(errBody, &be.BankIdv6ErrorResponseV3)
	return nil, be
}

// AuthV3 starts an auth order. twofer then streams a fresh QR frame each
// second for ~30 seconds. Close the stream when done with it.
func (c *BankIdV3Client) AuthV3(ctx context.Context, req api.BankIdv6AuthSignRequestV3) (*AuthV3Stream, error) {
	req.Once = false
	resp, err := c.post(ctx, "authv3", req)
	if err != nil {
		return nil, err
	}
	return newAuthV3Stream(resp), nil
}

// AuthV3Once starts an auth order and returns its first frame only, e.g. for
// an order opened in the BankID app on the same device rather than scanned.
func (c *BankIdV3Client) AuthV3Once(ctx context.Context, req api.BankIdv6AuthSignRequestV3) (api.BankIdV6AuthSignResponseV3, error) {
	req.Once = true
	resp, err := c.post(ctx, "authv3", req)
	if err != nil {
		return api.BankIdV6AuthSignResponseV3{}, err
	}
	defer resp.Body.Close()

	var ar api.BankIdV6AuthSignResponseV3
	err = json.NewDecoder(resp.Body).Decode(&ar)
	if err != nil {
		return api.BankIdV6AuthSignResponseV3{}, fmt.Errorf("err decode twofer authv3 response: %w", err)
	}
	return ar, nil
}

// CollectV3 returns the order's status. WaitForChange holds the request until
// the status or the hintCode changes, so the reply can still be pending, e.g.
// with hintCode userSign once the QR is scanned. WaitUntilFinished holds it
// until the order is complete or failed.
//
// BankID hands out a finished order once. A ctx that ends a wait just after
// twofer collected the result loses it, and the next collect gets
// invalidParameters "No such order".
func (c *BankIdV3Client) CollectV3(ctx context.Context, req api.BankIdv6CollectRequestV3) (api.BankIdV6CollectResponseV3, error) {
	resp, err := c.post(ctx, "collectV3", req)
	if err != nil {
		return api.BankIdV6CollectResponseV3{}, err
	}
	defer resp.Body.Close()

	var cr api.BankIdV6CollectResponseV3
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return api.BankIdV6CollectResponseV3{}, fmt.Errorf("err decode twofer collectV3 response: %w", err)
	}
	return cr, nil
}

func (c *BankIdV3Client) CancelV3(ctx context.Context, req api.BankIdv6CancelRequestV3) error {
	resp, err := c.post(ctx, "cancelV3", req)
	if err != nil {
		return err
	}
	// read to EOF so the connection goes back to the pool, Go before 1.27
	// closes it otherwise
	_, _ = io.Copy(io.Discard, resp.Body)
	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("err close twofer cancelV3 response: %w", err)
	}
	return nil
}

// AuthV3Stream reads the QR frames of an AuthV3 order. twofer sends them as SSE
// or NDJSON depending on its STREAM_ENCODER setting.
type AuthV3Stream struct {
	body   io.ReadCloser
	status int
	next   func() ([]byte, error)
}

func newAuthV3Stream(resp *http.Response) *AuthV3Stream {
	s := &AuthV3Stream{body: resp.Body, status: resp.StatusCode}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		s.next = sseData(bufio.NewReader(resp.Body))
		return s
	}
	s.next = ndjsonData(json.NewDecoder(resp.Body))
	return s
}

// Next returns the next frame, or io.EOF once twofer has ended the stream.
// A stream that breaks off returns another error, and twofer failing mid-stream
// returns a *BankIdError.
func (s *AuthV3Stream) Next() (api.BankIdV6AuthSignResponseV3, error) {
	b, err := s.next()
	if err != nil {
		return api.BankIdV6AuthSignResponseV3{}, err
	}

	var ar api.BankIdV6AuthSignResponseV3
	err = json.Unmarshal(b, &ar)
	if err != nil {
		return api.BankIdV6AuthSignResponseV3{}, fmt.Errorf("err decode twofer authv3 frame: %w", err)
	}
	if ar.QR != "" {
		return ar, nil
	}

	// Every frame carries a QR code. Older twofer appends an error reply to a
	// stream it fails to continue, which decodes to an empty frame.
	if len(b) > maxErrorBody {
		b = b[:maxErrorBody]
	}
	be := &BankIdError{Endpoint: "authv3", HTTPStatus: s.status, Body: string(b)}
	_ = json.Unmarshal(b, &be.BankIdv6ErrorResponseV3)
	return api.BankIdV6AuthSignResponseV3{}, be
}

func (s *AuthV3Stream) Close() error {
	err := s.body.Close()
	if err != nil {
		return fmt.Errorf("err close twofer authv3 stream: %w", err)
	}
	return nil
}

// sseData returns the data of each event. twofer only sends unnamed message
// events, so other fields are skipped.
func sseData(r *bufio.Reader) func() ([]byte, error) {
	return func() ([]byte, error) {
		var data []string
		for {
			line, err := r.ReadString('\n')
			if errors.Is(err, io.EOF) && len(data) == 0 && strings.TrimSpace(line) == "" {
				return nil, io.EOF
			}
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("err read twofer sse line: %w", err)
			}
			eof := err != nil
			line = strings.TrimRight(line, "\r\n")

			// not SSE: an error reply written onto the stream as raw JSON
			if len(data) == 0 && strings.HasPrefix(line, "{") {
				return []byte(line), nil
			}
			// an event cut off by the end of the body is a broken stream, not its end
			if eof {
				return nil, fmt.Errorf("err read twofer sse line: %w", io.ErrUnexpectedEOF)
			}

			if line == "" {
				if len(data) > 0 {
					return []byte(strings.Join(data, "\n")), nil
				}
				continue
			}
			field, value, _ := strings.Cut(line, ":")
			if field != "data" {
				continue
			}
			data = append(data, strings.TrimPrefix(value, " "))
		}
	}
}

func ndjsonData(dec *json.Decoder) func() ([]byte, error) {
	return func() ([]byte, error) {
		var raw json.RawMessage
		err := dec.Decode(&raw)
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		if err != nil {
			return nil, fmt.Errorf("err decode twofer ndjson frame: %w", err)
		}
		return raw, nil
	}
}
