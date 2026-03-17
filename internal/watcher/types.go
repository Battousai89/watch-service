package watcher

import (
	"os"
	"time"
)

type EventType int

const (
	EventCreate EventType = iota
	EventWrite
	EventRemove
	EventRename
	EventChmod
)

func (e EventType) String() string {
	switch e {
	case EventCreate:
		return "CREATE"
	case EventWrite:
		return "WRITE"
	case EventRemove:
		return "REMOVE"
	case EventRename:
		return "RENAME"
	case EventChmod:
		return "CHMOD"
	default:
		return "UNKNOWN"
	}
}

type FileEvent struct {
	Path string
	Name string
	Type EventType
	Time time.Time
}

func (e FileEvent) FullPath() string {
	if e.Name != "" && e.Path != "" {
		return e.Path + string(os.PathSeparator) + e.Name
	}
	if e.Path != "" {
		return e.Path
	}
	return e.Name
}

type WatcherInterface interface {
	Add(path string, recursive bool) error
	Start()
	Events() <-chan FileEvent
	Errors() <-chan error
	Close() error
}
