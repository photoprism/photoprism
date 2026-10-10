package classify

// priorityIgnore excludes labels from classification results.
const priorityIgnore = -3

// Data sources.
const (
	SrcAuto     = ""
	SrcManual   = "manual"
	SrcLocation = "location"
	SrcImage    = "image"
	SrcTitle    = "title"
	SrcCaption  = "caption"
	SrcSubject  = "subject"
	SrcKeyword  = "keyword"
)
