package request

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

type RequestOptions struct {
	Headers     map[string]string
	QueryParams map[string]string
	Timeout     time.Duration
	Context     context.Context
}

type Response struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	RawResponse *http.Response
}

func DefaultRequestOptions() *RequestOptions {
	return &RequestOptions{
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Timeout:     30 * time.Second,
		Context:     context.Background(),
	}
}

func Get(url string, options *RequestOptions) (*Response, error) {
	return makeRequest(http.MethodGet, url, nil, options)
}

func Post(url string, body interface{}, options *RequestOptions) (*Response, error) {
	return makeRequest(http.MethodPost, url, body, options)
}

func makeRequest(method, requestURL string, body interface{}, options *RequestOptions) (*Response, error) {
	if options == nil {
		options = DefaultRequestOptions()
	}

	var requestBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewReader(jsonBody)
	}

	finalURL, err := addQueryParams(requestURL, options.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to add query parameters: %w", err)
	}

	req, err := http.NewRequestWithContext(options.Context, method, finalURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Go-http-client/2.0")
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	client := InitClient()

	if options.Timeout > 0 {
		clientWithTimeout := *client
		clientWithTimeout.Timeout = options.Timeout
		client = &clientWithTimeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	encoding := resp.Header.Get("Content-Encoding")

	switch strings.ToLower(encoding) {
	case "br":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	responseBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode:  resp.StatusCode,
		Headers:     resp.Header,
		Body:        responseBody,
		RawResponse: resp,
	}, nil
}

func addQueryParams(baseURL string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	urlParse, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	query := urlParse.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	urlParse.RawQuery = query.Encode()

	return urlParse.String(), nil
}
