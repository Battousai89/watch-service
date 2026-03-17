package debouncer

import (
	"testing"
	"time"

	"watch-service/internal/watcher"
)

func BenchmarkBatchDebouncer(b *testing.B) {
	d := NewBatchDebounce(10 * time.Millisecond)
	defer d.Close()

	go func() {
		for range d.Events() {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Add(watcher.FileEvent{
			Path: "/test",
			Name: "file.txt",
			Type: watcher.EventWrite,
			Time: time.Now(),
		})
	}
}

func BenchmarkBatchDebouncer_Parallel(b *testing.B) {
	d := NewBatchDebounce(10 * time.Millisecond)
	defer d.Close()

	go func() {
		for range d.Events() {
		}
	}()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.Add(watcher.FileEvent{
				Path: "/test",
				Name: "file.txt",
				Type: watcher.EventWrite,
				Time: time.Now(),
			})
		}
	})
}
