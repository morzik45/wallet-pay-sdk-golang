package walletpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	apiTokenHeaderName = "Wpay-Store-Api-Key"
	apiUrl             = "https://pay.wallet.tg"
	orderPath          = "/wpay/store-api/v1/order"
	previewPath        = "/wpay/store-api/v1/order/preview"
)

// Client is a WalletPay client for making WalletPay API requests
type Client struct {
	ApiKey string

	httpClient *http.Client
}

type Options struct {
	ApiKey        string        // Required
	ClientTimeout time.Duration // Optional. Default: 20 seconds
}

// NewClient returns a new WalletPay API client
func NewClient(options Options) *Client {
	c := &Client{
		ApiKey: options.ApiKey,
	}
	clientTimeout := time.Second * 20
	if options.ClientTimeout != 0 {
		clientTimeout = options.ClientTimeout
	}
	c.httpClient = &http.Client{
		Timeout: clientTimeout,
	}
	return c
}

func (c *Client) getRequestUrl(path string) (rUrl string) {
	rUrl, _ = url.JoinPath(apiUrl, path)
	return
}

func (c *Client) encodeRequest(order OrderRequest) (io.Reader, error) {
	orderJson, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("error encoding order request: %w", err)
	}
	return bytes.NewReader(orderJson), nil
}

func (c *Client) decodeResponse(responseBodyReader io.Reader, targetPointer any) error {
	responseBody, err := io.ReadAll(responseBodyReader)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err = json.Unmarshal(responseBody, targetPointer); err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}
	return nil
}

func (c *Client) request(ctx context.Context, path string, body io.Reader, queryModifierFunc func(q url.Values) url.Values) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	if queryModifierFunc != nil {
		req.URL.RawQuery = queryModifierFunc(req.URL.Query()).Encode()
	}

	req.Header.Set(apiTokenHeaderName, c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}
