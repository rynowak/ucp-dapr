package api

import (
	"net/http"

	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

func (h *Handler) OperationStatusListHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, scope, resourceType := resources.ParseCollection(r.URL.Path)
	results, err := db.ListOperationsInStateStore(r.Context(), scope, resourceType)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	err = WriteOperationListToBody(w, http.StatusOK, results)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}
}
