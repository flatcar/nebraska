package updater

import (
	"bytes"
	"encoding/xml"
	"io"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kinvolk/go-omaha/omaha"
)

type HttpOmahaReqHandler struct {
	url        string
	httpClient *retryablehttp.Client
}

func NewHttpOmahaReqHandler(url string) *HttpOmahaReqHandler {
	return &HttpOmahaReqHandler{
		url,
		retryablehttp.NewClient(),
	}
}

func (h *HttpOmahaReqHandler) Handle(req *omaha.Request) (*omaha.Response, error) {
	requestByte, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Post(h.url, "text/xml", bytes.NewReader(requestByte))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// A response over 1M in size is certainly bogus.
	respBody := &io.LimitedReader{R: resp.Body, N: 1024 * 1024}
	contentType := resp.Header.Get("Content-Type")
	omahaResp, err := omaha.ParseResponse(contentType, respBody)
	if err != nil {
		return nil, err
	}

	return omahaResp, nil
}
