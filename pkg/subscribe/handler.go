package subscribe

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	daprclient "github.com/dapr/go-sdk/client"
	daprcommon "github.com/dapr/go-sdk/service/common"
	"github.com/rynowak/ucp-dapr/pkg/reconciler"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

var Client daprclient.Client

func ResourceEvent(ctx context.Context, wrapper *daprcommon.TopicEvent) (bool, error) {
	if !isOperation(wrapper) {
		return false, nil // Skipping because this isn't an operation.
	}

	event, err := unmarshalOperation(wrapper)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if event.Resource == nil {
		return false, nil
	}

	retry, err := resourceEvent(ctx, event)
	if err != nil {
		log.Default().Printf("Failed to process event: %v", err)
	}

	return retry, err
}

func resourceEvent(ctx context.Context, operation *resources.Operation) (bool, error) {
	log.Default().Printf("Received event for operation: %v", operation.Status.ID)

	_, err := Client.StartWorkflowBeta1(ctx, &daprclient.StartWorkflowRequest{
		WorkflowName: "ContainerReconcile",
		Input:        reconciler.ReconcileInput{ID: operation.Resource.ID, Uid: operation.Resource.SystemData.Uid},
		InstanceID:   fmt.Sprintf("reconcile-%s", operation.Resource.SystemData.Uid),
	})
	if isWorkflowAlreadyExistsErr(err) {
		// No need to retry, we just want to make sure the workload actually exists.
	} else if err != nil {
		return true, fmt.Errorf("failed to start workflow: %w", err)
	}

	event := reconciler.ReconcileEvent{
		OperationType: operation.OperationType,
		OperationID:   operation.Status.ID,
		Generation:    operation.Resource.SystemData.Generation,
		Uid:           operation.Resource.SystemData.Uid,
		Resource:      operation.Resource,
	}
	err = Client.RaiseEventWorkflowBeta1(ctx, &daprclient.RaiseEventWorkflowRequest{
		InstanceID: fmt.Sprintf("reconcile-%s", operation.Resource.SystemData.Uid),
		EventName:  "Reconcile",
		EventData:  event,
	})
	if err != nil {
		return true, fmt.Errorf("failed to send workflow event: %w", err)
	}

	return false, nil
}

func isWorkflowAlreadyExistsErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "an active workflow with ID")
}

func isOperation(wrapper *daprcommon.TopicEvent) bool {
	m, ok := wrapper.Data.(map[string]any)
	if !ok {
		return false
	}

	_, ok = m["operationType"]
	return ok
}

func unmarshalOperation(wrapper *daprcommon.TopicEvent) (*resources.Operation, error) {
	bs, err := json.Marshal(wrapper.Data)
	if err != nil {
		return nil, err
	}

	operation := &resources.Operation{}
	err = json.Unmarshal(bs, operation)
	if err != nil {
		return operation, err
	}

	return operation, nil
}
