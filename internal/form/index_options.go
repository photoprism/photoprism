package form

type IndexOptions struct {
	SkipUnchanged bool `json:"skipUnchanged"`
	CreateThumbs  bool `json:"createThumbs"`
	ConvertRaw    bool `json:"convertRaw"`
	GroomMetadata bool `json:"groomMetadata"`
}
