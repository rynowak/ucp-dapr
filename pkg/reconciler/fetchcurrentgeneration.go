package reconciler

import (
	"github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/db"
)

type FetchCurrentGenerationInput struct {
	ID  string `json:"id"`
	Uid string `json:"uid"`
}

type FetchCurrentGenerationOutput struct {
	Generation       int64 `json:"generation"`
	StatusGeneration int64 `json:"statusGeneration"`
}

func FetchCurrentGeneration(ctx workflow.ActivityContext) (any, error) {
	input := FetchCurrentGenerationInput{}
	err := ctx.GetInput(&input)
	if err != nil {
		return "", err
	}

	resource, _, err := db.ReadResourceFromStateStore(ctx.Context(), input.ID)
	if err != nil {
		return nil, err
	}

	if resource == nil {
		return &FetchCurrentGenerationOutput{Generation: 0, StatusGeneration: 0}, nil
	}

	return &FetchCurrentGenerationOutput{Generation: resource.SystemData.Generation, StatusGeneration: resource.SystemData.StatusGeneration}, nil
}
