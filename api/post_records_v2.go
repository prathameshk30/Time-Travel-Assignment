package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *APIV2) PostRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNum, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNum <= 0 {
		writeError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body map[string]*string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "could not parse json body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		writeError(w, "empty request body", http.StatusBadRequest)
		return
	}

	record, err := a.versioned.CreateOrUpdateRecord(ctx, int(idNum), body)
	if err != nil {
		handleV2Error(w, err, int(idNum))
		return
	}

	// 201 for brand new, 200 for updates
	status := http.StatusOK
	if record.Version == 1 {
		status = http.StatusCreated
	}
	writeJSON(w, record, status)
}
