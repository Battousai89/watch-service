package runner

import (
	"testing"
	"time"
)

func TestCommandRunner_Run(t *testing.T) {
	r := NewCommandRunner(1)
	defer r.Close()

	r.Run(CommandRequest{
		Cmd:  "cmd",
		Args: []string{"/C", "echo", "test"},
	})

	time.Sleep(200 * time.Millisecond)
}

func TestCommandRunner_Multiple(t *testing.T) {
	r := NewCommandRunner(5)
	defer r.Close()

	for i := 0; i < 3; i++ {
		r.Run(CommandRequest{
			Cmd:  "cmd",
			Args: []string{"/C", "echo", "test"},
		})
	}

	time.Sleep(500 * time.Millisecond)
}

func TestCommandRunner_Timeout(t *testing.T) {
	r := NewCommandRunner(1)
	defer r.Close()

	start := time.Now()
	r.Run(CommandRequest{
		Cmd:     "cmd",
		Args:    []string{"/C", "timeout /t 5"},
		Timeout: 100 * time.Millisecond,
	})

	time.Sleep(500 * time.Millisecond)

	if time.Since(start) > 1*time.Second {
		t.Error("timeout did not work")
	}
}

func BenchmarkCommandRunner(b *testing.B) {
	r := NewCommandRunner(10)
	defer r.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Run(CommandRequest{
			Cmd:  "cmd",
			Args: []string{"/C", "echo", "test"},
		})
	}

	time.Sleep(100 * time.Millisecond)
}
