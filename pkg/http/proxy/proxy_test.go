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
	t.Run("Origin", func(t *testing.T) {
		assert.Equal(t, "", OriginScheme)
		assert.Equal(t, "", OriginHost)
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
		previousScheme := OriginScheme
		previousHost := OriginHost
		t.Cleanup(func() {
			PathPrefix = previous
			OriginScheme = previousScheme
			OriginHost = previousHost
		})

		require.NoError(t, SetPathPrefix("instance"))
		assert.Equal(t, "/instance/", PathPrefix)
		assert.Equal(t, "", OriginScheme)
		assert.Equal(t, "", OriginHost)

		require.NoError(t, SetPathPrefix("/node-a"))
		assert.Equal(t, "/node-a/", PathPrefix)

		require.NoError(t, SetPathPrefix("/foo/bar"))
		assert.Equal(t, "/foo/bar/", PathPrefix)

		require.NoError(t, SetPathPrefix(""))
		assert.Equal(t, DefaultPathPrefix, PathPrefix)
	})
	t.Run("SetPathPrefixInvalid", func(t *testing.T) {
		previous := PathPrefix
		previousScheme := OriginScheme
		previousHost := OriginHost
		t.Cleanup(func() {
			PathPrefix = previous
			OriginScheme = previousScheme
			OriginHost = previousHost
		})

		require.Error(t, SetPathPrefix("/"))
		assert.Equal(t, previous, PathPrefix)

		require.Error(t, SetPathPrefix("/instance/*"))
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
	t.Run("SetProxyURIPathOnly", func(t *testing.T) {
		previous := PathPrefix
		previousScheme := OriginScheme
		previousHost := OriginHost
		t.Cleanup(func() {
			PathPrefix = previous
			OriginScheme = previousScheme
			OriginHost = previousHost
		})

		require.NoError(t, SetProxyURI("/instance/"))
		assert.Equal(t, "/instance/", PathPrefix)
		assert.Equal(t, "", OriginScheme)
		assert.Equal(t, "", OriginHost)
	})
	t.Run("SetProxyURIAbsolute", func(t *testing.T) {
		previous := PathPrefix
		previousScheme := OriginScheme
		previousHost := OriginHost
		t.Cleanup(func() {
			PathPrefix = previous
			OriginScheme = previousScheme
			OriginHost = previousHost
		})

		require.NoError(t, SetProxyURI("https://proxy.example.com:8443/instance/"))
		assert.Equal(t, "/instance/", PathPrefix)
		assert.Equal(t, "https", OriginScheme)
		assert.Equal(t, "proxy.example.com:8443", OriginHost)
	})
	t.Run("SetProxyURIAbsoluteDefaultPath", func(t *testing.T) {
		previous := PathPrefix
		previousScheme := OriginScheme
		previousHost := OriginHost
		t.Cleanup(func() {
			PathPrefix = previous
			OriginScheme = previousScheme
			OriginHost = previousHost
		})

		require.NoError(t, SetProxyURI("https://proxy.example.com"))
		assert.Equal(t, DefaultPathPrefix, PathPrefix)
		assert.Equal(t, "https", OriginScheme)
		assert.Equal(t, "proxy.example.com", OriginHost)
	})
	t.Run("SetProxyURIInvalid", func(t *testing.T) {
		previous := PathPrefix
		previousScheme := OriginScheme
		previousHost := OriginHost
		t.Cleanup(func() {
			PathPrefix = previous
			OriginScheme = previousScheme
			OriginHost = previousHost
		})

		require.Error(t, SetProxyURI("https://proxy.example.com/?q=1"))
		assert.Equal(t, previous, PathPrefix)
		assert.Equal(t, previousScheme, OriginScheme)
		assert.Equal(t, previousHost, OriginHost)
	})
}
