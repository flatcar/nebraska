package updater

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/kinvolk/go-omaha/omaha"
)

type HttpOmahaReqHandler struct {
	url        string
	debug      bool
	httpClient *retryablehttp.Client
}

func NewHttpOmahaReqHandler(url string, debug bool) *HttpOmahaReqHandler {
	return &HttpOmahaReqHandler{
		url,
		debug,
		retryablehttp.NewClient(),
	}
}

func (h *HttpOmahaReqHandler) Handle(req *omaha.Request) (*omaha.Response, error) {
	requestByte, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	if h.debug {
		fmt.Println("Raw Request:\n", string(requestByte))
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
	if h.debug {
		responseByte, err := xml.Marshal(omahaResp)
		if err == nil {
			fmt.Println("Raw Response:\n", string(responseByte))
		}
	}
	if err != nil {
		return nil, err
	}

	return omahaResp, nil
}
