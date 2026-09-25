package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Markers represents a list of markers.
type Markers []Marker

// Save stores the markers in the database.
func (m Markers) Save(file *File) (count int, err error) {
	if file == nil {
		return 0, fmt.Errorf("file required for saving markers")
	}

	for i := range m {
		if m[i].UpdateFile(file) {
			continue
		}

		if created, err := CreateMarkerIfNotExists(&m[i]); err != nil {
			log.Errorf("markers: %s (save)", err)
		} else {
			m[i] = *created
		}
	}

	return file.UpdatePhotoFaceCount()
}

// Unsaved tests if any marker hasn't been saved yet.
func (m Markers) Unsaved() bool {
	for i := range m {
		if m[i].Unsaved() {
			return true
		}
	}

	return false
}

// Contains returns true if a marker at the same position already exists.
func (m Markers) Contains(other Marker) bool {
	for i := range m {
		if m[i].OverlapPercent(other) > face.OverlapThreshold {
			return true
		}
	}

	return false
}

// Overlapping returns a pointer to the first non-rejected marker that overlaps
// other above the face-overlap threshold, or nil when none does. Used to
// reconcile an imported XMP region onto an existing marker instead of
// discarding the overlap the way Contains does.
func (m Markers) Overlapping(other Marker) *Marker {
	for i := range m {
		if m[i].MarkerInvalid {
			continue
		}
		if m[i].OverlapPercent(other) > face.OverlapThreshold {
			return &m[i]
		}
	}

	return nil
}

// OverlapsInvalid reports whether any rejected marker (MarkerInvalid) overlaps
// other above the threshold. An XMP region that lands on a rejected marker must
// be skipped so it is not resurrected as a fresh marker at a different thumb hash.
func (m Markers) OverlapsInvalid(other Marker) bool {
	for i := range m {
		if m[i].MarkerInvalid && m[i].OverlapPercent(other) > face.OverlapThreshold {
			return true
		}
	}

	return false
}

// DetectedFaceCount returns the number of automatically detected face markers.
func (m Markers) DetectedFaceCount() (count int) {
	for i := range m {
		if m[i].DetectedFace() {
			count++
		}
	}

	return count
}

// ValidFaceCount returns the number of valid face markers.
func (m Markers) ValidFaceCount() (count int) {
	for i := range m {
		if m[i].ValidFace() {
			count++
		}
	}

	return count
}

// SubjectUIDs returns the distinct uids of the people the face markers point at.
func (m Markers) SubjectUIDs() (uids []string) {
	return m.distinct(func(i int) string { return m[i].SubjUID })
}

// MarkerNames returns the distinct names the face markers carry, which need not be linked to a
// subject.
func (m Markers) MarkerNames() (names []string) {
	return m.distinct(func(i int) string { return m[i].MarkerName })
}

// distinct collects the non-empty values of a face-marker field without repeats, so a heavily
// marked file does not send one bind argument per marker.
func (m Markers) distinct(value func(i int) string) (result []string) {
	seen := make(map[string]struct{}, len(m))

	for i := range m {
		if m[i].MarkerType != MarkerFace {
			continue
		}

		if v := value(i); v != "" {
			if _, dup := seen[v]; !dup {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}

	return result
}

// SubjectNames returns known subject names, leaving out people whose name is withheld. Generated
// titles, captions and keywords derive from it and are stored for every session to read, so a
// failure to resolve the flags returns no names rather than risking one: the caller writes that,
// and only a later maintenance pass restores the names it dropped.
func (m Markers) SubjectNames() (names []string) {
	// Nothing to classify, and nothing SubjectName could return either.
	if len(m.SubjectUIDs()) == 0 && len(m.MarkerNames()) == 0 {
		return nil
	}

	withheld, err := FindWithheldPeople()

	if err != nil {
		log.Warnf("markers: %s while resolving people visibility, omitting all names from generated metadata", err)
		return nil
	}

	for i := range m {
		if m[i].MarkerInvalid || m[i].MarkerType != MarkerFace {
			continue
		}

		// Checked on the name as well as the link, since SubjectName prefers the name column and
		// a marker may carry one before it is linked.
		if withheld.Withholds(m[i].SubjUID, m[i].MarkerName) {
			continue
		}

		if n := m[i].SubjectName(); n != "" {
			names = append(names, n)
		}
	}

	return txt.UniqueNames(names)
}

// Labels returns a list of matching labels.
// TODO: Function is currently unused, decide how to proceed with it.
func (m Markers) Labels() (result classify.Labels) {
	faceCount := 0

	labelSrc := SrcImage
	labelUncertainty := 100

	for i := range m {
		if m[i].ValidFace() {
			faceCount++

			if u := m[i].Uncertainty(); u < labelUncertainty {
				labelUncertainty = u
			}

			if m[i].MarkerSrc != "" {
				labelSrc = m[i].MarkerSrc
			}
		}
	}

	if faceCount < 1 {
		return classify.Labels{}
	}

	var rule classify.LabelRule

	if faceCount == 1 {
		rule = classify.Rules["portrait"]
	} else {
		rule = classify.Rules["people"]
	}

	return classify.Labels{classify.Label{
		Name:        rule.Label,
		Source:      labelSrc,
		Uncertainty: labelUncertainty,
		Priority:    rule.Priority,
		Categories:  rule.Categories,
	}}
}

// Append adds a marker.
func (m *Markers) Append(marker Marker) {
	*m = append(*m, marker)
}

// AppendWithEmbedding adds a marker with face embedding.
func (m *Markers) AppendWithEmbedding(marker Marker) {
	if !marker.Embeddings().One() {
		// Ignore markers that don't have exactly one embedding.
		return
	}

	m.Append(marker)
}

// FindMarkers returns up to 1000 markers for a given file uid. Indexing and face clustering read
// through it, so it is never scoped to a session.
func FindMarkers(fileUid string) (Markers, error) {
	return FindVisibleMarkers(fileUid, false)
}

// FindVisibleMarkers returns up to 1000 markers for a given file uid, excluding those that point
// at a person whose name is withheld when omitWithheld is set.
func FindVisibleMarkers(fileUid string, omitWithheld bool) (Markers, error) {
	m := Markers{}

	markerTable := Marker{}.TableName()

	stmt := Db().
		Table(markerTable).
		Where(fmt.Sprintf("%s.file_uid = ?", markerTable), fileUid).
		Order(markerTable + ".x").
		Offset(0).Limit(1000)

	if omitWithheld {
		joins, cond := VisiblePeopleFilter(markerTable, true)

		// Selected explicitly, or the joined subject columns scan into the marker fields.
		stmt = stmt.Select(markerTable + ".*")

		for _, join := range joins {
			stmt = stmt.Joins(join)
		}

		stmt = stmt.Where(cond)
	}

	err := stmt.Find(&m).Error

	return m, err
}
