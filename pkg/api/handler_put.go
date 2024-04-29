package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

func (h *Handler) PutHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id, scope, resourceType, name := resources.ParseResource(r.URL.Path)

	input, err := ReadResourceFromBody(r)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	resource, etag, err := db.ReadResourceFromStateStore(r.Context(), id)
	if err != nil {
		WriteErrorToBody(w, http.StatusInternalServerError, "Internal", err.Error())
		return
	}

	if resource == nil {
		resource = &resources.Resource{
			ID:         id,
			Name:       name,
			Type:       resourceType,
			Scope:      scope,
			Properties: input.Properties,
			SystemData: resources.SystemData{
				Generation: 0,
				Uid:        uuid.New().String(),
			},
		}
	} else if resource.SystemData.IsDeleting {
		WriteErrorToBody(w, http.StatusConflict, "Conflict", "resource is being deleted")
		return
	}

	// Update to this resource is accepted. Commit the change and start the reconciliation process.

	resource.SystemData.Generation = resource.SystemData.Generation + 1
	resource.SetProvisioningStateIfTerminal("Updating")

	operationName := uuid.NewString()
	operation := &resources.Operation{
		OperationType: strings.ToUpper(resourceType) + "/PUT",
		Status: &resources.OperationStatusResource{
			ID:        resources.ParsePlaneScope(id) + "/providers/" + resources.ParseNamespace(id) + "/operationStatuses/" + operationName,
			Name:      operationName,
			Status:    "Updating",
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
