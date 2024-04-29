package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rynowak/ucp-dapr/pkg/resources"
)

func ReadResourceFromBody(req *http.Request) (*resources.Resource, error) {
	resource := resources.Resource{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&resource)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	return &resource, nil
}

func WriteResourceToBody(w http.ResponseWriter, statusCode int, resource *resources.Resource, headers map[string][]string) error {
	payload, err := resources.MarshalResource(*resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.WriteHeader(statusCode)
	w.Write(payload)

	return nil
}

func WriteResourceListToBody(w http.ResponseWriter, statusCode int, list []resources.Resource) error {
	payload, err := resources.MarshalResourceList(list)
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(payload)

	return nil
}

func WriteOperationToBody(w http.ResponseWriter, statusCode int, operation *resources.Operation, headers map[string][]string) error {
	payload, err := resources.MarshalOperation(*operation)
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.WriteHeader(statusCode)
	w.Write(payload)

	return nil
}

func WriteOperationListToBody(w http.ResponseWriter, statusCode int, list []resources.Operation) error {
	payload, err := resources.MarshalOperationList(list)
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(payload)

	return nil
}

func WriteErrorToBody(w http.ResponseWriter, statusCode int, errorCode string, message string) {
	e := resources.ErrorResponse{
		Error: resources.ErrorDetails{
			Code:    errorCode,
			Message: message,
		},
	}

	b, _ := json.Marshal(e)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}
