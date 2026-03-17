package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileState struct {
	modTime time.Time
	size    int64
}

type watchDir struct {
	path      string
	recursive bool
}

type FileWatcher struct {
	mu       sync.Mutex
	watches  []watchDir
	files    map[string]fileState
	events   chan FileEvent
	errors   chan error
	done     chan struct{}
	wg       sync.WaitGroup
	interval time.Duration
	closed   bool
}

func NewWatcher() WatcherInterface {
	return &FileWatcher{
		files:    make(map[string]fileState),
		events:   make(chan FileEvent, 4096),
		errors:   make(chan error, 100),
		done:     make(chan struct{}),
		interval: 500 * time.Millisecond,
	}
}

func (w *FileWatcher) Add(path string, recursive bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	w.watches = append(w.watches, watchDir{path: absPath, recursive: recursive})
	return nil
}

func (w *FileWatcher) Start() {
	w.mu.Lock()
	for _, watch := range w.watches {
		filepath.Walk(watch.path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			w.files[filePath] = fileState{modTime: info.ModTime(), size: info.Size()}
			return nil
		})
	}
	w.mu.Unlock()
	
	w.wg.Add(1)
	go w.run()
}

func (w *FileWatcher) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			w.check()
		}
	}
}

func (w *FileWatcher) check() {
	w.mu.Lock()
	defer w.mu.Unlock()

	currentFiles := make(map[string]bool)

	for _, watch := range w.watches {
		w.scanDir(watch.path, watch.recursive, currentFiles, true)
	}

	for filePath := range w.files {
		if !currentFiles[filePath] {
			delete(w.files, filePath)
			w.sendEvent(filePath, EventRemove)
		}
	}
}

func (w *FileWatcher) scanDir(path string, recursive bool, currentFiles map[string]bool, sendEvents bool) {
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		currentFiles[filePath] = true

		newState := fileState{
			modTime: info.ModTime(),
			size:    info.Size(),
		}

		oldState, exists := w.files[filePath]
		if !exists {
			w.files[filePath] = newState
			if sendEvents {
				w.sendEvent(filePath, EventCreate)
			}
		} else if oldState.modTime != newState.modTime || oldState.size != newState.size {
			w.files[filePath] = newState
			if sendEvents {
				w.sendEvent(filePath, EventWrite)
			}
		}

		return nil
	})
}

func (w *FileWatcher) sendEvent(fullPath string, eventType EventType) {
	dir := filepath.Dir(fullPath)
	name := filepath.Base(fullPath)

	select {
	case w.events <- FileEvent{
		Path: dir,
		Name: name,
		Type: eventType,
		Time: time.Now(),
	}:
	default:
	}
}

func (w *FileWatcher) Events() <-chan FileEvent {
	return w.events
}

func (w *FileWatcher) Errors() <-chan error {
	return w.errors
}

func (w *FileWatcher) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()
	
	close(w.done)
	w.wg.Wait()
	close(w.events)
	close(w.errors)
	return nil
}
