package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ccpkg "github.com/hyperledger/fabric/peer/chaincode"
)

func GetClient() *http.Client {
	if proxyClient != nil {
		return proxyClient
	}
	return http.DefaultClient
}

func ProxyTo(way string, to *url.URL) func(http.ResponseWriter, *http.Request, *RequestContext) {
	return func(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
		cli := GetClient()

		toPath := strings.TrimPrefix(req.URL.Path, way)
		log.Debugf("proxy %s -> %s", req.URL.Path, toPath)
		req.RequestURI = ""
		req.URL.Host = to.Host
		req.URL.Scheme = to.Scheme
		req.URL.Path = to.Path + toPath
		req.Close = true
		req.Header.Set("Connection", "close")

		log.Debugf("to %s", req.URL.String())

		resp, err := cli.Do(req)
		if err != nil {
			ctx.Error(400, err)
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			if strings.ToLower(k) == "content-length" {
				continue
			}
			for _, v := range vv {
				rw.Header().Set(k, v)
			}
		}

		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)
	}
}

func ProxyFabric(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	log.Debug("...")
	log.Debugf("proxy to %s", req.URL.Path)
	cli := GetClient()

	req.RequestURI = ""

	req.URL.Host = daemon.GetRESTAddr()
	req.URL.Scheme = "http"
	req.URL.Path = strings.TrimPrefix(req.URL.Path, API_PREFIX)
	req.Close = true
	req.Header.Set("Connection", "close")

	resp, err := cli.Do(req)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	for k, vv := range resp.Header {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		for _, v := range vv {
			rw.Header().Set(k, v)
		}
	}

	rw.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := ioutil.ReadAll(resp.Body)
		ctx.Error(resp.StatusCode, fmt.Errorf("%s", msg))
	} else {
		io.Copy(rw, resp.Body)
	}
}

func ProxyChaincode(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	log.Debugf("proxy chaincode to %s", req.URL.Path)
	cw := &ccpkg.ChaincodeWrapper{}
	err := json.NewDecoder(req.Body).Decode(cw)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	bs, err := json.Marshal(cw.ToJSONRPC())
	if err != nil {
		ctx.Error(401, err)
		return
	}

	buf := &bytes.Buffer{}
	n, err := buf.Write(bs)
	if err != nil {
		ctx.Error(401, err)
		return
	}
	req.ContentLength = int64(n)
	req.Header.Set("Content-Length", strconv.Itoa(n))
	log.Debugf("set new Content-Length %v", n)

	req.Body = ioutil.NopCloser(buf)
	ctx.mc.Next()
}
