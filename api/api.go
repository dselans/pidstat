package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/dselans/go-pidstat/deps"
	"github.com/dselans/go-pidstat/util"
)

var (
	sugar *zap.SugaredLogger
)

func init() {
	logger, err := util.CreateLogger(false, map[string]interface{}{"pkg": "api"})
	if err != nil {
		panic(fmt.Sprintf("unable to setup logger: %v", err))
	}

	sugar = logger.Sugar()
}

type API struct {
	listenAddress string
	dependencies  *deps.Dependencies
}

func New(listenAddress string, d *deps.Dependencies) (*API, error) {
	return &API{
		listenAddress: listenAddress,
		dependencies:  d,
	}, nil
}

func (a *API) Run() error {
	r := chi.NewRouter()

	r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir("public")).ServeHTTP(w, r)
	}))

	r.Get("/version", a.getVersion)

	sugar.Infof("server listening on '%v'", a.listenAddress)

	return http.ListenAndServe(a.listenAddress, r)
}

func (a *API) getVersion(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello there"))
}