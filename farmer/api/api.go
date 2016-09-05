package api

import (
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/farmer/api/views"
	daepkg "github.com/hyperledger/fabric/farmer/daemon"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/op/go-logging"
	// "github.com/googollee/go-socket.io"
)

const (
	API_PREFIX = "/api"
)

var (
	log     *logging.Logger
	headers map[string]string
	daemon  *daepkg.Daemon
)

type RequestContext struct {
	params martini.Params
	mc     martini.Context

	req *http.Request
	rnd render.Render
	res http.ResponseWriter

	eventChan chan string
}

type EventClient struct {
	eventChan chan string
}

type eventHandle struct {
	es map[string]EventClient
}

func init() {
	headers = make(map[string]string)
	headers["Access-Control-Allow-Origin"] = "*"
	headers["Access-Control-Allow-Methods"] = "GET,OPTIONS,POST,DELETE,PUT"
	headers["Access-Control-Allow-Credentials"] = "true"
	headers["Access-Control-Max-Age"] = "864000"
	headers["Access-Control-Expose-Headers"] = "Record-Count,Limt,Offset,Content-Type"
	headers["Access-Control-Allow-Headers"] = "Limt,Offset,Content-Type,Origin,Accept,Authorization"
}

func notFound(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, API_PREFIX) {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Serve(d *daepkg.Daemon) error {
	daemon = d
	log := d.GetLogger()
	listenAddr := d.ListenAddr

	m := NewMartini()
	// view := views.New()
	// m.NotFound(NotFound, view.ServeHTTP)

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     strings.Split(headers["Access-Control-Allow-Origin"], ","),
		AllowMethods:     strings.Split(headers["Access-Control-Allow-Methods"], ","),
		AllowHeaders:     strings.Split(headers["Access-Control-Allow-Headers"], ","),
		ExposeHeaders:    strings.Split(headers["Access-Control-Expose-Headers"], ","),
		AllowCredentials: true,
		MaxAge:           time.Second * 864000,
	}))

	view := views.New()
	m.NotFound(notFound, view.ServeHTTP)
	m.Use(requextCtx)

	m.Group(API_PREFIX, func(r martini.Router) {
		r.Group("/signup", func(r martini.Router) {
			r.Post("", Registry)
			r.Post("/:vtype", RegVerificationType)
			r.Put("/captcha", VerifyCaptcha)
		})
		r.Group("/account", func(r martini.Router) {
			r.Get("/state", GetAccountState)
			r.Post("/login", Registry)
			r.Delete("/logout", Hello)
			r.Patch("/setting", Hello)
		})
		r.Group("/device", func(r martini.Router) {
			r.Get("")
			r.Post("/bind", Hello)
			r.Delete("/unbind", Hello)
		})

		r.Any("/node", Hello)

		// setting
		// node
		// chaincode
		// network
	})

	server := &http.Server{
		Handler:  m,
		Addr:     listenAddr,
		ErrorLog: stdlog.New(os.Stderr, "", 0),
	}

	log.Info("server is starting on ", listenAddr)
	return server.ListenAndServe()
}

func NewMartini() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{Martini: m, Router: r}
}

func requextCtx(w http.ResponseWriter, req *http.Request, mc martini.Context, rnd render.Render) {
	ctx := &RequestContext{
		res:    w,
		req:    req,
		mc:     mc,
		rnd:    rnd,
		params: make(map[string]string),
	}

	req.ParseForm()
	if len(req.Form) > 0 {
		for k, v := range req.Form {
			ctx.params[k] = v[0]
		}
	}

	mc.Map(ctx)
	mc.Next()
}

func Hello(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	ctx.rnd.JSON(200, map[string]string{"msg": "hello"})
}
