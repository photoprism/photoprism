package header

import (
	"net/http"
)

// Canonical HTTP and WebDAV method names.
const (
	MethodHead            = http.MethodHead
	MethodGet             = http.MethodGet
	MethodPut             = http.MethodPut
	MethodPost            = http.MethodPost
	MethodPatch           = http.MethodPatch
	MethodDelete          = http.MethodDelete
	MethodOptions         = http.MethodOptions
	MethodMkcol           = "MKCOL"
	MethodCopy            = "COPY"
	MethodMove            = "MOVE"
	MethodLock            = "LOCK"
	MethodUnlock          = "UNLOCK"
	MethodPropfind        = "PROPFIND"
	MethodProppatch       = "PROPPATCH"
	MethodReport          = "REPORT"
	MethodSearch          = "SEARCH"
	MethodMkcalendar      = "MKCALENDAR"
	MethodACL             = "ACL"
	MethodBind            = "BIND"
	MethodUnbind          = "UNBIND"
	MethodRebind          = "REBIND"
	MethodVersionControl  = "VERSION-CONTROL"
	MethodCheckout        = "CHECKOUT"
	MethodUncheckout      = "UNCHECKOUT"
	MethodCheckin         = "CHECKIN"
	MethodUpdate          = "UPDATE"
	MethodLabel           = "LABEL"
	MethodMerge           = "MERGE"
	MethodMkworkspace     = "MKWORKSPACE"
	MethodMkactivity      = "MKACTIVITY"
	MethodBaselineControl = "BASELINE-CONTROL"
	MethodOrderpatch      = "ORDERPATCH"
)
