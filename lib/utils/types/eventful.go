package types

type Watcher[K any, V any] <-chan WatchMsg[K, V]

type WatchMsg[K any, V any] struct {
	Event EventType
	Item[K, V]
}

type EventType string

const (
	PutEvent    = "PUT"
	DeleteEvent = "DELETE"
)
