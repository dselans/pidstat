package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/dselans/pidstat/stat"
)

var (
	// Needed for swagger docs
	_ = stat.ProcInfo{}
)

// StatusResponse is a generic "status" response that is emitted by the API.
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// VersionResponse is emitted by the API and contains the build version of the application
type VersionResponse struct {
	Version string `json:"version"`
}

// @Summary Get all running processes
// @Description Get a list of all running processes; details include PID, name and cmd line args
// @Tags pid
// @Produce json
// @Success 200 {array} stat.ProcInfo "Contains zero or more process entries"
// @Failure 500 {object} api.StatusResponse "Unexpected server error"
// @Router /api/process [get]
func (a *API) getProcesses(w http.ResponseWriter, r *http.Request) {
	p, err := a.dependencies.Statter.GetProcesses()
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, StatusResponse{
			Status:  "error",
			Message: err.Error(),
		})

		return
	}

	render.JSON(w, http.StatusOK, p)
}

// @Summary Get metrics for a watched process
// @Description Get metrics for a watched process by ID
// @Tags pid
// @Produce json
// @Param pid path string true "Process ID (int)"
// @Param offset query int false "Fetch metrics at offset"
// @Success 200 {object} stat.ProcInfo "Process metrics"
// @Failure 400 {object} api.StatusResponse "Invalid PID (not int) or invalid offset (too high)"
// @Failure 404 {object} api.StatusResponse "PID is not being watched"
// @Failure 416 {object} api.StatusResponse "Invalid offset (too high)"
// @Failure 500 {object} api.StatusResponse "Unexpected server error"
// @Router /api/process/{pid} [get]
func (a *API) getProcess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		render.JSON(w, http.StatusBadRequest, StatusResponse{
			Status:  "error",
			Message: "shouldn't be possible to hit this?",
		})

		return
	}

	processID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, StatusResponse{
			Status:  "error",
			Message: fmt.Sprintf("unable to convert id to int: %v", err),
		})

		return
	}

	// Get QP
	offsetQueryParam := r.URL.Query().Get("offset")

	var offset int

	if offsetQueryParam != "" {
		var err error

		offset, err = strconv.Atoi(offsetQueryParam)
		if err != nil {
			render.JSON(w, http.StatusBadRequest, StatusResponse{
				Status:  "error",
				Message: "offset must be an integer",
			})

			return
		}
	}

	procInfo, err := a.dependencies.Statter.GetStatsForPID(int32(processID), offset)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("unable to fetch stats for processID '%v': %v", int32(processID), err)

		switch err {
		case stat.NotWatchedErr:
			statusCode = http.StatusNotFound
			errorMessage = fmt.Sprintf("processID '%v' is not being watched", int32(processID))
		case stat.InvalidOffsetErr:
			statusCode = http.StatusRequestedRangeNotSatisfiable
			errorMessage = fmt.Sprintf("provided offset is invalid")
		}

		render.JSON(w, statusCode, StatusResponse{
			Status:  "error",
			Message: errorMessage,
		})
		return
	}

	render.JSON(w, http.StatusOK, procInfo)
	return
}

// @Summary Start process watch
// @Description Start process watch for a specific PID
// @Tags pid
// @Produce json
// @Param pid path string true "Process ID (int)"
// @Success 200 {object} api.StatusResponse "Watch has been started for pid"
// @Failure 400 {object} api.StatusResponse "Invalid PID (not int?)"
// @Failure 409 {object} api.StatusResponse "PID is already being watched"
// @Failure 500 {object} api.StatusResponse "Unexpected server error"
// @Router /api/process/{pid} [post]
func (a *API) startProcessWatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	processID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, StatusResponse{
			Status:  "error",
			Message: fmt.Sprintf("unable to convert id to int: %v", err),
		})

		return
	}

	if err := a.dependencies.Statter.StartWatchProcess(int32(processID)); err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("unable to start watch for pid '%v': %v", processID, err)

		if err == stat.AlreadyWatchedErr {
			statusCode = http.StatusConflict
			errorMessage = fmt.Sprintf("pid '%v' is already being watched", processID)
		}

		render.JSON(w, statusCode, StatusResponse{
			Status:  "error",
			Message: errorMessage,
		})
		return
	}

	render.JSON(w, http.StatusOK, StatusResponse{
		Status:  "ok",
		Message: fmt.Sprintf("watch started for pid '%v'", processID),
	})
}

// @Summary Stop process watch
// @Description Stop process watch for a specific PID
// @Tags pid
// @Produce json
// @Param pid path string true "Process ID (int)"
// @Success 200 {object} api.StatusResponse "Watch has been stopped for pid"
// @Failure 400 {object} api.StatusResponse "Invalid PID (not int?)"
// @Failure 404 {object} api.StatusResponse "PID is not being watched"
// @Failure 500 {object} api.StatusResponse "Unexpected server error"
// @Router /api/process/{pid} [delete]
func (a *API) stopProcessWatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	processID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, StatusResponse{
			Status:  "error",
			Message: fmt.Sprintf("unable to convert id to int: %v", err),
		})

		return
	}

	if err := a.dependencies.Statter.StopWatchProcess(int32(processID)); err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("un	able to stop watch for pid '%v': %v", processID, err)

		if err == stat.NotWatchedErr {
			statusCode = http.StatusNotFound
			errorMessage = fmt.Sprintf("pid '%v' is not actively watched", processID)
		}

		render.JSON(w, statusCode, StatusResponse{
			Status:  "error",
			Message: errorMessage,
		})
		return
	}

	render.JSON(w, http.StatusOK, StatusResponse{
		Status:  "ok",
		Message: fmt.Sprintf("watch stopped for pid '%v'", processID),
	})
}

// @Summary Returns the current version of pidstat (api)
// @Description Another simple handler, similar to '/' - if this does not work, something is broken
// @Tags basic
// @Produce json
// @Success 200 {object} api.VersionResponse "Returns the build version. Super simple endpoint -- if this doesn't work, something is busted"
// @Router /api/version [get]
func (a *API) getVersion(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, VersionResponse{a.version})
}

// @Summary View API docs via Swagger-UI
// @Description This endpoint serves the API spec via Swagger-UI (using github.com/swaggo/swag)
// @Tags basic
// @Produce html
// @Success 200 {string} string "Swagger-UI"
// @Router /docs/index.html [get]
func dummyDocs() {}
