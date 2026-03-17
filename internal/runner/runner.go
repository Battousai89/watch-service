package runner

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

type CommandRequest struct {
	Cmd     string
	Args    []string
	Env     map[string]string
	Timeout time.Duration
	WorkDir string
}

type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
	Duration time.Duration
}

type CommandRunner struct {
	cmdChan chan CommandRequest
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewCommandRunner(maxParallel int) *CommandRunner {
	if maxParallel <= 0 {
		maxParallel = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	runner := &CommandRunner{
		cmdChan: make(chan CommandRequest, maxParallel),
		ctx:     ctx,
		cancel:  cancel,
	}

	for i := 0; i < maxParallel; i++ {
		runner.wg.Add(1)
		go runner.worker()
	}

	return runner
}

func (r *CommandRunner) worker() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return
		case req, ok := <-r.cmdChan:
			if !ok {
				return
			}
			result := r.execute(req)

			if result.Error != nil {
				fmt.Printf("[RUNNER] Command failed: %v, exit code: %d\n", result.Error, result.ExitCode)
			} else {
				fmt.Printf("[RUNNER] Command completed in %v\n", result.Duration)
			}
		}
	}
}

func (r *CommandRunner) execute(req CommandRequest) CommandResult {
	start := time.Now()

	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, req.Cmd, req.Args...)

	if req.WorkDir != "" {
		cmd.Dir = req.WorkDir
	}

	for k, v := range req.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if len(cmd.Env) == 0 {
		cmd.Env = nil
	}

	stdout, err := cmd.Output()
	duration := time.Since(start)

	result := CommandResult{
		Duration: duration,
	}

	if stdout != nil {
		result.Stdout = string(stdout)
	}

	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Stderr = string(exitErr.Stderr)
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

func (r *CommandRunner) Run(req CommandRequest) {
	select {
	case r.cmdChan <- req:
	case <-r.ctx.Done():
	}
}

func (r *CommandRunner) Close() {
	r.cancel()
	close(r.cmdChan)
	r.wg.Wait()
}
