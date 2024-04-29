package reconciler

import "github.com/rynowak/ucp-dapr/pkg/resources"

type WorkItem struct {
	OperationID   string `json:"operationId"`
	OperationType string `json:"operationType"`
	Resource      string `json:"resource"`
}

type Result struct {
	Retry  bool                    `json:"retry,omitempty"`
	Error  *resources.ErrorDetails `json:"error,omitempty"`
	Status any                     `json:"status,omitempty"`
}
