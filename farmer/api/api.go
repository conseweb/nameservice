package api

import (
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/farmer/api/views"
	daepkg "github.com/hyperledger/fabric/farmer/daemon"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/op/go-logging"
)

const (
	API_PREFIX      = "/api"
	SOCKETIO_PREFIX = "/socket.io"
)

var (
	log         *logging.Logger
	daemon      *daepkg.Daemon
	proxyClient *http.Client

	chaincodeManager struct {
		sync.Mutex
		ccs map[string]*ChaincodeWrapper
	}
)

func init() {
	chaincodeManager.ccs = map[string]*ChaincodeWrapper{
		"lepuscoin": &ChaincodeWrapper{
			Path: "github.com/conseweb/common/assets/lepuscoin",
			Name: "",
		},
	}
}

type RequestContext struct {
	params martini.Params
	mc     martini.Context

	req *http.Request
	rnd render.Render
	res http.ResponseWriter

	evt *EventHandler
}

type EventClient struct {
	eventChan chan string
}

type eventHandle struct {
	es map[string]EventClient
}

func notFound(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, API_PREFIX) ||
		strings.HasPrefix(req.URL.Path, SOCKETIO_PREFIX) {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Serve(d *daepkg.Daemon) error {
	daemon = d
	log = d.GetLogger()
	listenAddr := d.ListenAddr

	m := NewMartini()

	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PATCH", "POST", "DELETE", "PUT"},
		AllowHeaders:     []string{"Limt", "Offset", "Content-Type", "Origin", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Record-Count", "Limt", "Offset", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           time.Second * 864000,
	}))

	view := views.New()
	m.NotFound(notFound, view.ServeHTTP)

	evt := NewEventHandler()
	m.Map(evt)

	m.Use(requextCtx)

	m.Any(SOCKETIO_PREFIX, evt.ServeHTTP)
	m.Group(API_PREFIX, func(r martini.Router) {
		r.Group("/signup", func(r martini.Router) {
			r.Post("", Registry)
			r.Post("/:vtype", RegVerificationType)
			r.Put("/captcha", VerifyCaptcha)
		})

		r.Group("/account", func(r martini.Router) {
			r.Get("", GetAccountState)
			r.Post("/login", Login)
			r.Delete("/logout", Logout)
			r.Patch("/setting", Hello)
		})
		r.Group("/device", func(r martini.Router) {
			r.Post("/bind", Hello)
			r.Delete("/unbind", Hello)
			r.Post("/tx", NewTx)
			r.Get("/coinbase_tx/:addr", GetCoinBaseTx)
		})

		r.Group("/peer", func(r martini.Router) {
			r.Patch("/start", StartPeer)
			r.Patch("/stop", StopPeer)
			r.Patch("/restart", RestartPeer)
		})
		r.Group("/metrics", func(r martini.Router) {
			r.Get("", GetMetrics)
		})
		r.Group("/chaincode", func(r martini.Router) {
			r.Get("", ListChaincodes)
			r.Get("/:alias", GetChaincode)
			r.Post("/:alias", SetChaincode)

			r.Post("", ProxyChaincode, ProxyFabric)
		})

		r.Any("/chain", ProxyFabric)
		r.Any("/chain/**", ProxyFabric)
		r.Any("/devops/**", ProxyFabric)
		r.Any("/registrar/**", ProxyFabric)
		r.Any("/transactions/**", ProxyFabric)
		r.Any("/network/**", ProxyFabric)
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

func requextCtx(w http.ResponseWriter, req *http.Request, mc martini.Context, rnd render.Render, evt *EventHandler) {
	ctx := &RequestContext{
		res:    w,
		req:    req,
		mc:     mc,
		rnd:    rnd,
		evt:    evt,
		params: make(map[string]string),
	}

	req.ParseForm()
	if len(req.Form) > 0 {
		for k, v := range req.Form {
			ctx.params[k] = v[0]
		}
	}

	log.Debugf("[%s] %s", req.Method, req.URL.String())

	mc.Map(ctx)
	mc.Next()
}

func Hello(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	ctx.rnd.JSON(200, map[string]string{"message": "hello"})
}
