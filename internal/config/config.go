package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type WatchConfig struct {
	Path      string
	Recursive bool
	Command   string
	Args      []string
	Timeout   time.Duration
}

type Config struct {
	Delay       time.Duration
	MaxParallel int
	Watch       []WatchConfig
}

type parser struct {
	lines []string
	pos   int
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	p := &parser{
		lines: strings.Split(string(data), "\n"),
	}

	cfg := &Config{
		Delay:       500 * time.Millisecond,
		MaxParallel: 0,
	}

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		if strings.HasPrefix(line, "[[watch]]") {
			watch, err := p.parseWatch()
			if err != nil {
				return nil, err
			}
			cfg.Watch = append(cfg.Watch, watch)
			continue
		}

		if strings.Contains(line, "=") {
			key, value, _ := strings.Cut(line, "=")
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)

			switch key {
			case "delay":
				cfg.Delay, err = parseDuration(value)
				if err != nil {
					return nil, fmt.Errorf("parse delay: %w", err)
				}
			case "max_parallel":
				cfg.MaxParallel, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse max_parallel: %w", err)
				}
			}
		}

		p.pos++
	}

	return cfg, nil
}

func (p *parser) parseWatch() (WatchConfig, error) {
	watch := WatchConfig{
		Recursive: false,
		Timeout:   30 * time.Second,
	}

	p.pos++

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "" || strings.HasPrefix(line, "#") {
			p.pos++
			continue
		}

		if strings.HasPrefix(line, "[[") || strings.HasPrefix(line, "[") {
			break
		}

		if !strings.Contains(line, "=") {
			p.pos++
			continue
		}

		key, value, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "path":
			watch.Path = unquote(value)
		case "recursive":
			watch.Recursive = value == "true"
		case "command":
			watch.Command = unquote(value)
		case "args":
			watch.Args = parseArray(value)
		case "timeout":
			d, err := parseDuration(value)
			if err != nil {
				return watch, fmt.Errorf("parse timeout: %w", err)
			}
			watch.Timeout = d
		}

		p.pos++
	}

	return watch, nil
}

func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

func parseArray(s string) []string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return nil
	}

	s = strings.TrimPrefix(strings.TrimSuffix(s, "]"), "[")
	s = strings.TrimSpace(s)

	if s == "" {
		return nil
	}

	var result []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]

		if !inQuotes && (c == '"' || c == '\'') {
			inQuotes = true
			quoteChar = c
			continue
		}

		if inQuotes && c == quoteChar {
			inQuotes = false
			quoteChar = 0
			continue
		}

		if !inQuotes && c == ',' {
			val := strings.TrimSpace(current.String())
			if val != "" {
				result = append(result, val)
			}
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	val := strings.TrimSpace(current.String())
	if val != "" {
		result = append(result, val)
	}

	return result
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	s = unquote(s)

	if strings.HasSuffix(s, "ms") {
		ms, err := strconv.ParseInt(strings.TrimSuffix(s, "ms"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(ms) * time.Millisecond, nil
	}

	if strings.HasSuffix(s, "s") {
		sec, err := strconv.ParseInt(strings.TrimSuffix(s, "s"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(sec) * time.Second, nil
	}

	if strings.HasSuffix(s, "m") {
		min, err := strconv.ParseInt(strings.TrimSuffix(s, "m"), 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(min) * time.Minute, nil
	}

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(num), nil
}

func IsValidKeyChar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_'
}
