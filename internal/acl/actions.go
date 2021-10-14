package acl

type Action string
type Actions map[Action]bool

const (
	ActionDefault    Action = "*" // allows a subject/role to execute all other actions
	ActionSearch     Action = "search"
	ActionCreate     Action = "create"
	ActionRead       Action = "read"
	ActionUpdate     Action = "update"
	ActionUpdateSelf Action = "update-self"
	ActionDelete     Action = "delete"
	ActionArchive    Action = "archive" // includes restore
	ActionPrivate    Action = "private"
	ActionUpload     Action = "upload"
	ActionDownload   Action = "download"
	ActionShare      Action = "share"
	ActionLike       Action = "like"
	ActionComment    Action = "comment"
	ActionExport     Action = "export"
	ActionImport     Action = "import"
)
