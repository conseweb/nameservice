package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ChaincodeWrapper struct {
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	Method        string   `json:"method"`
	Function      string   `json:"function"`
	Args          []string `json:"args"`
	SecureContext string   `json:"secureContext"`
}

type ResultWrapper struct {
	Result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"result"`
}

type JSONRPCWrapper struct {
	ID      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  ParamsWrapper `json:"params"`
}

type ParamsWrapper struct {
	Type          int                    `json:"type"`
	ChaincodeID   map[string]string      `json:"chaincodeID"`
	CtorMsg       map[string]interface{} `json:"ctorMsg"`
	SecureContext string                 `json:"secureContext,omitempty"`
}

var (
	NewID = func(start int) func() int {
		id := start - 1
		return func() int {
			id++
			return id
		}
	}(1)
)

func (c *ChaincodeWrapper) ToJSONRPC() *JSONRPCWrapper {
	cc := &JSONRPCWrapper{
		ID:      NewID(),
		Jsonrpc: "2.0",
		Method:  c.Method,
		Params: ParamsWrapper{
			Type:        1,
			ChaincodeID: map[string]string{},
			CtorMsg: map[string]interface{}{
				"args": append([]string{c.Function}, c.Args...),
			},
		},
	}
	if c.Name != "" {
		cc.Params.ChaincodeID["name"] = c.Name
	} else if c.Path != "" {
		cc.Params.ChaincodeID["path"] = c.Path
	}
	if c.SecureContext != "" {
		cc.Params.SecureContext = c.SecureContext
	}

	return cc
}

func ProxyChaincode(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ChaincodeWrapper{}
	err := json.NewDecoder(req.Body).Decode(cw)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(cw.ToJSONRPC())
	if err != nil {
		ctx.Error(401, err)
		return
	}

	req.Body = ioutil.NopCloser(buf)
}
