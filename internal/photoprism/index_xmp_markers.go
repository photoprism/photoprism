package photoprism

import (
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

func xmpMarkerTarget(photoID uint, file *entity.File) *entity.File {
	if file == nil {
		return nil
	}

	if file.FilePrimary && len(file.FileUID) != 0 {
		return file
	}

	if photoID == 0 {
		return nil
	}

	primary := &entity.File{}

	if err := entity.UnscopedDb().
		Where("photo_id = ? AND file_primary = 1 AND file_error = ''", photoID).
		First(primary).Error; err != nil {
		return nil
	}

	return primary
}

func mergeXmpFaceRegion(file *entity.File, region meta.FaceRegion) (changed bool, err error) {
	if file == nil || !region.Valid() {
		return false, nil
	}

	area := crop.NewArea("face", region.Left(), region.Top(), region.W, region.H)
	marker := entity.NewMarker(*file, area, "", entity.SrcXmp, entity.MarkerFace, 100, 100)

	if marker == nil {
		return false, nil
	}

	marker.MarkerName = region.Name
	marker.SubjSrc = entity.SrcXmp
	marker.MarkerReview = false

	markers := file.Markers()

	for i := range *markers {
		existing := &(*markers)[i]

		if existing.MarkerType != entity.MarkerFace || existing.OverlapPercent(*marker) <= face.OverlapThreshold {
			continue
		}

		if region.Name == "" {
			return false, nil
		}

		if changed, err = existing.SetName(region.Name, entity.SrcXmp); err != nil {
			return false, err
		} else if changed && existing.MarkerUID != "" {
			return true, existing.Save()
		}

		return changed, nil
	}

	markers.Append(*marker)

	return true, nil
}

func importXmpFaceRegions(photoID uint, file *entity.File, regions meta.FaceRegions) (count int, err error) {
	if len(regions) == 0 {
		return 0, nil
	}

	target := xmpMarkerTarget(photoID, file)

	if target == nil {
		return 0, nil
	}

	for _, region := range regions {
		if changed, mergeErr := mergeXmpFaceRegion(target, region); mergeErr != nil {
			return count, mergeErr
		} else if changed {
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	if _, err = target.SaveMarkers(); err != nil {
		return count, err
	}

	_, err = target.UpdatePhotoFaceCount()

	return count, err
}
