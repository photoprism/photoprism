package clip

type Clip struct {
	Api      *ClipApi
	Db       *EmbeddingDB
	disabled bool
}

// New returns new Clip instance.
func New(collection string, vectorSize uint, disabled bool) *Clip {
	api := ClipApi{baseUrl: "http://clip-api:8000"}
	db := EmbeddingDB{
		Url:        "http://embedding-db:6333",
		Collection: collection,
		VectorSize: vectorSize,
	}
	return &Clip{&api, &db, disabled}
}

// EncodeImageAndSave get clip embeddings from image filename and saves matching embedding to db.
func (m *Clip) EncodeImageAndSave(filename string, photoId uint) error {
	embedding, err := m.Api.EncodeImage(filename)
	if err != nil {
		return err
	} else if err := m.Db.SaveEmbedding(embedding, photoId); err != nil {
		return err
	}
	return nil
}
