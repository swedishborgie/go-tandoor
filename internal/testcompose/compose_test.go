package testcompose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeStubBin creates an executable stub file at dir/name.
func makeStubBin(t *testing.T, dir, name string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
}

func TestCmdPrefersPodmanCompose(t *testing.T) {
	dir := t.TempDir()
	makeStubBin(t, dir, "podman-compose")
	t.Setenv("PATH", dir)

	cmd := Cmd("up", "-d")
	if cmd.Path != "podman-compose" && filepath.Base(cmd.Path) != "podman-compose" {
		t.Fatalf("expected podman-compose, got %q", cmd.Path)
	}
	if strings.Join(cmd.Args, " ") != "podman-compose up -d" {
		t.Fatalf("unexpected args: %v", cmd.Args)
	}
}

func TestCmdFallsBackToDockerCompose(t *testing.T) {
	dir := t.TempDir()
	makeStubBin(t, dir, "docker") // no podman-compose
	t.Setenv("PATH", dir)

	cmd := Cmd("up", "-d")
	if filepath.Base(cmd.Path) != "docker" {
		t.Fatalf("expected docker, got %q", cmd.Path)
	}
	if strings.Join(cmd.Args, " ") != "docker compose up -d" {
		t.Fatalf("unexpected args: %v", cmd.Args)
	}
}
