package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type CommandSpec struct {
	Shell     string
	ShellArgs []string
	Command   string
	WorkDir   string
}

type Result struct {
	Shell      string    `json:"shell"`
	ShellArgs  []string  `json:"shellArgs"`
	Command    string    `json:"command"`
	WorkDir    string    `json:"workDir,omitempty"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	DurationMs int64     `json:"durationMs"`
	ExitCode   int       `json:"exitCode"`
	Success    bool      `json:"success"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	Error      string    `json:"error,omitempty"`
}

func DefaultShell() (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{"-Command"}
	}
	return "/bin/sh", []string{"-c"}
}

func RunShellCommand(spec CommandSpec) (Result, error) {
	start := time.Now()
	result := Result{
		Shell:     spec.Shell,
		ShellArgs: append([]string{}, spec.ShellArgs...),
		Command:   spec.Command,
		WorkDir:   spec.WorkDir,
		StartTime: start,
	}

	args := append(append([]string{}, spec.ShellArgs...), spec.Command)
	cmd := exec.Command(spec.Shell, args...)
	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		result.EndTime = time.Now()
		result.Error = err.Error()
		result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
		result.ExitCode = 1
		return result, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		result.EndTime = time.Now()
		result.Error = err.Error()
		result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
		result.ExitCode = 1
		return result, err
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	if err := cmd.Start(); err != nil {
		result.EndTime = time.Now()
		result.Error = err.Error()
		result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
		result.ExitCode = 1
		return result, err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(io.MultiWriter(os.Stdout, &stdoutBuf), stdoutPipe)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(io.MultiWriter(os.Stderr, &stderrBuf), stderrPipe)
	}()

	waitErr := cmd.Wait()
	wg.Wait()

	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = waitErr.Error()
		result.Success = false
		return result, waitErr
	}

	result.ExitCode = 0
	result.Success = true
	return result, nil
}
