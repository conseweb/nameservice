package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-martini/martini"
	ccpkg "github.com/hyperledger/fabric/peer/chaincode"
)

func DeployCC(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ccpkg.ChaincodeWrapper{}
	err := json.NewDecoder(ctx.req.Body).Decode(cw)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ret, err := cw.Deploy()
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ctx.rnd.JSON(201, map[string]interface{}{"message": string(ret)})
}

func InvokeCC(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ccpkg.ChaincodeWrapper{}
	err := json.NewDecoder(ctx.req.Body).Decode(cw)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ret, err := cw.Invoke()
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ctx.rnd.JSON(200, map[string]interface{}{"message": string(ret)})
}

func QueryCC(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ccpkg.ChaincodeWrapper{}
	err := json.NewDecoder(ctx.req.Body).Decode(cw)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ret, err := cw.Query()
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ctx.rnd.JSON(200, map[string]interface{}{"message": string(ret)})
}

func SetChaincode(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	alias := params["alias"]
	if alias == "" {
		ctx.Error(400, fmt.Errorf("alias is invalied."))
		return
	}

	cc := &ccpkg.ChaincodeWrapper{}
	err := json.NewDecoder(ctx.req.Body).Decode(cc)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	chaincodeManager.Lock()
	defer chaincodeManager.Unlock()
	c, ok := chaincodeManager.ccs[alias]
	if !ok {
		chaincodeManager.ccs[alias] = cc
	} else {
		if cc.Name != "" {
			c.Name = cc.Name
		}
		if cc.Path != "" {
			cc.Path = cc.Path
		}
	}

	ctx.rnd.JSON(201, map[string]string{"message": "successful"})
	return
}

func ListChaincodes(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	chaincodeManager.Lock()
	defer chaincodeManager.Unlock()
	ctx.rnd.JSON(200, chaincodeManager.ccs)
}

func GetChaincode(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	alias := params["alias"]

	chaincodeManager.Lock()
	defer chaincodeManager.Unlock()
	cc, ok := chaincodeManager.ccs[alias]
	if !ok {
		ctx.Error(404, fmt.Errorf("not found"))
		return
	}
	if cc.Name == "" {
		ctx.Error(404, fmt.Errorf("not deploy chaincode"))
		return
	}
	ctx.rnd.JSON(200, cc)
}
