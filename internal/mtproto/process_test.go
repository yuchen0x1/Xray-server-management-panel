package mtproto

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

func TestProcessIsRunningUsesExitSignal(t *testing.T) {
	done := make(chan struct{})
	p := &Process{
		cmd:  &exec.Cmd{Process: &os.Process{}, ProcessState: &os.ProcessState{}},
		done: done,
	}
	if !p.IsRunning() {
		t.Fatal("process must remain running until wait completes")
	}
	close(done)
	if p.IsRunning() {
		t.Fatal("process must stop running after wait completes")
	}
}

func TestProcessLifecycleReadsConcurrentExit(t *testing.T) {
	if os.Getenv("XUI_MTPROTO_EXIT_HELPER") == "1" {
		os.Exit(2)
	}
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestProcessLifecycleReadsConcurrentExit$")
	cmd.Env = append(os.Environ(), "XUI_MTPROTO_EXIT_HELPER=1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	done := make(chan struct{})
	p := &Process{cmd: cmd, done: done, logWriter: &procLogWriter{}}
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for {
				_ = p.IsRunning()
				_ = p.GetResult()
				select {
				case <-done:
					return
				default:
				}
			}
		})
	}
	p.wait(cmd, done)
	readers.Wait()
	if p.IsRunning() {
		t.Fatal("exited helper still reports running")
	}
	if got := p.GetResult(); !strings.Contains(got, "exit status 2") {
		t.Fatalf("GetResult() = %q, want exit status 2", got)
	}
}

func TestProcessGetResultPrefersLog(t *testing.T) {
	p := newProcess("", "test")
	p.exitErr = errors.New("exit error")
	if got := p.GetResult(); got != "exit error" {
		t.Fatalf("GetResult() = %q, want exit error", got)
	}
	p.logWriter.lastLine = "last child message"
	if got := p.GetResult(); got != "last child message" {
		t.Fatalf("GetResult() = %q, want last child message", got)
	}
}
