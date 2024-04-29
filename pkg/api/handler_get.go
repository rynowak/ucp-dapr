package api

import (
	"net/http"
	"strings"

	"github.com/rynowak/ucp-dapr/pkg/db"
)

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := strings.ToLower(r.URL.Path)
	resource, _, err := db.ReadResourceFromStateStore(r.Context(), id)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	} else if resource == nil {
		WriteErrorToBody(w, http.StatusNotFound, "NotFound", "resource not found")
		return
	}

	err = WriteResourceToBody(w, http.StatusOK, resource, nil)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}
}
