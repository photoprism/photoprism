package viewer

import (
	"fmt"
	"math"

	"github.com/photoprism/photoprism/internal/thumb"
)

// DownloadUrl returns a download url based on hash, api uri, and download token.
func DownloadUrl(h, apiUri, downloadToken string) string {
	return fmt.Sprintf("%s/dl/%s?t=%s", apiUri, h, downloadToken)
}

// ThumbUrl returns a thumbnail url based on hash, thumb name, cdn uri, and preview token.
func ThumbUrl(h, name, contentUri, previewToken string) string {
	return fmt.Sprintf("%s/t/%s/%s/%s", contentUri, h, previewToken, name)
}

// Thumb represents a photo viewer thumbnail.
type Thumb struct {
	W   int    `json:"w"`
	H   int    `json:"h"`
	Src string `json:"src"`
}

// NewThumb creates a new photo viewer thumb.
func NewThumb(w, h int, hash string, s thumb.Size, contentUri, previewToken string) Thumb {
	if s.Width >= w && s.Height >= h {
		// Smaller
		return Thumb{W: w, H: h, Src: ThumbUrl(hash, s.Name.String(), contentUri, previewToken)}
	}

	srcAspectRatio := float64(w) / float64(h)
	maxAspectRatio := float64(s.Width) / float64(s.Height)

	var newW, newH int

	if srcAspectRatio > maxAspectRatio {
		newW = s.Width
		newH = int(math.Round(float64(newW) / srcAspectRatio))
	} else {
		newH = s.Height
		newW = int(math.Round(float64(newH) * srcAspectRatio))
	}

	return Thumb{W: newW, H: newH, Src: ThumbUrl(hash, s.Name.String(), contentUri, previewToken)}
}
