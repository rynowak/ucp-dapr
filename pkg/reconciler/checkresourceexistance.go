package reconciler

import (
	"github.com/dapr/go-sdk/workflow"
	"github.com/rynowak/ucp-dapr/pkg/db"
)

type CheckResourceExistanceInput struct {
	ID  string `json:"id"`
	Uid string `json:"uid"`
}

type CheckResourceExistanceOutput struct {
	Exists bool `json:"exists"`
}

func CheckResourceExistance(ctx workflow.ActivityContext) (any, error) {
	input := CheckResourceExistanceInput{}
	err := ctx.GetInput(&input)
	if err != nil {
		return "", err
	}

	resource, _, err := db.ReadResourceFromStateStore(ctx.Context(), input.ID)
	if err != nil {
		return nil, err
	}

	return &CheckResourceExistanceOutput{Exists: resource != nil && resource.SystemData.Uid == input.Uid}, nil
}
