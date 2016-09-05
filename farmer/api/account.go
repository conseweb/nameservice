package api

import (
	"encoding/json"
	"net/http"

	pb "github.com/conseweb/common/protos"
	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/farmer/account"
	"github.com/martini-contrib/render"
	"golang.org/x/net/context"
)

type accountWrapper struct {
	// sign type
	Type     string `json:"type"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Captcha  string `json:"captcha"`
	Language string `json:"language"`
	NickName string `json:"nickname"`
	Password string `json:"password"`
}

func (a *accountWrapper) SignUpArgus() (pb.SignUpType, string) {
	stype := pb.SignUpType_MOBILE
	svalue := a.Phone

	switch a.Type {
	case "phone":
	case "email":
		stype = pb.SignUpType_EMAIL
		svalue = a.Email
	default:
		if a.Email != "" {
			stype = pb.SignUpType_EMAIL
			svalue = a.Email
		}
	}
	return stype, svalue
}

func (a *accountWrapper) ToAccount() (*account.Account, error) {
	/// TODO: add check
	return account.NewAccount(a.NickName, a.Phone, a.Email, a.Password, a.Language), nil
}

func RegVerificationType(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	var user accountWrapper

	err := json.NewDecoder(ctx.req.Body).Decode(&user)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	stype, svalue := user.SignUpArgus()

	cli, err := daemon.GetIDPClient()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	resp, err := cli.AcquireCaptcha(context.Background(), &pb.AcquireCaptchaReq{stype, svalue})
	if err != nil {
		ctx.Error(501, err, "idp server error")
		return
	}
	if resp.Error != nil && !resp.Error.OK() {
		ctx.Error(500, resp.Error)
		return
	}
	rw.WriteHeader(200)
}

func VerifyCaptcha(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	var user accountWrapper

	err := json.NewDecoder(ctx.req.Body).Decode(&user)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	stype, svalue := user.SignUpArgus()

	cli, err := daemon.GetIDPClient()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	resp, err := cli.VerifyCaptcha(context.Background(), &pb.VerifyCaptchaReq{stype, svalue, user.Captcha})
	if err != nil {
		ctx.Error(501, err, "idp server error")
		return
	}
	if resp.Error != nil && !resp.Error.OK() {
		ctx.Error(500, resp.Error)
		return
	}
	rw.WriteHeader(200)
}

func Registry(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	var user accountWrapper
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	// registry
	cli, err := daemon.GetIDPClient()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	acc, err := user.ToAccount()
	if err != nil {
		ctx.Error(400, err)
		return
	}

	if err := acc.Registry(cli); err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.rnd.JSON(201, acc)
}

func Login(ctx *RequestContext) {

	ctx.res.WriteHeader(200)
}

func OnlineAccount(ctx *RequestContext) {

	ctx.res.WriteHeader(200)
}

func OfflineAccount(ctx *RequestContext) {

	ctx.res.WriteHeader(200)
}

func GetAccountState(rw http.ResponseWriter, req *http.Request, rnd render.Render) {
	rnd.JSON(200, map[string]string{"msg": "hello"})
}
