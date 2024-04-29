package api

import (
	"net/http"
	"strings"

	"github.com/rynowak/ucp-dapr/pkg/db"
)

func (h *Handler) OperationStatusGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := strings.ToLower(r.URL.Path)
	operation, _, err := db.ReadOperationFromStateStore(r.Context(), id)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	} else if operation == nil {
		WriteErrorToBody(w, http.StatusNotFound, "NotFound", "resource not found")
		return
	}

	err = WriteOperationToBody(w, http.StatusOK, operation, nil)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}
}
