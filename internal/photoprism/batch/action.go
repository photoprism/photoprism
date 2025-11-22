package batch

// Action enumerates the supported batch operations such as add/remove/update.
type Action = string

const (
	ActionNone   Action = "none"
	ActionUpdate Action = "update"
	ActionAdd    Action = "add"
	ActionRemove Action = "remove"
)
