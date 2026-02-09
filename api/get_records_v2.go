package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/service"
)

// GetRecord returns the latest version, or a historical version if ?at= is specified
func (a *APIV2) GetRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNum, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNum <= 0 {
		writeError(w, "invalid id; must be a positive integer", http.StatusBadRequest)
		return
	}

	// check if they want a specific point in time
	if atStr := r.URL.Query().Get("at"); atStr != "" {
		ts, err := time.Parse(time.RFC3339, atStr)
		if err != nil {
			writeError(w, "invalid 'at' param; use RFC3339 format like 2026-01-15T10:30:00Z", http.StatusBadRequest)
			return
		}

		record, err := a.versioned.GetRecordAtTime(ctx, int(idNum), ts)
		if err != nil {
			handleV2Error(w, err, int(idNum))
			return
		}
		writeJSON(w, record, http.StatusOK)
		return
	}

	record, err := a.versioned.GetLatestRecord(ctx, int(idNum))
	if err != nil {
		handleV2Error(w, err, int(idNum))
		return
	}
	writeJSON(w, record, http.StatusOK)
}

func (a *APIV2) GetRecordVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	idNum, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil || idNum <= 0 {
		writeError(w, "invalid id", http.StatusBadRequest)
		return
	}

	verNum, err := strconv.ParseInt(vars["version"], 10, 32)
	if err != nil || verNum <= 0 {
		writeError(w, "invalid version number", http.StatusBadRequest)
		return
	}

	record, err := a.versioned.GetRecordAtVersion(ctx, int(idNum), int(verNum))
	if err != nil {
		if errors.Is(err, service.ErrVersionNotFound) {
			writeError(w, fmt.Sprintf("version %d not found for record %d", verNum, idNum), http.StatusNotFound)
			return
		}
		handleV2Error(w, err, int(idNum))
		return
	}
	writeJSON(w, record, http.StatusOK)
}

func handleV2Error(w http.ResponseWriter, err error, id int) {
	switch {
	case errors.Is(err, service.ErrRecordDoesNotExist):
		writeError(w, fmt.Sprintf("record %d does not exist", id), http.StatusNotFound)
	case errors.Is(err, service.ErrRecordIDInvalid):
		writeError(w, "invalid id", http.StatusBadRequest)
	default:
		logError(err)
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}
