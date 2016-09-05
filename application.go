package clevergo

import (
	"github.com/clevergo/sessions"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

type Application struct {
	defaultRouter *Router
	routers       map[string]*Router
	sessionStore  sessions.Store
	logger        fasthttp.Logger
	Config        *Config
}

func NewApplication() *Application {
	return &Application{
		defaultRouter: NewRouter(),
		routers:       make(map[string]*Router, 0),
		Config:        NewConfig(),
	}
}

func (a *Application) SetDefaultRouter(r *Router) {
	a.defaultRouter = r
}

func (a *Application) SetLogger(logger fasthttp.Logger) {
	a.logger = logger
}

func (a *Application) SetSessionStore(store sessions.Store) {
	a.sessionStore = store
}

func (a *Application) AddRouter(domain string, r *Router) {
	// Set default session store.
	if r.sessionStore == nil {
		r.sessionStore = a.sessionStore
	}

	// Set default logger.
	if r.logger == nil {
		r.logger = a.logger
	}
	a.routers[domain] = r

	// Set the current router as default, if the domain is an empty string.
	if len(domain) == 0 {
		a.defaultRouter = r
	}
}

// Get application handler.
//
// If there is only one router and it also is the default router,
// return it's handler.
// Otherwise, returns the application.Handler.
func (a *Application) getHandler() func(ctx *fasthttp.RequestCtx) {
	if router, ok := a.routers[""]; len(a.routers) == 1 && ok {
		return router.Handler
	}

	return a.Handler
}

func (a *Application) Handler(ctx *fasthttp.RequestCtx) {
	host := strings.Split(string(ctx.Host()), ":")
	if r, ok := a.routers[host[0]]; ok {
		r.Handler(ctx)
		return
	}

	a.defaultRouter.Handler(ctx)
}

func (a *Application) Run() {
	if len(a.routers) == 0 {
		panic("No router.")
	}

	handler := a.getHandler()

	switch a.Config.ServerType {
	case ServerTypeUNIX:
		log.Fatal(ListenAndServeUNIX(
			a.Config.ServerAddr,
			a.Config.ServerMode,
			handler,
		))
	case ServerTypeTLS:
		log.Fatal(ListenAndServeTLS(
			a.Config.ServerAddr,
			a.Config.ServerCertFile,
			a.Config.ServerKeyFile,
			handler,
		))
	case ServerTypeTLSEmbed:
		log.Fatal(ListenAndServeTLSEmbed(
			a.Config.ServerAddr,
			a.Config.ServerCertData,
			a.Config.ServerKeyData,
			handler,
		))
	default:
		log.Fatal(ListenAndServe(a.Config.ServerAddr, handler))
	}

}
