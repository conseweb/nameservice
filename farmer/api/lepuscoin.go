package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/conseweb/common/assets/lepuscoin/client"
	"github.com/go-martini/martini"
)

type txWrapper struct {
	Founder    string  `json:"founder"`
	ChargeAddr string  `json:"charge_addr"`
	In         []txIn  `json:"in"`
	Out        []txOut `json:"out"`
}

type txIn struct {
	Addr       string `json:"addr"`
	PreTxHash  string `json:"pre_tx_hash"`
	TxOutIndex string `json:"tx_out_index"`
	Balance    uint64 `json:"balance"`
}

type txOut struct {
	Addr   string `json:"addr"`
	Amount uint64 `json:"amount"`
	Until  int64  `json:"until"`
}

func (t *txWrapper) Serialized() ([]byte, error) {
	if len(tx.Out) == 0 {
		return nil, fmt.Errorf("At least one out_addr is required")
	}

	founder := tx.Founder
	if len(tx.In) > 0 && founder == "" {
		// use the first in addr to be founder.
		founder = tx.In[0].Addr
	}
	txCli := client.NewTransactionV1(founder)

	var inAmount, outAmount uint64
	for _, in := range tx.In {
		index, err := strconv.Atoi(in.TxOutIndex)
		if err != nil {
			return nil, err
		}
		txCli.AddTxIn(client.NewTxIn(in.Addr, in.PreTxHash, uint32(index)))
		inAmount += in.Balance
	}

	for _, out := range tx.Out {
		txCli.AddTxOut(client.NewTxOut(out.Amount, out.Addr, time.Unix(out.Until, 0)))
		outAmount += out.Amount
	}

	// not a coinbase transaction
	if len(tx.In) > 0 {
		if inAmount < outAmount {
			log.Debugf("inAmount: %v, outAmount: %v", inAmount, outAmount)
			return nil, fmt.Errorf("Insufficient balance")
		} else if inAmount > outAmount {
			charge := tx.ChargeAddr
			if charge == "" {
				// use the first in addr to be charge address.
				charge = tx.In[0].Addr
			}
			txCli.AddTxOut(client.NewTxOut(inAmount-outAmount, charge, time.Time{}))
		}
	}

	return txCli.Base64Bytes()
}

// invoke chaincode coinbase
func GetCoinBaseTx(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, par martini.Params) {
	addr := par["addr"]

	if addr == "" {
		ctx.Error(400, fmt.Errorf("invoke coinbase transaction failed, address is nil"))
		return
	}
	cli := client.NewTransactionV1("")
	cli.AddTxOut(client.NewTxOut(1000000, addr, time.Time{}))
	bs, err := cli.Base64Bytes()
	if err != nil {
		log.Errorf("encode base64 failed, error: %s", err.Error())
		ctx.Error(400, err)
		return
	}
	log.Debugf("conbase tx: %s", bs)
	ctx.rnd.JSON(200, map[string]string{"message": string(bs)})
}

/// POST /lepuscoin/tx
func NewTx(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, par martini.Params) {
	var tx txWrapper
	err := json.NewDecoder(req.Body).Decode(&tx)
	if err != nil {
		log.Error(err)
		ctx.Error(400, err)
		return
	}

	bs, err := tx.Serialized()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	log.Debugf("conbase tx: %s", bs)
	ctx.Message(200, string(bs))
}

// POST /lepuscoin/deploy
func DeployLepuscoin(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	lcc, err := ccManager.Get("lepuscoin")
	if err == ErrNotDeploy {
		name, err := lcc.Deploy()
		if err != nil {
			ctx.Error(500, err)
			return
		}

		if name != "" {
			ccManager.SetName("lepuscoin", name)
			log.Debugf("set lepuscoin chaincode name: %s", name)
		}
	} else if err != nil {
		ctx.Error(400, err)
		return
	}

	log.Debugf("return lepuscoin Chaincode name", lcc.Name)
	ctx.Message(201, lcc.Name)
}

// POST /lepuscoin/coinbase
func DoCoinbase(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	var body struct {
		Addrs []string `json:"addrs"`
	}

	err := json.NewDecoder(ctx.req.Body).Decode(&body)
	if err != nil {
		log.Error("decode body failed, ", err)
		ctx.Error(400, err)
		return
	}

	tx := &txWrapper{}
	for _, addr := range body.Addrs {
		tx.Out = append(tx.Out, txOut{
			Addr:   addr,
			Amount: 1000000,
		})
	}
	txbs, err := tx.Serialized()
	if err != nil {
		log.Errorf("serialized tx failed, %v", err)
		ctx.Error(500, err)
		return
	}

	cw, err := ccManager.Get("lepuscoin", "invoke", "invoke_coinbase", string(txbs))
	if err != nil {
		log.Errorf("lepuscoin's chaincode not deploy")
		ctx.Error(500, err)
		return
	}

	retbs, err := cw.Invoke()
	if err != nil {
		log.Errorf("invoke coinbase failed, %v", err)
		ctx.Error(500, err)
		return
	}

	ctx.Message(200, string(retbs))
}

// POST /lepuscoin/transfer
func Transfer(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	var tx txWrapper
	err := json.NewDecoder(ctx.req.Body).Decode(&tx)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	bs, err := tx.Serialized()
	if err != nil {
		log.Errorf("try decode transfer tx failed, %v", err)
		ctx.Error(500, err)
		return
	}

	cw, err := ccManager.Get("lepuscoin", "invoke", "invoke_transfer", string(bs))
	if err != nil {
		log.Errorf("lepuscoin's chaincode not deploy")
		ctx.Error(500, err)
		return
	}

	retbs, err := cw.Invoke()
	if err != nil {
		log.Errorf("invoke coinbase failed, %v", err)
		ctx.Error(500, err)
		return
	}

	ctx.Message(200, string(retbs))
}

// GET /lepuscoin/addrs?addrs=[....]
func QueryAddrs(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {

}
