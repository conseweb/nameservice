package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/conseweb/common/assets/lepuscoin/client"
	"github.com/go-martini/martini"
	// "github.com/hyperledger/fabric/peer/common"
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
	log.Debugf("proxy chaincode to %s", req.URL.Path)
	cw := &ChaincodeWrapper{}
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

// invoke chaincode coinbase
func GetCoinBaseTx(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, par martini.Params) {
	addr := par["addr"]

	if addr == "" {
		ctx.Error(400, fmt.Errorf("invoke coinbase transaction failed, address is nil"))
		return
	}
	cli := client.NewTransactionV1("")
	cli.AddTxOut(client.NewTxOut(1000000, addr, 0))
	cli.AddTxIn(client.NewTxIn("", math.MaxUint32))
	bs, err := cli.Base64Bytes()
	if err != nil {
		log.Errorf("encode base64 failed, error: %s", err.Error())
		ctx.Error(400, err)
		return
	}
	log.Debugf("conbase tx: %s", bs)
	ctx.rnd.JSON(200, map[string]string{"message": string(bs)})
}

///
func GetTx(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, par martini.Params) {
	var tx struct {
		Founder    string `json:"founder"`
		ChargeAddr string `json:"charge_addr"`
		In         []struct {
			Addr       string `json:"addr"`
			PreTxHash  string `json:"pre_tx_hash"`
			TxOutIndex uint32 `json:"tx_out_index"`
		} `json:"in"`
		Out []struct {
			Addr   string `json:"addr"`
			Amount uint64 `json:"amount"`
			Until  int64  `json:"until"`
		} `json:"out"`
	}
	err := json.NewDecoder(req.Body).Decode(&tx)
	if err != nil {
		log.Error(err)
		ctx.Error(400, err)
		return
	}
	if len(tx.Out) == 0 {
		ctx.Error(400, fmt.Errorf("At least one out_addr is required"))
		return
	}

	txCli := client.NewTransactionV1(tx.Founder)
	for _, in := range tx.In {
		txCli.AddTxIn(client.NewTxIn(in.PreTxHash, in.TxOutIndex))
	}
	for _, out := range tx.Out {
		txCli.AddTxOut(client.NewTxOut(out.Amount, out.Addr, out.Until))
	}

	bs, err := txCli.Base64Bytes()
	if err != nil {
		log.Errorf("encode base64 failed, error: %s", err.Error())
		ctx.Error(400, err)
		return
	}
	log.Debugf("conbase tx: %s", bs)
	ctx.rnd.JSON(200, map[string]string{"message": string(bs)})
}
