package reconciler

import (
	"time"

	"github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

type CommitOperationInput struct {
	ID                string                  `json:"id"`
	OperationID       string                  `json:"operationId"`
	ProvisioningState string                  `json:"provisioningState"`
	Status            any                     `json:"status,omitempty"`
	Error             *resources.ErrorDetails `json:"error,omitempty"`
}

type CommitOperationOutput struct {
}

func CommitOperation(ctx workflow.ActivityContext) (any, error) {
	input := CommitOperationInput{}
	err := ctx.GetInput(&input)
	if err != nil {
		return "", err
	}

	err = commitOperation(ctx, &input)
	if err != nil {
		return nil, err
	}

	return &CommitOperationOutput{}, nil
}

func commitOperation(ctx workflow.ActivityContext, input *CommitOperationInput) error {
	resource, etag, err := db.ReadResourceFromStateStore(ctx.Context(), input.ID)
	if err != nil {
		return err
	}

	if resource == nil {
		return nil
	}

	operation, _, err := db.ReadOperationFromStateStore(ctx.Context(), input.OperationID)
	if err != nil {
		return err
	}

	if operation == nil {
		return nil
	}

	resource.SetProvisioningState(input.ProvisioningState)
	resource.SystemData.StatusGeneration = operation.Resource.SystemData.Generation

	endTime := time.Now().UTC()
	operation.Status.Status = input.ProvisioningState
	operation.Status.EndTime = &endTime
	operation.Status.Error = input.Error

	err = db.WriteResourceAndOperationToStateStore(ctx.Context(), false, resource, operation, etag)
	if err != nil {
		return err
	}

	return nil
}
