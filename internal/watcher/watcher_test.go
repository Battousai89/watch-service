package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_Add(t *testing.T) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := t.TempDir()

	err := w.Add(tmpDir, false)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	if len(w.watches) != 1 {
		t.Errorf("expected 1 watch, got %d", len(w.watches))
	}
}

func TestWatcher_Start(t *testing.T) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := t.TempDir()
	w.Add(tmpDir, false)
	w.Start()

	time.Sleep(100 * time.Millisecond)
}

func TestWatcher_EventCreate(t *testing.T) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := t.TempDir()
	w.Add(tmpDir, false)
	w.Start()

	// Ждем завершения инициализации
	time.Sleep(100 * time.Millisecond)

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	select {
	case event := <-w.Events():
		if event.Type != EventCreate {
			t.Errorf("expected CREATE event, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for create event")
	}
}

func TestWatcher_EventWrite(t *testing.T) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := t.TempDir()

	// Создаем файл до запуска watcher
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	w.Add(tmpDir, false)
	w.Start()

	// Ждем завершения инициализации и первого сканирования
	time.Sleep(700 * time.Millisecond)

	// Модифицируем файл
	os.WriteFile(testFile, []byte("modified"), 0644)

	select {
	case event := <-w.Events():
		if event.Type != EventWrite {
			t.Errorf("expected WRITE event, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for write event")
	}
}

func TestWatcher_EventRemove(t *testing.T) {
	w := NewWatcher().(*FileWatcher)
	defer w.Close()

	tmpDir := t.TempDir()

	// Создаем файл до запуска watcher
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	w.Add(tmpDir, false)
	w.Start()

	// Ждем завершения инициализации
	time.Sleep(700 * time.Millisecond)

	os.Remove(testFile)

	select {
	case event := <-w.Events():
		if event.Type != EventRemove {
			t.Errorf("expected REMOVE event, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for remove event")
	}
}
