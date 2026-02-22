package video

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// openTestFile opens a test fixture file and registers a cleanup to close it.
func openTestFile(t *testing.T, fileName string) *os.File {
	t.Helper()

	f, err := os.Open(fileName) //nolint:gosec // test fixture path
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	return f
}
