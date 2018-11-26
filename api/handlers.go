package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/dselans/go-pidstat/stat"
)

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

	procInfo, err := a.dependencies.Statter.GetStatsForPID(int32(processID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("unable to fetch stats for processID '%v': %v", int32(processID), err)

		if err == stat.NotWatchedErr {
			statusCode = http.StatusNotFound
			errorMessage = fmt.Sprintf("processID '%v' is not being watched", int32(processID))
		}

		render.JSON(w, statusCode, map[string]interface{}{"error": errorMessage})
		return
	}

	render.JSON(w, http.StatusOK, procInfo)
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

	if err := a.dependencies.Statter.StartWatchProcess(int32(processID)); err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("unable to start watch for pid '%v': %v", processID, err)

		if err == stat.AlreadyWatchedErr {
			statusCode = http.StatusBadRequest
			errorMessage = fmt.Sprintf("pid '%v' is already being watched", processID)
		}

		render.JSON(w, statusCode, map[string]interface{}{"error": errorMessage})
		return
	}

	render.JSON(w, http.StatusOK, map[string]string{"msg": fmt.Sprintf("watch started for pid '%v'", processID)})
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

	if err := a.dependencies.Statter.StopWatchProcess(int32(processID)); err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := fmt.Sprintf("un	able to stop watch for pid '%v': %v", processID, err)

		if err == stat.NotWatchedErr {
			statusCode = http.StatusBadRequest
			errorMessage = fmt.Sprintf("pid '%v' is not actively watched", processID)
		}

		render.JSON(w, statusCode, map[string]interface{}{"error": errorMessage})
		return
	}

	render.JSON(w, http.StatusOK, map[string]string{"msg": fmt.Sprintf("watch stopped for pid '%v'", processID)})
}

func (a *API) getVersion(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, map[string]string{"version": a.version})
}
