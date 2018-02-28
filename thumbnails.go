package photoprism

import (
	"github.com/disintegration/imaging"
	"log"
)

func CreateThumbnail() {
	src, err := imaging.Open("testdata/lena_512.png")

	if err != nil {
		log.Printf("Open failed: %s", err.Error())
	}

	// Crop the original image to 350x350px size using the center anchor.
	src = imaging.CropAnchor(src, 350, 350, imaging.Center)

	// Resize the cropped image to width = 256px preserving the aspect ratio.
	src = imaging.Resize(src, 256, 0, imaging.Lanczos)
}
