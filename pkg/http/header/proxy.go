package header

const (
	// ProxyPath defines the shared-domain proxy route prefix used by the Portal.
	// Keep all code in sync by referencing this constant instead of hard-coded "/p/" strings.
	ProxyPath = "/p/"
)

var (
	// ProxyMethods lists additional methods that proxy routes should support
	// beyond standard methods provided by gin.Engine.Any.
	ProxyMethods = []string{
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
)
