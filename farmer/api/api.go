package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-martini/martini"
	daepkg "github.com/hyperledger/fabric/farmer/daemon"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
)

const (
	API_PREFIX = "/api"
)

var (
	log     *logging.Logger
	headers map[string]string
	daemon  *daemob.Daemon
)

func init() {
	headers = make(map[string]string)
	headers["Access-Control-Allow-Origin"] = "*"
	headers["Access-Control-Allow-Methods"] = "GET,OPTIONS,POST,DELETE"
	headers["Access-Control-Allow-Credentials"] = "true"
	headers["Access-Control-Max-Age"] = "864000"
	headers["Access-Control-Expose-Headers"] = "Record-Count,Limt,Offset,Content-Type"
	headers["Access-Control-Allow-Headers"] = "Limt,Offset,Content-Type,Origin,Accept,Authorization"
}

func apiRouter() {
	return func(r martini.Router) {
		r.Post("/user", RegistryUser)
		r.Get("/user", GetUser)
	}
}

func Serve(d *daepkg.Daemon) {
	daemon = d
	log := d.GetLogger()

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

	m.Group(API_PREFIX, apiRouter(d))

	server := &http.Server{
		Handler:  m,
		Addr:     listenAddr,
		ErrorLog: stdlog.New(logger.Writer(), "", 0),
	}

	log.Info("server is starting on ", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Error(err)
	}
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
