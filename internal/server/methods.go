package server

// Canonical HTTP and WebDAV method names reused across handlers.
const (
	MethodHead      = "HEAD"
	MethodGet       = "GET"
	MethodPut       = "PUT"
	MethodPost      = "POST"
	MethodPatch     = "PATCH"
	MethodDelete    = "DELETE"
	MethodOptions   = "OPTIONS"
	MethodMkcol     = "MKCOL"
	MethodCopy      = "COPY"
	MethodMove      = "MOVE"
	MethodLock      = "LOCK"
	MethodUnlock    = "UNLOCK"
	MethodPropfind  = "PROPFIND"
	MethodProppatch = "PROPPATCH"
)
