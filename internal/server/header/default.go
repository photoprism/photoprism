package header

var (
	DefaultAccessControlAllowOrigin = os.Hostname()
	DefaultContentSecurityPolicy    = "frame-ancestors 'none';"
	DefaultFrameOptions             = "DENY"
)
