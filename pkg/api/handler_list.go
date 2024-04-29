package api

import (
	"net/http"

	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, scope, resourceType := resources.ParseCollection(r.URL.Path)
	results, err := db.ListResourcesInStateStore(r.Context(), scope, resourceType)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	payload, err := resources.MarshalResourceList(results)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
