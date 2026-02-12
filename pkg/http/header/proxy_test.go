package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	t.Run("Path", func(t *testing.T) {
		assert.Equal(t, "/p/", ProxyPath)
	})
	t.Run("Methods", func(t *testing.T) {
		expected := []string{
			MethodMkcol,
			MethodCopy,
			MethodMove,
			MethodLock,
			MethodUnlock,
			MethodPropfind,
			MethodProppatch,
			MethodReport,
			MethodSearch,
			MethodMkcalendar,
			MethodACL,
			MethodBind,
			MethodUnbind,
			MethodRebind,
			MethodVersionControl,
			MethodCheckout,
			MethodUncheckout,
			MethodCheckin,
			MethodUpdate,
			MethodLabel,
			MethodMerge,
			MethodMkworkspace,
			MethodMkactivity,
			MethodBaselineControl,
			MethodOrderpatch,
		}

		assert.Equal(t, expected, ProxyMethods)
	})
}
