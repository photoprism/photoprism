package header

var (
	DefaultContentSecurityPolicy    = "frame-ancestors 'none';"
	DefaultFrameOptions             = "DENY"
	DefaultAccessControlAllowOrigin = os.Hostname()
)
