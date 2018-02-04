package photoprism

import (
	"testing"
	"fmt"
)

func TestImage_GetExifData(t *testing.T) {
	config := NewTestConfig()

	converter := NewConverter(config.DarktableCli)

	image1 := NewMediaFile("storage/import/IMG_9083.jpg")

	info1, _ := image1.GetExifData()

	fmt.Printf("%+v\n", info1)

	image2,_ := converter.ConvertToJpeg(NewMediaFile("storage/import/IMG_5901.JPG"))

	info2, _ := image2.GetExifData()

	fmt.Printf("%+v\n", info2)

	image3, _ := converter.ConvertToJpeg(NewMediaFile("storage/import/IMG_9087.CR2"))

	info3, _ := image3.GetExifData()

	fmt.Printf("%+v\n", info3)
}

