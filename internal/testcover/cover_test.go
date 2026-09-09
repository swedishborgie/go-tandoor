package testcover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	finish := Setup("unit-test", ".")
	require.Empty(t, os.Getenv("GOCOVERDIR"))
	finish() // must not panic
}

func TestSetupEnabledWritesProfile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "coverage.txt")
	t.Setenv(OutEnv, out)
	t.Cleanup(func() { os.Unsetenv("GOCOVERDIR") })
	finish := Setup("unit-test", ".")
	require.NotEmpty(t, os.Getenv("GOCOVERDIR"))
	// No subprocess ran, so the covdata dir is empty and textfmt writes an
	// empty file — the file must still exist.
	finish()
	_, err := os.Stat(out)
	require.NoError(t, err)
}
