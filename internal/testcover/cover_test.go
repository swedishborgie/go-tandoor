package testcover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// restoreGoverdir snapshots the ambient GOCOVERDIR (go test -cover sets one
// internally since Go 1.20) and restores it after the test.
func restoreGoverdir(t *testing.T) string {
	t.Helper()
	prev := os.Getenv("GOCOVERDIR")
	// Restore the ambient value after the test; t.Setenv can't restore a
	// value captured before the test mutated the env, and leaving
	// GOCOVERDIR pointing at the suite's temp dir would misdirect the test
	// binary's own coverage flush under `go test -cover`.
	//
	//nolint:usetesting // os.Setenv in cleanup restores the pre-test value
	t.Cleanup(func() { _ = os.Setenv("GOCOVERDIR", prev) })
	return prev
}

func TestCoverBuildArgsDisabled(t *testing.T) {
	t.Setenv(OutEnv, "")
	require.False(t, Enabled())
	require.Empty(t, CoverBuildArgs())
}

func TestCoverBuildArgsEnabled(t *testing.T) {
	t.Setenv(OutEnv, "/tmp/whatever.txt")
	require.True(t, Enabled())
	require.Equal(t, []string{"-cover", "-coverpkg=./..."}, CoverBuildArgs())
}

func TestSetupDisabledIsNoop(t *testing.T) {
	t.Setenv(OutEnv, "")
	prev := restoreGoverdir(t)
	finish := Setup("unit-test", ".")
	require.Equal(t, prev, os.Getenv("GOCOVERDIR"), "Setup must not touch GOCOVERDIR when disabled")
	finish() // must not panic
}

func TestSetupEnabledWritesProfile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "coverage.txt")
	t.Setenv(OutEnv, out)
	restoreGoverdir(t)
	finish := Setup("unit-test", ".")
	require.NotEmpty(t, os.Getenv("GOCOVERDIR"))
	// No subprocess ran, so the covdata dir is empty and textfmt writes an
	// empty file — the file must still exist.
	finish()
	_, err := os.Stat(out)
	require.NoError(t, err)
}
