package reconciler

import (
	"fmt"
	"log"
	"time"

	daprworkflow "github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

type ReconcileInput struct {
	ID  string `json:"id"`
	Uid string `json:"uid"`
}

type ReconcileEvent struct {
	OperationType string              `json:"operationType"`
	OperationID   string              `json:"operationId"`
	Generation    int64               `json:"generation"`
	Uid           string              `json:"uid"`
	Resource      *resources.Resource `json:"resource"`
}

func Reconcile(ctx *daprworkflow.WorkflowContext) (any, error) {
	input := ReconcileInput{}
	err := ctx.GetInput(&input)
	if err != nil {
		return nil, err
	}

	if input.ID == "" || input.Uid == "" {
		return nil, fmt.Errorf("invalid input, ID and Uid are required")
	}

	exists, err := resourceExists(ctx, input.ID, input.Uid)
	if err != nil {
		return nil, err
	}

	if !exists {
		// Resource was deleted or replaced. We can ignore this event.
		return nil, nil
	}

	// Wait up to an hour for activity. If we already have an event queued, this will return immediately.
	//
	// We expect an event for every change to the state of the resource. This is for safety
	// and ensures that reconciliation loops will shut themselves down if they are no longer needed.
	event := &ReconcileEvent{}
	ctx.WaitForExternalEvent("Reconcile", time.Duration(1*time.Hour)).Await(&event)

	shouldProcess, err := shouldProcessOperation(ctx, input.ID, input.Uid, event)
	if err != nil {
		return nil, err
	}

	if !shouldProcess {
		err = cancelOperation(ctx, input.ID, event.OperationID)
		if err != nil {
			return nil, err
		}

		// Start over to process more events.
		ctx.ContinueAsNew(&input, true)
		return nil, nil
	}

	result, err := processOperation(ctx, event.OperationType, event.OperationID)
	if err != nil {
		return nil, err
	}

	err = completeOperation(ctx, input.ID, event.OperationID, result)
	if err != nil {
		return nil, err
	}

	// Start over to process more events.
	ctx.ContinueAsNew(&input, true)
	return nil, nil
}

func resourceExists(ctx *daprworkflow.WorkflowContext, id string, uid string) (bool, error) {
	input := CheckResourceExistanceInput{ID: id, Uid: uid}
	output := CheckResourceExistanceOutput{}
	err := ctx.CallActivity("CheckResourceExistance", daprworkflow.ActivityInput(&input)).Await(&output)
	if err != nil {
		return false, err
	}

	return output.Exists, nil
}

func shouldProcessOperation(ctx *daprworkflow.WorkflowContext, id string, uid string, event *ReconcileEvent) (bool, error) {
	input := FetchCurrentGenerationInput{ID: id, Uid: uid}
	output := FetchCurrentGenerationOutput{}
	err := ctx.CallActivity("FetchCurrentGeneration", daprworkflow.ActivityInput(&input)).Await(&output)
	if err != nil {
		return false, err
	}

	if output.Generation > event.Generation {
		// This event is stale. We can ignore it.
		return false, nil
	}

	if output.Generation == event.Generation && output.StatusGeneration < event.Resource.SystemData.StatusGeneration {
		// This event is the operation for the current generation. We should process it.
		return true, nil
	}

	return false, nil // This is a duplicate event. We can ignore it.
}

func cancelOperation(ctx *daprworkflow.WorkflowContext, id string, operationID string) error {
	result := &CommitOperationInput{
		ID:                id,
		OperationID:       operationID,
		ProvisioningState: "Cancelled",
		Error: &resources.ErrorDetails{
			Code:    "Cancelled",
			Message: "Operation was cancelled because the resource is already up to date or another operation was started.",
		},
	}
	err := ctx.CallActivity("CommitOperation", daprworkflow.ActivityInput(result)).Await(nil)
	if err != nil {
		return err
	}

	return nil
}

func completeOperation(ctx *daprworkflow.WorkflowContext, id string, operationID string, result *Result) error {
	input := &CommitOperationInput{
		ID:          id,
		OperationID: operationID,
	}

	// We can ignore the retry field, because retries already happened before we got to this code.
	if result.Error == nil {
		input.ProvisioningState = "Succeeded"
		input.Status = result.Status
	} else if result.Error != nil {
		input.ProvisioningState = "Failed"
		input.Error = result.Error
		input.Status = result.Status
	}

	err := ctx.CallActivity("CommitOperation", daprworkflow.ActivityInput(result)).Await(nil)
	if err != nil {
		return err
	}

	return nil
}

func processOperation(ctx *daprworkflow.WorkflowContext, operationType string, operationID string) (*Result, error) {
	// Sleep for a bit to simulate work being done.
	log.Default().Printf("Starting operation: %+v", operationID)
	ctx.CreateTimer(time.Duration(1 * time.Second)).Await(nil)
	log.Default().Printf("Completed operation: %+v", operationID)

	return nil
}
