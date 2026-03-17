package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkWatcher(b *testing.B) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := b.TempDir()
	w.Add(tmpDir, false)
	w.Start()

	go func() {
		for range w.Events() {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tmpDir, "file.txt")
		os.WriteFile(filePath, []byte("test"), 0644)
		time.Sleep(600 * time.Millisecond)
	}
}

func BenchmarkWatcher_Recursive(b *testing.B) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := b.TempDir()
	for i := 0; i < 10; i++ {
		os.MkdirAll(filepath.Join(tmpDir, "subdir", string(rune(i))), 0755)
	}

	w.Add(tmpDir, true)
	w.Start()

	go func() {
		for range w.Events() {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tmpDir, "subdir", "file.txt")
		os.WriteFile(filePath, []byte("test"), 0644)
		time.Sleep(600 * time.Millisecond)
	}
}
