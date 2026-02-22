package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
)

// videoSearchResults runs a video-only search and applies offset/count after merging related files.
func videoSearchResults(query string, count int, offset int) ([]search.Photo, error) {
	if offset < 0 {
		offset = 0
	}

	if count <= 0 {
		return []search.Photo{}, nil
	}

	frm := form.SearchPhotos{
		Query:   query,
		Primary: false,
		Merged:  true,
		Video:   true,
		Order:   sortby.Name,
	}

	target := max(count+offset, 0)

	collected := make([]search.Photo, 0, target)
	index := make(map[string]int, target)
	searchOffset := 0
	batchSize := max(count, 200)

	needComplete := false

	for len(collected) < target || needComplete {
		frm.Count = batchSize
		frm.Offset = searchOffset

		results, rawCount, err := search.Photos(frm)

		if err != nil {
			return nil, err
		}

		if len(results) == 0 || rawCount == 0 {
			break
		}

		for _, found := range results {
			key := videoSearchKey(found)
			if idx, ok := index[key]; ok {
				collected[idx].Files = videoMergeFiles(collected[idx].Files, found.Files)
				if len(collected[idx].Files) > 1 {
					collected[idx].Merged = true
				}
				continue
			}

			collected = append(collected, found)
			index[key] = len(collected) - 1
		}

		searchOffset += rawCount

		needComplete = false
		if len(collected) >= target {
			lastNeededKey := videoSearchKey(collected[target-1])
			lastBatchKey := videoSearchKey(results[len(results)-1])
			if lastNeededKey == lastBatchKey && rawCount == batchSize {
				needComplete = true
			}
		}

		if rawCount < batchSize {
			break
		}
	}

	if offset >= len(collected) {
		return []search.Photo{}, nil
	}

	end := offset + count

	if end > len(collected) {
		end = len(collected)
	}

	return collected[offset:end], nil
}

// videoSearchKey returns a stable key for de-duplicating merged photo results.
func videoSearchKey(found search.Photo) string {
	if found.ID > 0 {
		return fmt.Sprintf("id:%d", found.ID)
	}

	return found.PhotoUID
}

// videoMergeFiles appends unique files from additions into the existing list.
func videoMergeFiles(existing []entity.File, additions []entity.File) []entity.File {
	if len(additions) == 0 {
		return existing
	}

	if len(existing) == 0 {
		return additions
	}

	seen := make(map[string]struct{}, len(existing))
	for _, file := range existing {
		seen[videoFileKey(file)] = struct{}{}
	}

	for _, file := range additions {
		key := videoFileKey(file)
		if _, ok := seen[key]; ok {
			continue
		}
		existing = append(existing, file)
		seen[key] = struct{}{}
	}

	return existing
}

// videoFileKey returns a stable identifier for a file entry when merging search results.
func videoFileKey(file entity.File) string {
	if file.FileUID != "" {
		return "uid:" + file.FileUID
	}
	if file.ID != 0 {
		return fmt.Sprintf("id:%d", file.ID)
	}
	if file.FileHash != "" {
		return "hash:" + file.FileHash
	}

	return fmt.Sprintf("name:%s/%s:%d", file.FileRoot, file.FileName, file.FileSize)
}
