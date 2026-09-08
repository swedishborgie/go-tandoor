// Package e2e contains CLI end-to-end tests.
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	composeProject = "tandoor-test-cli"
	composeDir     string
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	composeDir = filepath.Dir(file)
}

func composeCmd(args ...string) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), "podman-compose", args...)
	cmd.Dir = composeDir
	cmd.Env = os.Environ()
	return cmd
}

// Start starts the test compose stack.
func Start() error {
	_ = Stop()
	cmd := composeCmd("-f", "docker-compose.test.yml", "-p", composeProject, "up", "-d")
	out, err := runCmd(cmd)
	fmt.Printf("START out: %s, err: %v\n", out, err)
	if err != nil {
		return fmt.Errorf("podman-compose up failed: %w\n%s", err, out)
	}
	return nil
}

// Stop stops the test compose stack.
func Stop() error {
	cmd := composeCmd("-f", "docker-compose.test.yml", "-p", composeProject, "down", "-v")
	out, err := runCmd(cmd)
	if err != nil {
		return fmt.Errorf("podman-compose down failed: %w\n%s", err, out)
	}
	return nil
}

func runCmd(cmd *exec.Cmd) (string, error) {
	outBytes, err := cmd.CombinedOutput()
	return string(outBytes), err
}
