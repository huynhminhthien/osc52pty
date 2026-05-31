//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func (s *shell) startCommand() error {
	s.command = exec.Command(s.cmdLine[0], s.cmdLine[1:]...)

	stdin, err := s.command.StdinPipe()
	if err != nil {
		return fmt.Errorf("create command stdin failed: %v", err)
	}

	outputReader, outputWriter := io.Pipe()
	s.command.Stdout = outputWriter
	s.command.Stderr = outputWriter

	if err := s.command.Start(); err != nil {
		stdin.Close()
		outputReader.Close()
		outputWriter.Close()
		return fmt.Errorf("start command failed: %v", err)
	}

	s.commandIn = stdin
	s.commandOut = outputReader
	s.outputClose = outputWriter
	s.cleanups = append(s.cleanups, func(*shell) {
		stdin.Close()
		outputReader.Close()
		outputWriter.Close()
	})
	return nil
}

func (s *shell) makeTerminalRaw() error {
	return nil
}

func (s *shell) resizePTY() {
}

func getShellName() string {
	shellName, ok := os.LookupEnv("COMSPEC")
	if !ok || shellName == "" {
		shellName = "cmd.exe"
	}

	return shellName
}
