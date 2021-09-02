package updater

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kinvolk/go-omaha/omaha"
)

type httpOmahaReqHandler struct {
	url        string
	httpClient *retryablehttp.Client
}

// NewDefaultOmahaRequestHandler returns a OmahaRequestHandler which uses default retryable http client
// to handle the omaha request.
func NewDefaultOmahaRequestHandler(url string) OmahaRequestHandler {
	return NewHttpOmahaRequestHandler(url, retryablehttp.NewClient())
}

// NewHttpOmahaRequestHandler returns a OmahaRequestHandler which uses the provided retryable http client
// to handle the omaha request.
func NewHttpOmahaRequestHandler(url string, client *retryablehttp.Client) OmahaRequestHandler {
	return &httpOmahaReqHandler{
		url:        url,
		httpClient: client,
	}
}

// Handle uses the httpClient to process the omaha request and returns omaha response
// and error.
func (h *httpOmahaReqHandler) Handle(req *omaha.Request) (*omaha.Response, error) {
	requestByte, err := xml.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding request as XML: %w", err)
	}

	resp, err := h.httpClient.Post(h.url, "text/xml", bytes.NewReader(requestByte))
	if err != nil {
		return nil, fmt.Errorf("http post request: %w", err)
	}
	defer resp.Body.Close()

	// A response over 1M in size is certainly bogus.
	respBody := &io.LimitedReader{R: resp.Body, N: 1024 * 1024}
	contentType := resp.Header.Get("Content-Type")
	omahaResp, err := omaha.ParseResponse(contentType, respBody)
	if err != nil {
		return nil, fmt.Errorf("parse response to omaha response: %w", err)
	}

	return omahaResp, nil
}
