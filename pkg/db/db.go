package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	daprclient "github.com/dapr/go-sdk/client"
	"github.com/rynowak/ucp-dapr/pkg/resources"
	"google.golang.org/grpc/status"
)

const (
	stateStoreName = "statestore"
)

var Client daprclient.Client

func ReadResourceFromStateStore(ctx context.Context, id string) (*resources.Resource, *string, error) {
	response, err := Client.GetState(ctx, stateStoreName, strings.ToLower(id), map[string]string{
		"contentType": "application/json",
	})

	st, ok := status.FromError(err)
	if err != nil && ok {
		return nil, nil, fmt.Errorf("failed to get resource: %w + %v", err, st)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup resource: %w", err)
	}

	if len(response.Value) == 0 {
		return nil, nil, nil
	}

	resource := resources.Resource{}
	err = json.Unmarshal(response.Value, &resource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal resource data: %w", err)
	}

	return &resource, &response.Etag, nil
}

func WriteResourceToStateStore(ctx context.Context, resource *resources.Resource, etag *string) error {
	rb, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource data: %w", err)
	}

	if etag == nil {
		err = Client.SaveState(ctx, stateStoreName, strings.ToLower(resource.ID), rb, map[string]string{
			"contentType": "application/json",
		})
	} else {
		err = Client.SaveStateWithETag(ctx, stateStoreName, strings.ToLower(resource.ID), rb, *etag, map[string]string{
			"contentType": "application/json",
		})
	}
	if err != nil {
		return fmt.Errorf("failed to save resource data: %w", err)
	}

	return nil
}

func WriteResourceAndOperationToStateStore(ctx context.Context, notify bool, resource *resources.Resource, operation *resources.Operation, etag *string) error {
	rb, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource data: %w", err)
	}

	ob, err := json.Marshal(operation)
	if err != nil {
		return fmt.Errorf("failed to marshal operation data: %w", err)
	}

	resourceItem := &daprclient.StateOperation{
		Type: daprclient.StateOperationTypeUpsert,
		Item: &daprclient.SetStateItem{
			Key:   strings.ToLower(resource.ID),
			Value: rb,
		},
	}
	if etag != nil {
		resourceItem.Item.Etag = &daprclient.ETag{Value: *etag}
	}

	operationItem := &daprclient.StateOperation{
		Type: daprclient.StateOperationTypeUpsert,
		Item: &daprclient.SetStateItem{
			Key:      strings.ToLower(operation.Status.ID),
			Value:    ob,
			Metadata: map[string]string{"ttlInSeconds": fmt.Sprintf("%v", (48 * time.Hour).Seconds())},
		},
	}

	statestore := "statestore"
	if notify {
		statestore = "statestore-outbox"
	}

	metadata := map[string]string{"contentType": "application/json"}
	err = Client.ExecuteStateTransaction(ctx, statestore, metadata, []*daprclient.StateOperation{resourceItem, operationItem})
	if err != nil {
		return fmt.Errorf("failed to save resource data: %w", err)
	}

	return nil
}

func DeleteResourceFromStateStore(ctx context.Context, id string, etag *string) error {
	var err error
	if etag == nil {
		err = Client.DeleteState(ctx, stateStoreName, strings.ToLower(id), map[string]string{
			"contentType": "application/json",
		})
	} else {
		err = Client.DeleteStateWithETag(ctx, stateStoreName, strings.ToLower(id), &daprclient.ETag{Value: *etag}, map[string]string{
			"contentType": "application/json",
		}, nil)
	}
	if err != nil {
		return fmt.Errorf("failed to save resource data: %w", err)
	}

	return nil
}

func ListResourcesInStateStore(ctx context.Context, scope string, resourceType string) ([]resources.Resource, error) {
	query := `{
		"filter": {
			"EQ": { "scope": "%s" },
			"EQ": { "type": "%s" }
		},
		"sort": [
			{
				"key": "name",
				"order": "ASC"
			}
		]
	}`
	query = fmt.Sprintf(query, scope, resourceType)

	response, err := Client.QueryStateAlpha1(ctx, stateStoreName, query, map[string]string{
		"contentType": "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query resource data: %w", err)
	}

	return resources.UnmarshalResourceQuery(response)
}

func ReadOperationFromStateStore(ctx context.Context, id string) (*resources.Operation, *string, error) {
	response, err := Client.GetState(ctx, stateStoreName, strings.ToLower(id), map[string]string{
		"contentType": "application/json",
	})

	st, ok := status.FromError(err)
	if err != nil && ok {
		return nil, nil, fmt.Errorf("failed to get resource: %w + %v", err, st)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup resource: %w", err)
	}

	if len(response.Value) == 0 {
		return nil, nil, nil
	}

	operation := resources.Operation{}
	err = json.Unmarshal(response.Value, &operation)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal resource data: %w", err)
	}

	return &operation, &response.Etag, nil
}

func WriteOperationToStateStore(ctx context.Context, operation *resources.Operation, etag *string) error {
	rb, err := json.Marshal(operation)
	if err != nil {
		return fmt.Errorf("failed to marshal resource data: %w", err)
	}

	if etag == nil {
		err = Client.SaveState(ctx, stateStoreName, strings.ToLower(operation.Status.ID), rb, map[string]string{
			"contentType": "application/json",
		})
	} else {
		err = Client.SaveStateWithETag(ctx, stateStoreName, strings.ToLower(operation.Status.ID), rb, *etag, map[string]string{
			"contentType": "application/json",
		})
	}
	if err != nil {
		return fmt.Errorf("failed to save resource data: %w", err)
	}

	return nil
}

func ListOperationsInStateStore(ctx context.Context, scope string, namespace string) ([]resources.Operation, error) {
	query := `{
		"filter": {
			"EQ": { "scope": "%s" },
			"EQ": { "type": "%s" }
		},
		"sort": [
			{
				"key": "name",
				"order": "ASC"
			}
		]
	}`
	query = fmt.Sprintf(query, scope, fmt.Sprintf("%s/operations", namespace))

	response, err := Client.QueryStateAlpha1(ctx, stateStoreName, query, map[string]string{
		"contentType": "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query resource data: %w", err)
	}

	return resources.UnmarshalOperationQuery(response)
}
