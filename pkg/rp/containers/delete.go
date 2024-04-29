package containers

import (
	"log"
	"time"

	daprworkflow "github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/reconciler"
)

func ContainerDelete(ctx *daprworkflow.WorkflowContext) (any, error) {
	workitem := reconciler.WorkItem{}
	err := ctx.GetInput(&workitem)
	if err != nil {
		return nil, err
	}

	return containerDelete(ctx, &workitem)
}

func containerDelete(ctx *daprworkflow.WorkflowContext, workitem *reconciler.WorkItem) (*reconciler.Result, error) {
	// Sleep for a bit to simulate work being done.
	log.Default().Printf("Starting operation: %v %v", workitem.OperationType, workitem.OperationID)
	ctx.CreateTimer(time.Duration(1 * time.Second)).Await(nil)
	log.Default().Printf("Completed operation: %v %v", workitem.OperationType, workitem.OperationID)

	return &reconciler.Result{}, nil
}
