package api

import (
	"database/sql"
	stdlog "log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/farmer/api/views"
	daepkg "github.com/hyperledger/fabric/farmer/daemon"
	ccpkg "github.com/hyperledger/fabric/peer/chaincode"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	API_PREFIX      = "/api"
	SOCKETIO_PREFIX = "/socket.io"
)

var (
	log         *logging.Logger
	daemon      *daepkg.Daemon
	proxyClient *http.Client

	ccManager = &chaincodeManager{}
)

func init() {
	ccManager.ccs = map[string]*ccpkg.ChaincodeWrapper{
		"lepuscoin": &ccpkg.ChaincodeWrapper{
			Path: "github.com/conseweb/common/assets/lepuscoin",
			Name: "",
		},
		"poe": &ccpkg.ChaincodeWrapper{
			Path: "github.com/conseweb/common/assets/poe",
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

	db  *sql.DB
	evt *EventHandler
}

type EventClient struct {
	eventChan chan string
}

type eventHandle struct {
	es map[string]EventClient
}

func notFound(gateways map[string]*url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for way, _ := range gateways {
			if strings.HasPrefix(req.URL.Path, way) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		for _, v := range []string{API_PREFIX, SOCKETIO_PREFIX} {
			if strings.HasPrefix(req.URL.Path, v) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
	}
}

func getGatewayRouters() (map[string]*url.URL, error) {
	gateways := map[string]*url.URL{}

	for k, v := range viper.GetStringMapString("daemon.gateway") {
		log.Noticef("get setting proxy router, %s --> %s", k, v)
		way := "/" + strings.Trim(k, "/ \n")
		to, err := url.Parse(v)
		if err != nil {
			log.Errorf("formt URL<%s> failed, error: %s", v, err.Error())
			return nil, err
		}
		gateways[way] = to
	}

	return gateways, nil
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

	// gateway.
	gateways, err := getGatewayRouters()
	if err != nil {
		return err
	}

	for way, to := range gateways {
		m.Any(way+"/**", ProxyTo(way, to))
		log.Noticef("add proxy router %s --> %s", way, to.String())
	}

	view := views.New()
	m.NotFound(notFound(gateways), view.ServeHTTP)

	evt := NewEventHandler()
	m.Map(evt)

	m.Use(requextCtx)

	m.Any(SOCKETIO_PREFIX, evt.ServeHTTP)
	m.Group(API_PREFIX, func(r martini.Router) {
		/// no auth
		r.Post("/signup/:vtype", RegVerificationType)
		r.Post("/signup", Registry)
		r.Post("/account/login", Login)

		/// need user auth
		r.Group("", func(r martini.Router) {
			r.Get("/account", GetAccountState)
			r.Delete("/account/logout", Logout)
			r.Patch("/account/setting", Hello)

			// local contacts
			r.Get("/account/contacts", ListContacts)
			r.Post("/account/contacts", AddContacts)
			r.Patch("/account/contacts/:id", UpdateContacts)
			r.Delete("/account/contacts", RemoveAllContacts)
			r.Delete("/account/contacts/:id", RemoveContacts)

			r.Post("/device/bind", Hello)
			r.Delete("/device/unbind", Hello)
			r.Post("/device/tx", NewTx)
			r.Get("/device/coinbase_tx/:addr", GetCoinBaseTx)

			/// need deploy lepuscoin chaincode.
			r.Group("/lepuscoin", func(r martini.Router) {
				r.Post("/tx", NewTx)
				r.Post("/deploy", DeployLepuscoin)
				r.Post("/coinbase", DoCoinbase)
				r.Post("/transfer", Transfer)
				r.Get("/balance", QueryAddrs)
				r.Get("/coin", QueryCoin)
				r.Get("/tx/:tx", QueryTx)
			}, DeployLepuscoinMW)
		}, AuthMW)

		r.Post("/cc/deploy", DeployCC)
		r.Post("/cc/invoke", InvokeCC)
		r.Post("/cc/query", QueryCC)

		r.Patch("/peer/start", StartPeer)
		r.Patch("/peer/stop", StopPeer)
		r.Patch("/peer/restart", RestartPeer)

		r.Get("/metrics", GetMetrics)

		r.Get("/chaincode", ListChaincodes)
		r.Get("/chaincode/:alias", GetChaincode)

		/// proxy fo fabirc
		r.Post("/chaincode", ProxyChaincode, ProxyFabric)

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
		db:     daemon.GetDB(),
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
