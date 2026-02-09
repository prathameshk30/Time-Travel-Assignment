package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type VersionListResponse struct {
	ID       int           `json:"id"`
	Versions []VersionMeta `json:"versions"`
}

type VersionMeta struct {
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
}

func (a *APIV2) ListVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNum, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNum <= 0 {
		writeError(w, "invalid id", http.StatusBadRequest)
		return
	}

	versions, err := a.versioned.ListVersions(ctx, int(idNum))
	if err != nil {
		handleV2Error(w, err, int(idNum))
		return
	}

	resp := VersionListResponse{
		ID:       int(idNum),
		Versions: make([]VersionMeta, 0, len(versions)),
	}
	for _, v := range versions {
		resp.Versions = append(resp.Versions, VersionMeta{
			Version:   v.Version,
			CreatedAt: v.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	if len(resp.Versions) == 0 {
		writeError(w, fmt.Sprintf("record %d not found", idNum), http.StatusNotFound)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}
