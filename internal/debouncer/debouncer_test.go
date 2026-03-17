package debouncer

import (
	"testing"
	"time"

	"watch-service/internal/watcher"
)

func TestBatchDebouncer_Add(t *testing.T) {
	d := NewBatchDebounce(100 * time.Millisecond)
	defer d.Close()

	event := watcher.FileEvent{
		Path: "/test",
		Name: "file.txt",
		Type: watcher.EventWrite,
		Time: time.Now(),
	}

	d.Add(event)

	select {
	case batch := <-d.Events():
		if len(batch) != 1 {
			t.Errorf("expected 1 event, got %d", len(batch))
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for event")
	}
}

func TestBatchDebouncer_Batch(t *testing.T) {
	d := NewBatchDebounce(100 * time.Millisecond)
	defer d.Close()

	for i := 0; i < 5; i++ {
		d.Add(watcher.FileEvent{
			Path: "/test",
			Name: "file.txt",
			Type: watcher.EventWrite,
			Time: time.Now(),
		})
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case batch := <-d.Events():
		if len(batch) != 5 {
			t.Errorf("expected 5 events, got %d", len(batch))
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for batch")
	}
}

func TestBatchDebouncer_Close(t *testing.T) {
	d := NewBatchDebounce(100 * time.Millisecond)

	d.Add(watcher.FileEvent{
		Path: "/test",
		Name: "file.txt",
		Type: watcher.EventWrite,
		Time: time.Now(),
	})

	d.Close()
	d.Close() // double close should not panic
}
