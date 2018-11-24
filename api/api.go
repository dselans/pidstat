package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dselans/go-pidstat/pid"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	renderPkg "github.com/unrolled/render"
	"go.uber.org/zap"

	"github.com/dselans/go-pidstat/deps"
	"github.com/dselans/go-pidstat/util"
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
		http.FileServer(http.Dir("public")).ServeHTTP(w, r)
	}))

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

// Get all running (and watched) processes
func (a *API) getProcesses(w http.ResponseWriter, r *http.Request) {
	p, err := a.dependencies.Statter.GetProcesses()
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})

		return
	}

	render.JSON(w, http.StatusOK, p)
}

func (a *API) getProcess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "shouldn't be possible to hit this?",
		})

		return
	}

	processID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": fmt.Sprintf("unable to convert id to int: %v", err),
		})

		return
	}

	stats, err := a.dependencies.Statter.GetStatsForPID(int32(processID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("unable to fetch stats for processID '%v': %v", int32(processID), err)

		if err == pid.NotWatchedErr {
			statusCode = http.StatusNotFound
			errorMessage = fmt.Sprintf("processID '%v' is not being watched", int32(processID))
		}

		render.JSON(w, statusCode, map[string]interface{}{"error": errorMessage})
		return
	}

	render.JSON(w, http.StatusOK, stats)
	return
}

func (a *API) startProcessWatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	processID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": fmt.Sprintf("unable to convert id to int: %v", err),
		})

		return
	}

	render.JSON(w, http.StatusOK, map[string]string{"msg": fmt.Sprintf("start process watch for %v", processID)})
}

func (a *API) stopProcessWatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	processID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": fmt.Sprintf("unable to convert id to int: %v", err),
		})

		return
	}

	render.JSON(w, http.StatusOK, map[string]string{"msg": fmt.Sprintf("stop process watch for %v", processID)})
}

func (a *API) getVersion(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, map[string]string{"version": a.version})
}
