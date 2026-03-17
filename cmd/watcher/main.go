package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"watch-service/internal/config"
	"watch-service/internal/debouncer"
	"watch-service/internal/logger"
	"watch-service/internal/runner"
	"watch-service/internal/watcher"
)

func main() {
	configPath := flag.String("config", "config.toml", "path to config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	log := logger.Default()
	if *debug {
		log = logger.New(os.Stderr, logger.LevelDebug, "")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("load config: %v", err)
		os.Exit(1)
	}

	log.Info("config loaded: delay=%v, max_parallel=%d, watches=%d",
		cfg.Delay, cfg.MaxParallel, len(cfg.Watch))

	w := watcher.NewWatcher()
	defer w.Close()

	for _, watchCfg := range cfg.Watch {
		if err := w.Add(watchCfg.Path, watchCfg.Recursive); err != nil {
			log.Error("add watch %s: %v", watchCfg.Path, err)
			continue
		}
		absPath, _ := filepath.Abs(watchCfg.Path)
		log.Info("watching: %s (recursive=%v)", absPath, watchCfg.Recursive)
	}

	w.Start()

	debounce := debouncer.NewBatchDebounce(cfg.Delay)
	defer debounce.Close()

	run := runner.NewCommandRunner(cfg.MaxParallel)

	go func() {
		for event := range w.Events() {
			log.Debug("event received: %s %s", event.Type, event.FullPath())
			debounce.Add(event)
		}
	}()

	go func() {
		for batch := range debounce.Events() {
			if len(batch) == 0 {
				continue
			}

			firstEvent := batch[0]
			watchCfg := findWatchConfig(firstEvent.Path, cfg.Watch)
			if watchCfg == nil || watchCfg.Command == "" {
				continue
			}

			var files []string
			for _, event := range batch {
				files = append(files, event.FullPath())
			}

			env := map[string]string{
				"WATCH_FILES":      strings.Join(files, ";"),
				"WATCH_FILE_COUNT": fmt.Sprintf("%d", len(files)),
				"WATCH_EVENT_TYPE": firstEvent.Type.String(),
			}

			args := replacePlaceholders(watchCfg.Args, firstEvent)
			args = replaceFilesPlaceholder(args, files)

			log.Info("executing: %s (files: %d)", watchCfg.Command, len(files))

			run.Run(runner.CommandRequest{
				Cmd:     watchCfg.Command,
				Args:    args,
				Env:     env,
				Timeout: watchCfg.Timeout,
				WorkDir: watchCfg.Path,
			})
		}
	}()

	go func() {
		for err := range w.Errors() {
			log.Error("watcher error: %v", err)
		}
	}()

	log.Info("watcher started, press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down...")

	w.Close()
	run.Close()
	debounce.Close()
}

func findWatchConfig(eventPath string, watches []config.WatchConfig) *config.WatchConfig {
	for i := range watches {
		watchPath := watches[i].Path
		if !filepath.IsAbs(watchPath) {
			if abs, err := filepath.Abs(watchPath); err == nil {
				watchPath = abs
			}
		}

		if strings.HasPrefix(eventPath, watchPath) {
			return &watches[i]
		}
	}
	return nil
}

func replacePlaceholders(args []string, event watcher.FileEvent) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		arg = strings.ReplaceAll(arg, "{file}", event.FullPath())
		arg = strings.ReplaceAll(arg, "{path}", event.Path)
		arg = strings.ReplaceAll(arg, "{name}", event.Name)
		result[i] = arg
	}
	return result
}

func replaceFilesPlaceholder(args []string, files []string) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		arg = strings.ReplaceAll(arg, "{files}", strings.Join(files, " "))
		result[i] = arg
	}
	return result
}

func init() {
	if err := checkArgs(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func checkArgs() error {
	return nil
}
