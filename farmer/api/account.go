package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	pb "github.com/conseweb/common/protos"
	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/farmer/account"
	"github.com/martini-contrib/render"
	"golang.org/x/net/context"
)

type accountWrapper struct {
	// sign type email/phone
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

func (a *accountWrapper) SignInArgus() (pb.SignInType, string) {
	stype := pb.SignInType_SI_MOBILE
	svalue := a.Phone

	switch a.Type {
	case "phone":
	case "email":
		stype = pb.SignInType_SI_EMAIL
		svalue = a.Email
	default:
		if a.Email != "" {
			stype = pb.SignInType_SI_EMAIL
			svalue = a.Email
		}
	}
	return stype, svalue
}

func (a *accountWrapper) ToAccount() (*account.Account, error) {
	/// TODO: add check
	return account.NewAccount(a.NickName, a.Phone, a.Email, a.Password, a.Language), nil
}

// POST /signup/:vtype
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

// POST /signup
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

	stype, svalue := user.SignUpArgus()
	resp, err := cli.VerifyCaptcha(context.Background(), &pb.VerifyCaptchaReq{stype, svalue, user.Captcha})
	if err != nil {
		ctx.Error(501, err, "idp server error")
		return
	}
	if resp.Error != nil && !resp.Error.OK() {
		ctx.Error(400, resp.Error)
		return
	}

	// Try to new a Account.
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
	var user accountWrapper
	json.NewDecoder(ctx.req.Body).Decode(&user)

	cli, err := daemon.GetIDPClient()
	if err != nil {
		ctx.Error(500, err)
		return
	}

	st, su := user.SignInArgus()
	a, err := account.Login(cli, st, su, user.Password)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	daemon.ResetAccount(a)
	log.Debugf("set nuew account %+v", a)

	ctx.rnd.JSON(200, a)
}

func Logout(ctx *RequestContext) {
	daemon.ResetAccount(nil)
	ctx.res.WriteHeader(200)
}

func UnbindDevide(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	// registry
	// cli, err := daemon.GetIDPClient()
	// if err != nil {
	// 	ctx.Error(500, err)
	// 	return
	// }
}

func BindDevide(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {

}

func OnlineAccount(ctx *RequestContext) {

	ctx.res.WriteHeader(200)
}

func OfflineAccount(ctx *RequestContext) {

	ctx.res.WriteHeader(200)
}

func GetAccountState(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, rnd render.Render) {
	if daemon.Account == nil {
		ctx.Error(404, "not found account")
		return
	}
	rnd.JSON(200, daemon.Account)
}

// contacts.
func ListContacts(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	ret, err := (*account.Contact).List(nil, ctx.db)
	if err != nil {
		ctx.Error(500, err)
		return
	}
	ctx.rnd.JSON(200, ret)
}

func AddContacts(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	cont := &account.Contact{}
	err := json.NewDecoder(ctx.req.Body).Decode(cont)
	if err != nil {
		ctx.Error(400, fmt.Errorf("invalied data, %s", err))
		return
	}

	if err = cont.Insert(ctx.db); err != nil {
		ctx.Error(500, err)
		return
	}
	ctx.rnd.JSON(201, cont)
}

func UpdateContacts(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		ctx.Error(400, fmt.Errorf("invalied id<%v>, %s", params["id"], err))
		return
	}

	newc := &account.Contact{}
	err = json.NewDecoder(ctx.req.Body).Decode(newc)
	if err != nil {
		ctx.Error(400, fmt.Errorf("invalied data, %s", err))
		return
	}

	upd := &account.Contact{Id: id}
	err = upd.Update(ctx.db, newc)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.rnd.JSON(200, upd)
}

func RemoveAllContacts(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	if err := (*account.Contact).RemoveAll(nil, ctx.db); err != nil {
		ctx.Error(400, err)
		return
	}
	ctx.Message(200, "successful")
}

func RemoveContacts(rw http.ResponseWriter, req *http.Request, ctx *RequestContext, params martini.Params) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		ctx.Error(400, fmt.Errorf("invalied id<%v>, %s", params["id"], err))
		return
	}

	err = (*account.Contact).Remove(nil, ctx.db, id)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.Message(200, "successful")
}
