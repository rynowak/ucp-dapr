package subscribe

import (
	daprcommon "github.com/dapr/go-sdk/service/common"
)

var Subscriptions = []daprcommon.Subscription{
	{
		PubsubName: "pubsub",
		Topic:      "outbox",
		Route:      "/events/outbox",
	},
}
