package resources

import "time"

type Operation struct {
	OperationType string                   `json:"operationType"`
	Resource      *Resource                `json:"resource"`
	Status        *OperationStatusResource `json:"operation"`
}

// AsyncOperationStatus represents an OperationStatus resource.
type OperationStatusResource struct {
	// Id represents the async operation id.
	ID string `json:"id,omitempty"`

	// Name represents the async operation name and is usually set to the async operation id.
	Name string `json:"name,omitempty"`

	// Status represents the provisioning state of the resource.
	Status string `json:"status,omitempty"`

	// StartTime represents the async operation start time.
	StartTime time.Time `json:"startTime,omitempty"`

	// EndTime represents the async operation end time.
	EndTime *time.Time `json:"endTime,omitempty"`

	// Error represents the error occurred during provisioning.
	Error *ErrorDetails `json:"error,omitempty"`
}
