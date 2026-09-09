// Package testcompose provides compose command construction that works with
// either podman-compose (preferred, e.g. local dev machines) or the docker
// compose v2 plugin (e.g. GitHub Actions runners, which ship docker but not
// podman).
package testcompose

import "os/exec"

// Cmd returns an exec.Cmd running the given compose arguments. The
// podman-compose single-binary form is used when available; otherwise the
// command falls back to "docker compose <args>".
func Cmd(args ...string) *exec.Cmd {
	if _, err := exec.LookPath("podman-compose"); err == nil {
		return exec.Command("podman-compose", args...)
	}
	return exec.Command("docker", append([]string{"compose"}, args...)...)
}
