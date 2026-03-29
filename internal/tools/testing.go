package tools

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ExecResult holds the result of running a test command.
type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	TimedOut bool   `json:"timed_out"`
}

type opendevConfig struct {
	TestCommand string `yaml:"test_command"`
}

func readTestCommand(worktreeRoot string) (string, error) {
	configPath := filepath.Join(worktreeRoot, ".opendev", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("no .opendev/config.yaml found in worktree %q: %w", worktreeRoot, err)
	}

	var cfg opendevConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parsing .opendev/config.yaml: %w", err)
	}

	if cfg.TestCommand == "" {
		return "", fmt.Errorf("no test_command defined in .opendev/config.yaml")
	}

	return cfg.TestCommand, nil
}

func RunTest(worktreeRoot string, timeout time.Duration, logger *slog.Logger) (*ExecResult, error) {
	command, err := readTestCommand(worktreeRoot)
	if err != nil {
		return nil, err
	}

	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	logger.Info("test run: executing command", "worktree", worktreeRoot, "command", command, "timeout_s", timeout.Seconds())

	stdout, stderr, exitCode, timedOut, err := ExecuteTerminalCommand(worktreeRoot, command, timeout)
	if err != nil {
		return nil, fmt.Errorf("executing test command: %w", err)
	}

	return &ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		TimedOut: timedOut,
	}, nil
}
