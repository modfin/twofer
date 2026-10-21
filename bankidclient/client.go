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

// BankIdV3Client calls twofer's BankID V3 API: /bankid/v6/authv3, signv3,
// collectV3 and cancelV3.
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
	return c.startStream(ctx, "authv3", req)
}

// AuthV3Once starts an auth order and returns its first frame only, e.g. for
// an order opened in the BankID app on the same device rather than scanned.
func (c *BankIdV3Client) AuthV3Once(ctx context.Context, req api.BankIdv6AuthSignRequestV3) (api.BankIdV6AuthSignResponseV3, error) {
	return c.startOnce(ctx, "authv3", req)
}

// SignV3 starts a sign order, streamed like AuthV3. BankID requires
// UserVisibleData, the text the user signs.
func (c *BankIdV3Client) SignV3(ctx context.Context, req api.BankIdv6AuthSignRequestV3) (*AuthV3Stream, error) {
	return c.startStream(ctx, "signv3", req)
}

// SignV3Once starts a sign order and returns its first frame only, like
// AuthV3Once.
func (c *BankIdV3Client) SignV3Once(ctx context.Context, req api.BankIdv6AuthSignRequestV3) (api.BankIdV6AuthSignResponseV3, error) {
	return c.startOnce(ctx, "signv3", req)
}

func (c *BankIdV3Client) startStream(ctx context.Context, endpoint string, req api.BankIdv6AuthSignRequestV3) (*AuthV3Stream, error) {
	req.Once = false
	resp, err := c.post(ctx, endpoint, req)
	if err != nil {
		return nil, err
	}
	return newAuthV3Stream(endpoint, resp), nil
}

func (c *BankIdV3Client) startOnce(ctx context.Context, endpoint string, req api.BankIdv6AuthSignRequestV3) (api.BankIdV6AuthSignResponseV3, error) {
	req.Once = true
	resp, err := c.post(ctx, endpoint, req)
	if err != nil {
		return api.BankIdV6AuthSignResponseV3{}, err
	}
	defer resp.Body.Close()

	var ar api.BankIdV6AuthSignResponseV3
	err = json.NewDecoder(resp.Body).Decode(&ar)
	if err != nil {
		return api.BankIdV6AuthSignResponseV3{}, fmt.Errorf("err decode twofer %s response: %w", endpoint, err)
	}
	return ar, nil
}

// CollectV3 returns the order's status. WaitForChange holds the request until
// the status or the hintCode changes, so the reply can still be pending, e.g.
// with hintCode userSign once the QR is scanned. WaitUntilFinished holds it
// until the order is complete or failed.
//
// BankID hands out a finished order once. twofer finishes a poll even when the
// request is cut and keeps a finished order for a minute, so a repeated collect
// gets it, but older twofer servers lose it. Bound a wait with MaxWait rather
// than ctx, and give ctx headroom for twofer's last BankID poll.
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

// AuthV3Stream reads the QR frames of an AuthV3 or SignV3 order. twofer sends
// them as SSE or NDJSON depending on its STREAM_ENCODER setting.
type AuthV3Stream struct {
	endpoint string
	body     io.ReadCloser
	next     func() ([]byte, error)
}

func newAuthV3Stream(endpoint string, resp *http.Response) *AuthV3Stream {
	s := &AuthV3Stream{endpoint: endpoint, body: resp.Body, next: ndjsonData(json.NewDecoder(resp.Body))}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		s.next = sseData(bufio.NewReader(resp.Body))
	}
	return s
}

// Next returns the next frame, or io.EOF once twofer has ended the stream. A
// stream that breaks off returns another error. Frames without a QR code are
// skipped, so a newer twofer can add other frames.
func (s *AuthV3Stream) Next() (api.BankIdV6AuthSignResponseV3, error) {
	for {
		b, err := s.next()
		if errors.Is(err, io.EOF) {
			return api.BankIdV6AuthSignResponseV3{}, io.EOF
		}
		if err != nil {
			return api.BankIdV6AuthSignResponseV3{}, fmt.Errorf("err read twofer %s stream: %w", s.endpoint, err)
		}

		var ar api.BankIdV6AuthSignResponseV3
		err = json.Unmarshal(b, &ar)
		if err != nil {
			return api.BankIdV6AuthSignResponseV3{}, fmt.Errorf("err decode twofer %s frame: %w", s.endpoint, err)
		}
		if ar.QR != "" {
			return ar, nil
		}
	}
}

func (s *AuthV3Stream) Close() error {
	err := s.body.Close()
	if err != nil {
		return fmt.Errorf("err close twofer %s stream: %w", s.endpoint, err)
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
			// an event cut off by the end of the body is a broken stream, not its end
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("err read twofer sse line: %w", io.ErrUnexpectedEOF)
			}
			if err != nil {
				return nil, fmt.Errorf("err read twofer sse line: %w", err)
			}
			line = strings.TrimRight(line, "\r\n")

			if line == "" && len(data) > 0 {
				return []byte(strings.Join(data, "\n")), nil
			}
			if line == "" {
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
