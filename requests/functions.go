package requests

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

// make a request and return information
func (Request *DoRequest) MakeRequest() *RequestResponse {

	// If no context was found then use context background
	if Request.CTX == nil {
		Request.CTX = context.Background()
	}

	var SendData io.Reader

	if Request.Req["Data"] == "nil" {
		SendData = nil
	} else {
		SendData = strings.NewReader(Request.Req["Data"])
	}

	req, err := http.NewRequestWithContext(Request.CTX, Request.Req["Method"], Request.Req["URL"], SendData)
	if err != nil {
		return &RequestResponse{Error: err}
	}

	ArrangedHeaders := strings.Join(Request.Headers["header-order"], ",")

	for d := range Request.Headers {
		if len(Request.Headers[d]) == 0 {
			continue
		}

		if strings.ToLower(d) == "header-order" {
			continue
		} else if strings.ToLower(d) == "cookie" {
			req.Header.Add(d, Request.Headers[d][0])
		} else if strings.ToLower(d) == "host" {
			req.Host = Request.Headers[d][0]
		} else {
			req.Header.Set(d, Request.Headers[d][0])
		}

	}

	if ArrangedHeaders != "" {
		req.Header.Set("Header-Order", ArrangedHeaders)
	}

	resp, err := Request.Client.Do(req)
	if err != nil {
		return &RequestResponse{Error: err}
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RequestResponse{Error: err}
	}

	if strings.Join(resp.Header["Content-Encoding"], "") == "gzip" {
		rdata := strings.NewReader(string(bodyText))
		r, err := gzip.NewReader(rdata)
		if err != nil {
			return &RequestResponse{Error: err}
		}
		bodyText, err = io.ReadAll(r)
		if err != nil {
			return &RequestResponse{Error: err}
		}
	} else if strings.Join(resp.Header["Content-Encoding"], "") == "br" {
		rdata := strings.NewReader(string(bodyText))
		r := brotli.NewReader(rdata)
		bodyText, err = io.ReadAll(r)
		if err != nil {
			return &RequestResponse{Error: err}
		}
	}

	bodyString := string(bodyText)

	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	var statusAccepted bool
	for _, as := range Request.AcceptedStatus {
		if as == resp.StatusCode {
			statusAccepted = true
		}
	}
	if !statusAccepted {
		err = errors.New("status not accepted " + resp.Status)
	}
	return &RequestResponse{
		RespStatus:      resp.StatusCode,
		ResponseBody:    bodyString,
		ResponseHeaders: resp.Header,
		ResponseRequest: resp.Request,
		Error:           err,
	}
}
