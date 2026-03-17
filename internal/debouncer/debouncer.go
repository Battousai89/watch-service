package debouncer

import (
	"sync"
	"time"

	"github.com/watch-service/internal/watcher"
)

type BatchDebouncer struct {
	mu       sync.Mutex
	delay    time.Duration
	timer    *time.Timer
	events   []watcher.FileEvent
	out      chan []watcher.FileEvent
	done     chan struct{}
	closed   bool
}

func NewBatchDebounce(delay time.Duration) *BatchDebouncer {
	d := &BatchDebouncer{
		delay:  delay,
		events: make([]watcher.FileEvent, 0),
		out:    make(chan []watcher.FileEvent, 100),
		done:   make(chan struct{}),
	}
	return d
}

func (d *BatchDebouncer) Add(event watcher.FileEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	d.events = append(d.events, event)

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, d.flush)
}

func (d *BatchDebouncer) flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.events) == 0 {
		return
	}

	select {
	case d.out <- d.events:
		d.events = make([]watcher.FileEvent, 0)
	default:
	}
}

func (d *BatchDebouncer) Events() <-chan []watcher.FileEvent {
	return d.out
}

func (d *BatchDebouncer) Close() {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	d.closed = true
	d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	if len(d.events) > 0 {
		d.flush()
	}

	close(d.done)
	close(d.out)
}
