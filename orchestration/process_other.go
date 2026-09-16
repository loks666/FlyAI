//go:build !windows

package orchestration

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

type managedProcess struct {
	cmd     *exec.Cmd
	done    chan struct{}
	running atomic.Bool
	errMu   sync.Mutex
	exitErr error
}

func startManagedProcess(executable string, args []string, dir string, stdout, stderr io.Writer) (*managedProcess, error) {
	cmd := exec.Command(executable, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	process := &managedProcess{cmd: cmd, done: make(chan struct{})}
	process.running.Store(true)
	go func() {
		err := cmd.Wait()
		process.errMu.Lock()
		process.exitErr = err
		process.errMu.Unlock()
		process.running.Store(false)
		close(process.done)
	}()
	return process, nil
}

func (p *managedProcess) isRunning() bool { return p != nil && p.running.Load() }

func (p *managedProcess) pid() int {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p *managedProcess) err() error {
	if p == nil {
		return nil
	}
	p.errMu.Lock()
	defer p.errMu.Unlock()
	return p.exitErr
}

func (p *managedProcess) stop() {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return
	}
	_ = p.cmd.Process.Kill()
	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
	}
}

func startDockerDesktop() error { return fmt.Errorf("仅支持 Windows") }
