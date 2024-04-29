package resources

import (
	"encoding/json"

	daprclient "github.com/dapr/go-sdk/client"
)

type Resource struct {
	Name       string         `json:"name"`
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Scope      string         `json:"scope"`
	Properties map[string]any `json:"properties,omitempty"`
	Status     map[string]any `json:"status,omitempty"`
	SystemData SystemData     `json:"systemData"`
}

func (r Resource) GetProvisioningState() string {
	if r.Properties == nil {
		return ""
	}

	obj, ok := r.Properties["provisioningState"]
	if !ok {
		return ""
	}

	value, ok := obj.(string)
	if !ok {
		return ""
	}

	return value
}

func (r Resource) SetProvisioningState(value string) {
	if r.Properties == nil {
		r.Properties = map[string]any{}
	}

	r.Properties["provisioningState"] = value
}

func (r Resource) SetProvisioningStateIfTerminal(value string) {
	current := r.GetProvisioningState()
	if current == "" || current == "Succeeded" || current == "Failed" || current == "Canceled" {
		r.SetProvisioningState(value)
	}
}

type SystemData struct {
	Generation       int64  `json:"generation"`
	StatusGeneration int64  `json:"statusGeneration"`
	Uid              string `json:"uid"`
	IsDeleting       bool   `json:"isDeleting"`
}

func MarshalResource(r Resource) ([]byte, error) {
	return json.Marshal(r)
}

func MarshalResourceList(resources []Resource) ([]byte, error) {
	wrapper := struct {
		Value []Resource `json:"value"`
	}{resources}
	return json.Marshal(wrapper)
}

func UnmarshalResource(data []byte) (Resource, error) {
	r := Resource{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func UnmarshalResourceQuery(response *daprclient.QueryResponse) ([]Resource, error) {
	resources := []Resource{}
	for _, result := range response.Results {
		r, err := UnmarshalResource(result.Value)
		if err != nil {
			return nil, err
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func MarshalOperation(op Operation) ([]byte, error) {
	return json.Marshal(op.Status)
}

func MarshalOperationList(ops []Operation) ([]byte, error) {
	wrapper := struct {
		Value []OperationStatusResource `json:"value"`
	}{}

	for _, op := range ops {
		wrapper.Value = append(wrapper.Value, *op.Status)
	}

	return json.Marshal(wrapper)
}

func UnmarshalOperation(data []byte) (Operation, error) {
	r := Operation{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func UnmarshalOperationQuery(response *daprclient.QueryResponse) ([]Operation, error) {
	operations := []Operation{}
	for _, result := range response.Results {
		r, err := UnmarshalOperation(result.Value)
		if err != nil {
			return nil, err
		}

		operations = append(operations, r)
	}

	return operations, nil
}
