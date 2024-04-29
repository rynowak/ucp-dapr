package reconciler

import (
	"github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/db"
	"github.com/rynowak/ucp-dapr/pkg/resources"
)

type SaveResourceInput struct {
	Resource *resources.Resource `json:"resource"`
	ETag     *string             `json:"etag"`
}

type SaveResourceOutput struct {
}

func SaveResource(ctx workflow.ActivityContext) (any, error) {
	input := SaveResourceInput{}
	err := ctx.GetInput(&input)
	if err != nil {
		return "", err
	}

	err = db.WriteResourceToStateStore(ctx.Context(), input.Resource, input.ETag)
	if err != nil {
		return nil, err
	}

	return &SaveResourceOutput{}, nil
}
