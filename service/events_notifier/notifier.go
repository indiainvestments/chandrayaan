package events_notifier

type EventsNotifier interface {
	Register(c chan interface{})
	Unregister(c chan interface{})
}
