package photoprism

import (
	"testing"
	"fmt"
)

func TestMediaFile_ConvertToJpeg(t *testing.T) {
	converterCommand := "/Applications/darktable.app/Contents/MacOS/darktable-cli"

	converter := NewConverter(converterCommand)

	image1,_ := converter.ConvertToJpeg(NewMediaFile("storage/import/IMG_5901.JPG"))

	info1, _ := image1.GetExifData()

	fmt.Printf("%+v\n", info1)

	image2, _ := converter.ConvertToJpeg(NewMediaFile("storage/import/IMG_9087.CR2"))

	info2, _ := image2.GetExifData()

	fmt.Printf("%+v\n", info2)
}

func TestMediaFile_FindRelatedImages(t *testing.T) {
	image := NewMediaFile("storage/import/IMG_9079.jpg")

	related, err := image.GetRelatedFiles()

	if err != nil {
		t.Error(err)
	}

	for _, result := range related {
		info, _ := result.GetExifData()
		fmt.Printf("%s %+v\n", result.GetFilename(), info)
	}
}

func TestMediaFile_GetPerceptiveHash(t *testing.T) {
	image := NewMediaFile("storage/import/IMG_9079.jpg")

	hash, _ := image.GetPerceptiveHash()
	fmt.Printf("Perceptive Hash (large): %s\n", hash)

	image2 := NewMediaFile("storage/import/IMG_9079_small.jpg")

	hash2, _ := image2.GetPerceptiveHash()
	fmt.Printf("Perceptive Hash (small): %s\n", hash2)
}


func TestMediaFile_GetMimeType(t *testing.T) {
	image1 := NewMediaFile("storage/import/IMG_9083.jpg")

	fmt.Println("MimeType: " + image1.GetMimeType())

	image2 := NewMediaFile("storage/import/IMG_9082.CR2")

	fmt.Println("MimeType: " + image2.GetMimeType())
}