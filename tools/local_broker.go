package tools

import (
	"context"
	"log/slog"
)

type LocalBroker struct {
	ch                 chan EventMessage
	serviceDispatchers map[string]LocalDispatcher
}

func NewLocalBroker() *LocalBroker {
	return &LocalBroker{
		ch:                 make(chan EventMessage),
		serviceDispatchers: make(map[string]LocalDispatcher),
	}
}
func (l *LocalBroker) RegistreService(serviceName string, dispathcer LocalDispatcher) {
	l.serviceDispatchers[serviceName] = dispathcer
}
func (l *LocalBroker) SendEvent(msg EventMessage) {
	go func() {
		l.ch <- msg
	}()

}
func (l *LocalBroker) Run(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case message := <-l.ch:
				if dispatcher, ok := l.serviceDispatchers[message.GetReceiver()]; ok {
					dispatcher.MessagePipe(message)
				}
			case <-ctx.Done():
				slog.Info("Broker stoped")
				return
			}
		}

	}(ctx)
}
