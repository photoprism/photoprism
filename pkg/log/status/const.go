package status

// Generic outcomes for use in system and audit logs.
const (
	Failed              = "failed"
	Denied              = "denied"
	Granted             = "granted"
	Added               = "added"
	Updated             = "updated"
	Created             = "created"
	Deleted             = "deleted"
	Succeeded           = "succeeded"
	Verified            = "verified"
	Activated           = "activated"
	Deactivated         = "deactivated"
	Joined              = "joined"
	Confirmed           = "confirmed"
	Skipped             = "skipped"
	NotFound            = "not found"
	Unsupported         = "not supported"
	RateLimited         = "rate limit exceeded"
	InsufficientStorage = "insufficient storage"
)
