package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	renderPkg "github.com/unrolled/render"
	"go.uber.org/zap"

	"github.com/dselans/pidstat/deps"
	"github.com/dselans/pidstat/util"
)

var (
	sugar  *zap.SugaredLogger
	render *renderPkg.Render
)

func init() {
	logger, err := util.CreateLogger(false, map[string]interface{}{"pkg": "api"})
	if err != nil {
		panic(fmt.Sprintf("unable to setup logger: %v", err))
	}

	sugar = logger.Sugar()

	// instantiate render
	render = renderPkg.New()
}

type API struct {
	listenAddress string
	dependencies  *deps.Dependencies
	version       string
}

func New(listenAddress, version string, d *deps.Dependencies) (*API, error) {
	return &API{
		listenAddress: listenAddress,
		dependencies:  d,
		version:       version,
	}, nil
}

func (a *API) Run() error {
	r := chi.NewRouter()

	// Don't need a whole lot
	r.Use(middleware.Logger)

	// Serve static files
	r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(a.dependencies.PackrBox).ServeHTTP(w, r)
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/version", a.getVersion)
		r.Get("/process", a.getProcesses)
		r.Get("/process/{id}", a.getProcess)
		r.Post("/process/{id}", a.startProcessWatch)
		r.Delete("/process/{id}", a.stopProcessWatch)
	})

	sugar.Infof("server listening on '%v'", a.listenAddress)

	return http.ListenAndServe(a.listenAddress, r)
}
