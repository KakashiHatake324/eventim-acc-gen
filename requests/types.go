package requests

import (
	"context"
	"net/http"
)

type DoRequest struct {
	Client         *http.Client
	CTX            context.Context
	AcceptedStatus []int
	Req            map[string]string
	Headers        map[string][]string
}

type RequestResponse struct {
	RespStatus      int
	ResponseBody    string
	ResponseHeaders http.Header
	ResponseRequest *http.Request
	Error           error
}
