package clip

import (
	"fmt"
	"testing"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

var assetsPath = fs.Abs("../../assets")
var examplesPath = assetsPath + "/examples"

func checkEncodeText(t *testing.T, text string) face.Embedding {
	zeros := make(face.Embedding, 512)
	api := ClipApi{baseUrl: "http://clip-api:8000"}
	embedding, err := api.EncodeText(text)
	assert.NoError(t, err)
	assert.Equal(t, 512, len(embedding))
	assert.NotEqual(t, zeros, embedding)
	return embedding
}

func TestEncodeText(t *testing.T) {
	photoTextEmbedding := checkEncodeText(t, "a photo 📷")
	pictureTextEmbedding := checkEncodeText(t, "a picture")
	dogTextEmbedding := checkEncodeText(t, "a dog")

	photo2cameraSimilarity := photoTextEmbedding.CosineSimilarity(pictureTextEmbedding)
	photo2dogSimilarity := photoTextEmbedding.CosineSimilarity(dogTextEmbedding)

	fmt.Printf("photo2cameraSimilarity=%f > photo2dogSimilarity=%f ?\n", photo2cameraSimilarity, photo2dogSimilarity)
	assert.Greater(t, photo2cameraSimilarity, photo2dogSimilarity)
}

func checkEncodeImage(t *testing.T, filePath string) face.Embedding {
	zeros := make(face.Embedding, 512)
	api := ClipApi{baseUrl: "http://clip-api:8000"}
	embedding, err := api.EncodeImage(filePath)
	assert.NoError(t, err)
	assert.Equal(t, 512, len(embedding))
	assert.NotEqual(t, zeros, embedding)
	return embedding
}

func TestEncodeImage(t *testing.T) {
	beachSandEmbedding := checkEncodeImage(t, examplesPath+"/beach_sand.jpg")
	beachWoodEmbedding := checkEncodeImage(t, examplesPath+"/beach_wood.jpg")
	elephantsEmbedding := checkEncodeImage(t, examplesPath+"/elephants.jpg")
	beach2beachSimilarity := beachSandEmbedding.CosineSimilarity(beachWoodEmbedding)
	beach2elephantsSimilarity := beachSandEmbedding.CosineSimilarity(elephantsEmbedding)
	fmt.Printf("beach2beachSimilarity=%f > beach2elephantsSimilarity=%f ?\n", beach2beachSimilarity, beach2elephantsSimilarity)
	assert.Greater(t, beach2beachSimilarity, beach2elephantsSimilarity)
}

func TestCompareTextAndImage(t *testing.T) {

	doorImgEmbedding := checkEncodeImage(t, examplesPath+"/door_cyan.jpg")
	doorTextEmbedding := checkEncodeText(t, "a cyan door")
	elephantTextEmbedding := checkEncodeText(t, "a gray elephant")
	door2doorSimilarity := doorImgEmbedding.CosineSimilarity(doorTextEmbedding)
	door2elephantSimilarity := doorImgEmbedding.CosineSimilarity(elephantTextEmbedding)
	fmt.Printf("door2doorSimilarity=%f > door2elephantSimilarity=%f ?\n", door2doorSimilarity, door2elephantSimilarity)
	assert.Greater(t, door2doorSimilarity, door2elephantSimilarity)
}
