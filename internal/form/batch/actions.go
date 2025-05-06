package batch

type Action = string

const (
	ActionNone   Action = "none"
	ActionUpdate Action = "update"
	ActionAdd    Action = "add"
	ActionRemove Action = "remove"
)
