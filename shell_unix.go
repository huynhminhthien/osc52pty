//go:build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/crypto/ssh/terminal"
)

func (s *shell) startCommand() error {
	s.command = exec.Command(s.cmdLine[0], s.cmdLine[1:]...)

	ptmx, err := pty.Start(s.command)
	if err != nil {
		return fmt.Errorf("start pty failed: %v", err)
	}

	s.ptmx = ptmx
	s.commandIn = ptmx
	s.commandOut = ptmx
	s.cleanups = append(s.cleanups, func(s *shell) { s.ptmx.Close() })
	return nil
}

func (s *shell) makeTerminalRaw() error {
	stdin, ok := s.stdin.(*os.File)
	if !ok {
		return nil
	}

	oldState, err := terminal.MakeRaw(int(stdin.Fd()))
	if err != nil {
		return fmt.Errorf("make terminal raw failed: %v", err)
	}

	s.cleanups = append(s.cleanups, func(*shell) { terminal.Restore(int(stdin.Fd()), oldState) })
	return nil
}

func (s *shell) resizePTY() {
	stdin, ok := s.stdin.(*os.File)
	if !ok {
		return
	}

	signals := make(chan os.Signal, 1)
	s.cleanups = append(s.cleanups, func(*shell) { close(signals) })
	signal.Notify(signals, syscall.SIGWINCH)

	go func() {
		for range signals {
			if err := pty.InheritSize(stdin, s.ptmx); err != nil {
				log.Printf("resize pty failed: %s", err)
			}
		}
	}()

	signals <- syscall.SIGWINCH
}

func getShellName() string {
	shellName, ok := os.LookupEnv("SHELL")
	if !ok {
		shellName = "sh"
	}

	return shellName
}
