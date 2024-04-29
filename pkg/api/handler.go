package api

import daprclient "github.com/dapr/go-sdk/client"

type Handler struct {
	Dapr             daprclient.Client
	StateStoreName   string
	PubSubName       string
	OutboxPubSubName string
	ResourceType     string
}
