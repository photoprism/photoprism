package batch

// Action enumerates the supported batch operations such as add/remove/update.
type Action = string

const (
	// ActionNone indicates that no change should be applied.
	ActionNone Action = "none"
	// ActionUpdate applies changes to existing values.
	ActionUpdate Action = "update"
	// ActionAdd adds a value to a collection (e.g., labels/albums).
	ActionAdd Action = "add"
	// ActionRemove removes a value from a collection (e.g., labels/albums).
	ActionRemove Action = "remove"
)
