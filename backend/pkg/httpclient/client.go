package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxErrorBodySize = 4 << 10

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("http status %d", e.StatusCode)
	}

	return fmt.Sprintf("http status %d: %s", e.StatusCode, e.Body)
}

type Options struct {
	Headers http.Header
}

func Do(ctx context.Context, client *http.Client, method string,
	endpoint string, opts Options,
) error {
	resp, err := send(ctx, client, method, endpoint, nil, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

func SendJSON[Req any](ctx context.Context, client *http.Client, method string,
	endpoint string, request Req, opts Options,
) error {
	body, err := marshalJSON(request)
	if err != nil {
		return err
	}

	resp, err := send(ctx, client, method, endpoint, body, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

func FetchJSON[Resp any](ctx context.Context, client *http.Client, method string,
	endpoint string, opts Options,
) (*Resp, error) {
	resp, err := send(ctx, client, method, endpoint, nil, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return decodeJSON[Resp](resp.Body)
}

func SendAndFetchJSON[Req, Resp any](ctx context.Context, client *http.Client, method string,
	endpoint string, request Req, opts Options,
) (*Resp, error) {
	body, err := marshalJSON(request)
	if err != nil {
		return nil, err
	}

	resp, err := send(ctx, client, method, endpoint, body, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return decodeJSON[Resp](resp.Body)
}

func send(ctx context.Context, client *http.Client, method string,
	endpoint string, body io.Reader, opts Options,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for name, values := range opts.Headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return resp, nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	data, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))

	return &HTTPError{
		StatusCode: resp.StatusCode,
		Body:       strings.TrimSpace(string(data)),
	}
}

func marshalJSON(value any) (io.Reader, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	return bytes.NewReader(data), nil
}

func decodeJSON[T any](body io.Reader) (*T, error) {
	var result T

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func NormalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	addr = strings.TrimRight(addr, "/")

	if addr == "" {
		return ""
	}

	if strings.HasPrefix(addr, "http://") ||
		strings.HasPrefix(addr, "https://") {
		return addr
	}

	return "http://" + addr
}
