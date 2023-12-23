package tools

type AggregateRoot struct {
	events []RootHelper
}

func (a *AggregateRoot) WithEvent(events ...RootHelper) {
	a.events = append(a.events, events...)
}
func (a *AggregateRoot) GetEvents() []RootHelper {
	return a.events
}

type RootHelper interface {
	ServiceTarget() string
}
