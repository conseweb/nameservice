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

///
func NewTx(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, par martini.Params) {
	var tx struct {
		Founder    string `json:"founder"`
		ChargeAddr string `json:"charge_addr"`
		In         []struct {
			Addr       string `json:"addr"`
			PreTxHash  string `json:"pre_tx_hash"`
			TxOutIndex string `json:"tx_out_index"`
			Balance    uint64 `json:"balance"`
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
			ctx.Error(404, err)
			return
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
			ctx.Error(404, fmt.Errorf("Insufficient balance"))
			return
		} else if inAmount > outAmount {
			charge := tx.ChargeAddr
			if charge == "" {
				// use the first in addr to be charge address.
				charge = tx.In[0].Addr
			}
			txCli.AddTxOut(client.NewTxOut(inAmount-outAmount, charge, time.Time{}))
		}
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
