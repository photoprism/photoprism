package entity

import (
	"strings"

	"github.com/gosimple/slug"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// folderAlbumSlugCandidates returns the current and legacy slug candidates for a folder path.
func folderAlbumSlugCandidates(albumPath string) []string {
	albumPath = clean.SlashPath(albumPath)

	if albumPath == "" {
		return nil
	}

	candidates := make([]string, 0, 2)

	for _, value := range []string{txt.Slug(albumPath), legacyFolderAlbumSlug(albumPath)} {
		if value == "" {
			continue
		}

		duplicate := false

		for _, existing := range candidates {
			if existing == value {
				duplicate = true
				break
			}
		}

		if !duplicate {
			candidates = append(candidates, value)
		}
	}

	return candidates
}

// legacyFolderAlbumSlug reproduces the folder album slug logic used before the 2026 collision fixes.
func legacyFolderAlbumSlug(albumPath string) string {
	albumPath = strings.TrimSpace(clean.SlashPath(albumPath))

	if albumPath == "" || albumPath == "-" {
		return albumPath
	}

	if albumPath[0] == txt.SlugEncoded && txt.ContainsAlnumLower(albumPath[1:]) {
		return txt.Clip(albumPath, txt.ClipSlug)
	}

	result := slug.Make(albumPath)

	if result == "" {
		result = string(txt.SlugEncoded) + txt.SlugEncoding.EncodeToString([]byte(albumPath))
	}

	return txt.Clip(result, txt.ClipSlug)
}
