package tools

import "context"

type EventDispatcher interface {
	Handle(ctx context.Context, operation ...EventOp)
	Produce(ctx context.Context, message EventMessage)
}

type LocalDispatcher interface {
	MessagePipe(msg EventMessage)
}
