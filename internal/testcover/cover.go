// Package testcover wires subprocess coverage for e2e suites using Go's
// automatic coverage of subprocesses (Go 1.20+).
//
// When E2E_COVERAGE_OUT is set, the suite builds its target binary with
// "-cover -coverpkg=./..." and exports GOCOVERDIR; every child process then
// writes its coverage data into GOCOVERDIR on exit. The returned finish
// function converts the collected data into a text coverage profile at the
// path named by E2E_COVERAGE_OUT.
package testcover

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// OutEnv names the env var that enables subprocess coverage. Its value is
// the path of the text coverage profile written when the suite finishes.
const OutEnv = "E2E_COVERAGE_OUT"

// Enabled reports whether subprocess coverage is requested.
func Enabled() bool { return os.Getenv(OutEnv) != "" }

// CoverBuildArgs returns extra go build flags that instrument the target
// binary for subprocess coverage (empty when disabled).
func CoverBuildArgs() []string {
	if !Enabled() {
		return nil
	}
	return []string{"-cover", "-coverpkg=./..."}
}

// Setup exports GOCOVERDIR (when enabled) and returns a finish function that
// converts the collected coverage data to a text profile. name namespaces
// the temp dir per suite; workDir is the module root for the conversion
// command. finish is a no-op when coverage is disabled.
func Setup(name, workDir string) func() {
	if !Enabled() {
		return func() {}
	}
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("tandoor-%s-cov-%d", name, os.Getpid()))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "testcover: %v\n", err)
		return func() {}
	}
	os.Setenv("GOCOVERDIR", dir)
	return func() {
		out := os.Getenv(OutEnv)
		cmd := exec.CommandContext(context.Background(), "go", "tool", "covdata", "textfmt", "-i="+dir, "-o="+out)
		cmd.Dir = workDir
		if o, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "testcover: covdata textfmt: %v\n%s\n", err, o)
		} else {
			fmt.Fprintf(os.Stderr, "testcover: wrote %s\n", out)
			os.RemoveAll(dir)
		}
	}
}
