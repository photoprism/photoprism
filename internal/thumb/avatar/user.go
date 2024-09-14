package avatar

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// SetUserImageURL sets a new user avatar URL.
func SetUserImageURL(m *entity.User, imageUrl, imageSrc, thumbPath string) error {
	if imageUrl == "" {
		return nil
	}

	u, err := url.Parse(imageUrl)
	if err != nil {
		return fmt.Errorf("invalid avatar URL (%w)", err)
	}

	var imageName string

	tmpName := filepath.Join(os.TempDir(), rnd.Base36(64))

	if err = fs.Download(tmpName, u.String()); err != nil {
		return fmt.Errorf("failed to download avatar image (%w)", err)
	}

	if mimeType, mimeErr := mimetype.DetectFile(tmpName); mimeErr != nil {
		return fmt.Errorf("failed to detect avatar type (%w)", mimeErr)
	} else {
		switch {
		case mimeType.Is(fs.MimeTypePNG):
			imageName = tmpName + fs.ExtPNG
		case mimeType.Is(fs.MimeTypeJPEG):
			imageName = tmpName + fs.ExtJPEG
		default:
			return fmt.Errorf("invalid avatar image type %s", mimeType)
		}
	}

	if err = fs.Move(tmpName, imageName); err != nil {
		return fmt.Errorf("failed to rename avatar image (%w)", err)
	}

	if err = SetUserImage(m, imageName, imageSrc, thumbPath); err != nil {
		return fmt.Errorf("failed to set avatar image (%w)", err)
	}

	if err = os.Remove(imageName); err != nil {
		return fmt.Errorf("failed to remove temporary file (%w)", err)
	}

	return nil
}

// SetUserImage sets a new user avatar image.
func SetUserImage(m *entity.User, imageName, imageSrc, thumbPath string) error {
	var conf *config.Config

	if conf = get.Config(); conf == nil {
		return fmt.Errorf("config required")
	}

	if thumbPath == "" {
		thumbPath = conf.ThumbCachePath()
	}

	if mediaFile, mediaErr := photoprism.NewMediaFile(imageName); mediaErr != nil {
		return mediaErr
	} else if err := mediaFile.GenerateThumbnails(thumbPath, false); err != nil {
		return err
	} else {
		return m.SetAvatar(mediaFile.Hash(), imageSrc)
	}
}
