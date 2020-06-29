package acl

type Action string
type Actions map[Action]bool

const (
	ActionDefault    Action = "*"
	ActionSearch     Action = "search"
	ActionCreate     Action = "create"
	ActionRead       Action = "read"
	ActionUpdate     Action = "update"
	ActionUpdateSelf Action = "update-self"
	ActionDelete     Action = "delete"
	ActionPrivate    Action = "private"
	ActionUpload     Action = "upload"
	ActionDownload   Action = "download"
	ActionShare      Action = "share"
	ActionLike       Action = "like"
	ActionComment    Action = "comment"
	ActionExport     Action = "export"
	ActionImport     Action = "import"
)
