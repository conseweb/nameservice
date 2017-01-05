package api

import (
	"fmt"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/farmer/indexer"
	"github.com/hyperledger/fabric/storage/localfs"
	"github.com/spf13/viper"
)

// need user login
func AuthMW(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	if !daemon.IsLogin() {
		ctx.Message(401, "login required.")
		return
	}
}

// if lepuscoin chaincode not deploy, try to deploy and return a 501 error.
func DeployLepuscoinMW(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
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
		ctx.Error(501, fmt.Errorf("lepuscoin chaincode is deploying, please wait."))
		return
	} else if err != nil {
		ctx.Error(400, err)
		return
	}
}

func DeployNameSrvnMW(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
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
		ctx.Error(501, fmt.Errorf("nameservice chaincode is deploying, please wait."))
		return
	} else if err != nil {
		ctx.Error(400, err)
		return
	}
}

func SetIndexerDBMW(ctx *RequestContext, mc martini.Context) {
	orm, err := indexer.InitDB()
	if err != nil {
		ctx.Error(500, err)
		return
	}
	mc.Map(orm)
}

func SetFsDriverMW(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, mc martini.Context) {
	if fsDriver != nil {
		mc.Map(fsDriver)
		return
	}

	fstype := viper.GetString("farmer.fstype")
	rootPath := viper.GetString("farmer.localChroot")
	var err error

	switch fstype {
	case "ipfs":
		ctx.Error(500, "TODO")
		return
	case "local":
		log.Infof("farmer use local filesystem")
		fsDriver, err = localfs.NewDriver(rootPath)
		if err != nil {
			ctx.Error(500, "get chroot path failed.")
			return
		}
	default:
		ctx.Error(500, fmt.Errorf("unknown storage type %s", fstype))
		return
	}
	mc.Map(fsDriver)
}
