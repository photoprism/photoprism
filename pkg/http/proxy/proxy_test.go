package proxy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestProxy(t *testing.T) {
	t.Run("Path", func(t *testing.T) {
		assert.Equal(t, DefaultPathPrefix, PathPrefix)
	})
	t.Run("Methods", func(t *testing.T) {
		expected := []string{
			header.MethodMkcol,
			header.MethodCopy,
			header.MethodMove,
			header.MethodLock,
			header.MethodUnlock,
			header.MethodPropfind,
			header.MethodProppatch,
			header.MethodReport,
			header.MethodSearch,
			header.MethodMkcalendar,
			header.MethodACL,
			header.MethodBind,
			header.MethodUnbind,
			header.MethodRebind,
			header.MethodVersionControl,
			header.MethodCheckout,
			header.MethodUncheckout,
			header.MethodCheckin,
			header.MethodUpdate,
			header.MethodLabel,
			header.MethodMerge,
			header.MethodMkworkspace,
			header.MethodMkactivity,
			header.MethodBaselineControl,
			header.MethodOrderpatch,
		}

		assert.Equal(t, expected, Methods)
	})
	t.Run("Settings", func(t *testing.T) {
		assert.Equal(t, 60*time.Second, Timeout)
		assert.Equal(t, 60*time.Second, CacheTTL)
		assert.Equal(t, 2*time.Second, CacheNegativeTTL)
		assert.Equal(t, 1*time.Minute, CacheCleanup)
	})
	t.Run("SetPathPrefix", func(t *testing.T) {
		previous := PathPrefix
		t.Cleanup(func() {
			PathPrefix = previous
		})

		require.NoError(t, SetPathPrefix("tenant"))
		assert.Equal(t, "/tenant/", PathPrefix)

		require.NoError(t, SetPathPrefix("/node-a"))
		assert.Equal(t, "/node-a/", PathPrefix)

		require.NoError(t, SetPathPrefix("/foo/bar"))
		assert.Equal(t, "/foo/bar/", PathPrefix)

		require.NoError(t, SetPathPrefix(""))
		assert.Equal(t, DefaultPathPrefix, PathPrefix)
	})
	t.Run("SetPathPrefixInvalid", func(t *testing.T) {
		previous := PathPrefix
		t.Cleanup(func() {
			PathPrefix = previous
		})

		require.Error(t, SetPathPrefix("/"))
		assert.Equal(t, previous, PathPrefix)

		require.Error(t, SetPathPrefix("/tenant/*"))
		assert.Equal(t, previous, PathPrefix)

		require.Error(t, SetPathPrefix("/foo//bar"))
		assert.Equal(t, previous, PathPrefix)

		require.Error(t, SetPathPrefix("/foo/./bar"))
		assert.Equal(t, previous, PathPrefix)

		require.Error(t, SetPathPrefix("/foo/../bar"))
		assert.Equal(t, previous, PathPrefix)

		require.Error(t, SetPathPrefix(`/foo\bar`))
		assert.Equal(t, previous, PathPrefix)
	})
}
