package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-martini/martini"
)

// POST /nameservice/deploy
func DeployNameService(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	lcc, err := ccManager.Get("nameservice", "deploy", "deploy")
	if err == ErrNotDeploy {
		name, err := lcc.Deploy()
		if err != nil {
			ctx.Error(500, err)
			return
		}

		if name != "" {
			ccManager.SetName("nameservice", name)
			lcc.Name = name // for return frontend
			log.Debugf("set nameservice chaincode name: %s", name)
		}
	} else if err != nil {
		ctx.Error(400, err)
		return
	}

	log.Debugf("return nameservice Chaincode name", lcc.Name)
	ctx.Message(201, lcc.Name)
}

func NewNameServiceKV(ctx *RequestContext) {
	var kv struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := json.NewDecoder(ctx.req.Body).Decode(&kv)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	cc, err := ccManager.Get("nameservice", "invoke", "addoto", kv.Key, kv.Value)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	bs, err := cc.Invoke()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.res.Write(bs)
	ctx.res.WriteHeader(201)
}

func RemoveNameServiceKV(ctx *RequestContext, params martini.Params) {
	cc, err := ccManager.Get("nameservice", "invoke", "deloto", params["key"])
	if err != nil {
		ctx.Error(500, err)
		return
	}

	bs, err := cc.Invoke()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.res.Write(bs)
	ctx.res.WriteHeader(201)
}
