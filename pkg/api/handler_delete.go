package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id, _, resourceType, _ := resources.ParseResource(r.URL.Path)
	resource, etag, err := db.ReadResourceFromStateStore(r.Context(), id)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	if resource == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Update to this resource is accepted. Commit the change and start the reconciliation process.
	resource.SystemData.Generation = resource.SystemData.Generation + 1
	resource.SetProvisioningStateIfTerminal("Deleting")

	operationName := uuid.NewString()
	operation := &resources.Operation{
		OperationType: strings.ToUpper(resourceType) + "/DELETE",
		Status: &resources.OperationStatusResource{
			ID:        resources.ParsePlaneScope(id) + "/providers/" + resources.ParseNamespace(id) + "/operationStatuses/" + operationName,
			Name:      operationName,
			Status:    "Deleting",
			StartTime: time.Now().UTC(),
		},
		Resource: resource,
	}

	err = db.WriteResourceAndOperationToStateStore(r.Context(), true, resource, operation, etag)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	err = WriteResourceToBody(w, 200, resource, map[string][]string{
		"Location": {operation.Status.ID},
	})
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}
}
