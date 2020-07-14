package entity

type LabelMap map[string]Label

func (m LabelMap) Get(name string) Label {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewLabel(name, 0)
}

func (m LabelMap) Pointer(name string) *Label {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewLabel(name, 0)
}

func (m LabelMap) PhotoLabel(photoId uint, labelName string, uncertainty int, source string) PhotoLabel {
	label := m.Get(labelName)

	photoLabel := NewPhotoLabel(photoId, label.ID, uncertainty, source)
	photoLabel.Label = &label

	return *photoLabel
}

var LabelFixtures = LabelMap{
	"landscape": {
		ID:               1000000,
		LabelUID:         "lt9k3pw1wowuy3c2",
		LabelSlug:        "landscape",
		CustomSlug:       "landscape",
		LabelName:        "Landscape",
		LabelPriority:    0,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       1,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"flower": {
		ID:               1000001,
		LabelUID:         "lt9k3pw1wowuy3c3",
		LabelSlug:        "flower",
		CustomSlug:       "flower",
		LabelName:        "Flower",
		LabelPriority:    1,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       2,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"cake": {
		ID:               1000002,
		LabelUID:         "lt9k3pw1wowuy3c4",
		LabelSlug:        "cake",
		CustomSlug:       "kuchen",
		LabelName:        "Cake",
		LabelPriority:    5,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       3,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"cow": {
		ID:               1000003,
		LabelUID:         "lt9k3pw1wowuy3c5",
		LabelSlug:        "cow",
		CustomSlug:       "kuh",
		LabelName:        "COW",
		LabelPriority:    -1,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       4,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"batchdelete": {
		ID:               1000004,
		LabelUID:         "lt9k3pw1wowuy3c6",
		LabelSlug:        "batchdelete",
		CustomSlug:       "batchDelete",
		LabelName:        "BatchDelete",
		LabelPriority:    1,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       5,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"updateLabel": {
		ID:               1000005,
		LabelUID:         "lt9k3pw1wowuy3c7",
		LabelSlug:        "updatelabel",
		CustomSlug:       "updateLabel",
		LabelName:        "updateLabel",
		LabelPriority:    2,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       1,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"updatePhotoLabel": {
		ID:               1000006,
		LabelUID:         "lt9k3pw1wowuy3c8",
		LabelSlug:        "updatephotolabel",
		CustomSlug:       "updateLabelPhoto",
		LabelName:        "updatePhotoLabel",
		LabelPriority:    2,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       1,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"likeLabel": {
		ID:               1000007,
		LabelUID:         "lt9k3pw1wowuy3c9",
		LabelSlug:        "likeLabel",
		CustomSlug:       "likeLabel",
		LabelName:        "likeLabel",
		LabelPriority:    3,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       1,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"no-jpeg": {
		ID:               1000008,
		LabelUID:         "lt9k3aa1wowuy310",
		LabelSlug:        "no-jpeg",
		CustomSlug:       "no-jpeg",
		LabelName:        "NO JPEG",
		LabelPriority:    -1,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       4,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"apilikeLabel": {
		ID:               1000009,
		LabelUID:         "lt9k3pw1wowuy311",
		LabelSlug:        "apilikeLabel",
		CustomSlug:       "apilikeLabel",
		LabelName:        "apilikeLabel",
		LabelPriority:    -1,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       1,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
	"apidislikeLabel": {
		ID:               1000010,
		LabelUID:         "lt9k3pw1wowuy312",
		LabelSlug:        "apidislikeLabel",
		CustomSlug:       "apidislikeLabel",
		LabelName:        "apidislikeLabel",
		LabelPriority:    -2,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		PhotoCount:       1,
		LabelCategories:  []*Label{},
		CreatedAt:        Timestamp(),
		UpdatedAt:        Timestamp(),
		DeletedAt:        nil,
		New:              false,
	},
}

// CreateLabelFixtures inserts known entities into the database for testing.
func CreateLabelFixtures() {
	for _, entity := range LabelFixtures {
		Db().Create(&entity)
	}
}
