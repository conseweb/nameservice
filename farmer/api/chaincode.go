package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	ccpkg "github.com/hyperledger/fabric/peer/chaincode"
)

var (
	ErrNotFound  = errors.New("not found chaincode.")
	ErrNotDeploy = errors.New("not deploy chaincode.")
)

type chaincodeManager struct {
	sync.Mutex
	ccs map[string]*ccpkg.ChaincodeWrapper
}

// args[...] 0=method, 1=funtion >2=args...
func (c *chaincodeManager) Get(key string, args ...string) (*ccpkg.ChaincodeWrapper, error) {
	c.Lock()
	defer c.Unlock()

	cc, ok := c.ccs[key]
	if !ok {
		return nil, fmt.Errorf("chaincode %s, %v", key, ErrNotFound)
	}

	if cc.Name == "" {
		return cc, ErrNotDeploy
	}

	var newcc ccpkg.ChaincodeWrapper = *cc

	newcc.Args = []string{}
	newcc.Attributes = []string{}

	l := len(args)
	switch {
	case l >= 1:
		newcc.Method = args[0]
		fallthrough
	case l >= 2:
		newcc.Functon = args[1]
		fallthrough
	case l >= 3:
		newcc.Args = args[2:]
	default:
	}
	return &newcc, nil
}

func (c *chaincodeManager) SetName(key, name string) {
	c.Lock()
	defer c.Unlock()

	cc, ok := c.ccs[key]
	if !ok {
		c.ccs[key] = &ccpkg.ChaincodeWrapper{Name: name}
	}

	cc.Name = name
}

func (c *chaincodeManager) Set(key string, cc *ccpkg.ChaincodeWrapper) {
	c.Lock()
	defer c.Unlock()

	c.ccs[key] = cc
}

/// POST /cc/deploy?alisa=xxx
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

	name := ctx.params["alisa"]
	if name == "" {
		name = ret
	}

	ccManager.Set(name, cw)

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

func ListChaincodes(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	ctx.rnd.JSON(200, chaincodeManager.ccs)
}

func GetChaincode(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	alias := params["alias"]

	cc, err := ccManager.Get(alias)
	if err == ErrNotFound {
		ctx.Error(404, fmt.Errorf("not found"))
		return
	}

	ctx.rnd.JSON(200, cc)
}
