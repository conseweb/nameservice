package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/conseweb/common/assets/lepuscoin/client"
	pb "github.com/conseweb/common/assets/lepuscoin/protos"
	"github.com/go-martini/martini"
	ccpkg "github.com/hyperledger/fabric/peer/chaincode"
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

func queryLepuscoinAddrs(cc *ccpkg.ChaincodeWrapper, addrs ...string) (*pb.QueryAddrResults, error) {
	if cc == nil || len(addrs) == 0 {
		return nil, fmt.Errorf("Parameter error")
	}

	cc.Args = addrs

	ret, err := cc.Query()
	if err != nil {
		log.Errorf("query failed, ", err)
		return nil, err
	}
	log.Debugf("query lepuscoin addrs: %s", ret)

	ar := &pb.QueryAddrResults{}
	err = json.Unmarshal(ret, ar)
	if err != nil {
		log.Errorf("decode query'e result failed, body: %s, error:", ret, err)
		return nil, err
	}

	return ar, nil
}

//"{"accounts":{"mtCLPxw18uxFMK1tbWLCVxJa4Tby7My7aM":
//{"addr":"mtCLPxw18uxFMK1tbWLCVxJa4Tby7My7aM","balance":99999990,
//"txouts":{"f7d9424e5c025e92c029ddc345336e7ec4a72ebbda2b615277e66eb8ec67878e:1":{"value":99999990,"addr":"mtCLPxw18uxFMK1tbWLCVxJa4Tby7My7aM"}}}}}"
func getTxIn(cc *ccpkg.ChaincodeWrapper, addrs ...string) ([]txIn, error) {
	qar, err := queryLepuscoinAddrs(cc, addrs...)
	if err != nil {
		log.Errorf("get tx in failed,", err)
		return nil, err
	}

	retIns := []txIn{}
	for _, v := range qar.GetAccounts() {
		for phi, advl := range v.GetTxouts() {
			phis := strings.Split(phi, ":")
			in := txIn{
				Addr:       advl.Addr,
				PreTxHash:  phis[0],
				TxOutIndex: phis[1],
				Balance:    advl.Value,
			}
			retIns = append(retIns, in)
		}
	}
	return retIns, err
}

func (tx *txWrapper) Serialized() ([]byte, error) {
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
	lcc, err := ccManager.Get("lepuscoin", "deploy", "deploy")
	if err == ErrNotDeploy {
		name, err := lcc.Deploy()
		if err != nil {
			ctx.Error(500, err)
			return
		}

		if name != "" {
			ccManager.SetName("lepuscoin", name)
			lcc.Name = name // for return frontend
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
// body:
// {
// 	"in": [{
// 		"addr": "from xxx..."
// 		}]
// 	"out": [{
// 		"addr": "to addr",
// 		"amount": 100xxx
// 		}]
// }
func Transfer(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	var tranf txWrapper
	err := json.NewDecoder(ctx.req.Body).Decode(&tranf)
	if err != nil {
		ctx.Error(400, err)
		return
	}
	if len(tranf.In) == 0 {
		ctx.Error(400, fmt.Errorf("At least one in address account is required"))
		return
	}

	addrs := make([]string, 0, len(tranf.In))
	for _, v := range tranf.In {
		addrs = append(addrs, v.Addr)
	}

	// get tx's in
	qAddrCc, err := ccManager.Get("lepuscoin", "query", "query_addrs")
	if err != nil {
		ctx.Error(500, err)
		return
	}
	in, err := getTxIn(qAddrCc, addrs...)
	if err != nil {
		log.Errorf("got tx in failed, %v", err)
		ctx.Error(500, err)
		return
	}

	tx := &txWrapper{
		In:  in,
		Out: tranf.Out,
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

// GET /lepuscoin/balance?addrs=[....]
func QueryAddrs(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	param := ctx.params["addrs"]
	if len(param) == 0 {
		ctx.Error(400, "need addrs")
		return
	}
	addrs := strings.Split(param, ",")

	qAddrCc, err := ccManager.Get("lepuscoin", "query", "query_addrs")
	if err != nil {
		ctx.Error(500, err)
		return
	}
	ret, err := queryLepuscoinAddrs(qAddrCc, addrs...)
	if err != nil {
		log.Errorf("got tx in failed, %v", err)
		ctx.Error(500, err)
		return
	}

	ctx.rnd.JSON(200, ret)
}
