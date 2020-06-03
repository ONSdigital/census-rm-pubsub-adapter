package models

import "cloud.google.com/go/pubsub"

type OutboundMessage struct {
	EventMessage  *RmMessage
	SourceMessage *pubsub.Message
}
