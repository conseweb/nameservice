package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/core"
	"github.com/hyperledger/fabric/peer/common"
	"github.com/hyperledger/fabric/peer/util"
	pb "github.com/hyperledger/fabric/protos"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

type ChaincodeWrapper struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Method   string   `json:"method"`
	Language string   `json:"language"`
	Function string   `json:"function"`
	Args     []string `json:"args"`

	UserName      string   `json:"user_name"`
	Attributes    []string `json:"attributes"`
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
	Attributes    []string               `json:"attributes"`
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
			Attributes: c.Attributes,
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

func (c *ChaincodeWrapper) ToSpec() (*pb.ChaincodeSpec, error) {
	input := make([][]byte, len(c.Args))
	for i, s := range c.Args {
		input[i] = []byte(s)
	}

	if c.Language == "" {
		c.Language = "GOLANG"
	} else {
		c.Language = strings.ToUpper(c.Language)
	}

	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[c.Language]),
		ChaincodeID: &pb.ChaincodeID{
			Path: c.Path,
			Name: c.Name,
		},
		CtorMsg: &pb.ChaincodeInput{
			Args: input,
		},
		Attributes: c.Attributes,
	}

	if core.SecurityEnabled() {
		if c.UserName == common.UndefinedParamValue {
			return spec, errors.New("Must supply username for chaincode when security is enabled")
		}

		// Retrieve the CLI data storage path
		localStore := util.GetCliFilePath()

		// Check if the user is logged in before sending transaction
		if _, err := os.Stat(localStore + "loginToken_" + c.UserName); err == nil {
			log.Infof("Local user '%s' is already logged in. Retrieving login token.\n", c.UserName)

			// Read in the login token
			token, err := ioutil.ReadFile(localStore + "loginToken_" + c.UserName)
			if err != nil {
				return nil, fmt.Errorf("Fatal error when reading client login token: %s\n", err)
			}

			// Add the login token to the chaincodeSpec
			spec.SecureContext = string(token)

			// If privacy is enabled, mark chaincode as confidential
			if viper.GetBool("security.privacy") {
				log.Info("Set confidentiality level to CONFIDENTIAL.\n")
				spec.ConfidentialityLevel = pb.ConfidentialityLevel_CONFIDENTIAL
			}
		} else {
			// Check if the token is not there and fail
			if os.IsNotExist(err) {
				return spec, fmt.Errorf("User '%s' not logged in. Use the 'peer network login' command to obtain a security token.", c.UserName)
			}
			// Unexpected error
			return nil, fmt.Errorf("Fatal error when checking for client login token: %s\n", err)
		}
	} else {
		if c.UserName != common.UndefinedParamValue {
			log.Warning("Username supplied but security is disabled.")
		}
		if viper.GetBool("security.privacy") {
			return nil, errors.New("Privacy cannot be enabled as requested because security is disabled")
		}
	}

	return spec, nil
}

func (c *ChaincodeWrapper) Deploy() (string, error) {
	spec, err := c.ToSpec()
	if err != nil {
		return "", err
	}

	devopsClient, err := common.GetDevopsClient(nil)
	if err != nil {
		return "", fmt.Errorf("Error building %s: %s", c.Function, err)
	}

	log.Infof("deploy: %+v", spec)
	chaincodeDeploymentSpec, err := devopsClient.Deploy(context.Background(), spec)
	if err != nil {
		return "", fmt.Errorf("Error building %s: %s\n", c.Function, err)
	}
	log.Infof("Deploy result: %s", chaincodeDeploymentSpec.ChaincodeSpec)

	return chaincodeDeploymentSpec.ChaincodeSpec.ChaincodeID.Name, nil
}

func (c *ChaincodeWrapper) Invoke() ([]byte, error) {
	spec, err := c.ToSpec()
	if err != nil {
		return nil, err
	}

	devopsClient, err := common.GetDevopsClient(nil)
	if err != nil {
		return nil, fmt.Errorf("Error building %s: %s", c.Function, err)
	}

	log.Infof("deploy: %+v", spec)
	resp, err := devopsClient.Invoke(context.Background(), &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec})
	if err != nil {
		return nil, fmt.Errorf("Error building %s: %s\n", c.Function, err)
	}
	log.Infof("Deploy result: %s", resp.Msg)

	return resp.Msg, nil
}

func (c *ChaincodeWrapper) Query() ([]byte, error) {
	return c.Invoke()
}

func DeployCC(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ChaincodeWrapper{}
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

	ctx.rnd.JSON(201, map[string]interface{}{"message": ret})
}

func InvokeCC(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ChaincodeWrapper{}
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

	ctx.rnd.JSON(200, map[string]interface{}{"message": ret})
}

func QueryCC(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cw := &ChaincodeWrapper{}
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

	ctx.rnd.JSON(200, map[string]interface{}{"message": ret})
}

func SetChaincode(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	alias := params["alias"]
	if alias == "" {
		ctx.Error(400, fmt.Errorf("alias is invalied."))
		return
	}

	cc := &ChaincodeWrapper{}
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
