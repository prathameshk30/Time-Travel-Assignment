package api

import (
	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/service"
)

type APIV2 struct {
	versioned service.VersionedRecordService
}

func NewAPIV2(versioned service.VersionedRecordService) *APIV2 {
	return &APIV2{versioned: versioned}
}

func (a *APIV2) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").Methods("GET").HandlerFunc(a.GetRecord)
	routes.Path("/records/{id}").Methods("POST").HandlerFunc(a.PostRecord)
	routes.Path("/records/{id}/versions").Methods("GET").HandlerFunc(a.ListVersions)
	routes.Path("/records/{id}/versions/{version}").Methods("GET").HandlerFunc(a.GetRecordVersion)
}
