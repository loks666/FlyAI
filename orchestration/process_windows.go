//go:build windows

package orchestration

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type managedProcess struct {
	cmd     *exec.Cmd
	job     windows.Handle
	done    chan struct{}
	running atomic.Bool
	errMu   sync.Mutex
	exitErr error
}

func startManagedProcess(executable string, args []string, dir string, stdout, stderr io.Writer) (*managedProcess, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("创建进程作业失败: %w", err)
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		windows.CloseHandle(job)
		return nil, fmt.Errorf("配置进程作业失败: %w", err)
	}

	cmd := exec.Command(executable, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW | windows.CREATE_NEW_PROCESS_GROUP,
	}
	if err := cmd.Start(); err != nil {
		windows.CloseHandle(job)
		return nil, err
	}
	processHandle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = cmd.Process.Kill()
		windows.CloseHandle(job)
		return nil, fmt.Errorf("打开桥接器进程失败: %w", err)
	}
	assignErr := windows.AssignProcessToJobObject(job, processHandle)
	windows.CloseHandle(processHandle)
	if assignErr != nil {
		_ = cmd.Process.Kill()
		windows.CloseHandle(job)
		return nil, fmt.Errorf("托管桥接器进程树失败: %w", assignErr)
	}

	process := &managedProcess{cmd: cmd, job: job, done: make(chan struct{})}
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
	if p == nil {
		return
	}
	if p.job != 0 {
		_ = windows.CloseHandle(p.job)
		p.job = 0
	}
	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
		if p.cmd != nil && p.cmd.Process != nil {
			_ = p.cmd.Process.Kill()
		}
	}
}

func startDockerDesktop() error {
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Docker", "Docker", "Docker Desktop.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Docker", "Docker Desktop.exe"),
	}
	for _, candidate := range candidates {
		if !fileExists(candidate) {
			continue
		}
		cmd := exec.Command(candidate)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NEW_PROCESS_GROUP}
		return cmd.Start()
	}
	return fmt.Errorf("未找到 Docker Desktop.exe")
}
