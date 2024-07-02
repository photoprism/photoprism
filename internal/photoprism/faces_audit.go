package photoprism

import (
	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/tensorflow/face"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Audit face clusters and subjects.
func (w *Faces) Audit(fix bool) (err error) {
	invalidFaces, invalidSubj, err := query.MarkersWithNonExistentReferences()

	if err != nil {
		return err
	}

	subj, err := query.SubjectMap()

	if err != nil {
		log.Errorf("faces: %s (find subjects)", err)
	}

	if n := len(subj); n == 0 {
		log.Infof("faces: found no subjects")
	} else {
		log.Infof("faces: found %s", english.Plural(n, "subject", "subjects"))
	}

	// Fix non-existent marker subjects references?
	if n := len(invalidSubj); n == 0 {
		log.Infof("faces: found no invalid marker subjects")
	} else if !fix {
		log.Infof("faces: %s with non-existent subjects", english.Plural(n, "marker", "markers"))
	} else if removed, err := query.RemoveNonExistentMarkerSubjects(); err != nil {
		log.Errorf("faces: %s (remove orphan subjects)", err)
	} else if removed > 0 {
		log.Infof("faces: removed %d / %d markers with non-existent subjects", removed, n)
	}

	// Fix non-existent marker face references?
	if n := len(invalidFaces); n == 0 {
		log.Infof("faces: found no invalid marker faces")
	} else if !fix {
		log.Infof("faces: %s with non-existent faces", english.Plural(n, "marker", "markers"))
	} else if removed, err := query.RemoveNonExistentMarkerFaces(); err != nil {
		log.Errorf("faces: %s (remove orphan embeddings)", err)
	} else if removed > 0 {
		log.Infof("faces: removed %d / %d markers with non-existent faces", removed, n)
	}

	conflicts := 0
	resolved := 0

	faces, ids, err := query.FacesByID(true, false, false, false)

	if err != nil {
		return err
	}

	// Remembers matched combinations.
	done := make(map[string]bool, len(ids)*len(ids))

	// Find face assignment collisions.
	for _, i := range ids {
		for _, j := range ids {
			var f1, f2 entity.Face

			if f, ok := faces[i]; ok {
				f1 = f
			} else {
				continue
			}

			if f, ok := faces[j]; ok {
				f2 = f
			} else {
				continue
			}

			var matchId string

			// Skip?
			if matchId = f1.MatchId(f2); matchId == "" || done[matchId] {
				continue
			}

			// Compare face 1 with face 2.
			if matched, dist := f1.Match(face.Embeddings{f2.Embedding()}); matched {
				if f1.SubjUID == f2.SubjUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.MatchDist

				log.Infof("faces: face %s has ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjUID != "" {
					log.Infof("faces: face %s belongs to subject %s (%s %s)", f1.ID, entity.SubjNames.Log(f1.SubjUID), f1.SubjUID, entity.SrcString(f1.FaceSrc))
				} else {
					log.Infof("faces: face %s has no subject assigned (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjUID != "" {
					log.Infof("faces: face %s belongs to subject %s (%s %s)", f2.ID, entity.SubjNames.Log(f2.SubjUID), f2.SubjUID, entity.SrcString(f2.FaceSrc))
				} else {
					log.Infof("faces: face %s has no subject assigned (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				// Skip fix?
				if !fix {
					continue
				}

				// Resolve.
				success, failed := f1.ResolveCollision(face.Embeddings{f2.Embedding()})

				// Failed?
				if failed != nil {
					log.Errorf("faces: conflict resolution for %s failed, face %s has collisions with other persons (%s)", entity.SubjNames.Log(f1.SubjUID), f1.ID, failed)
					continue
				}

				// Success?
				if success {
					log.Infof("faces: successful conflict resolution for %s, face %s had collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
					resolved++
					faces, _, err = query.FacesByID(true, false, false, false)
					logErr("faces", "refresh", err)
				} else {
					log.Infof("faces: conflict resolution for %s not successful, face %s still has collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
				}

				done[matchId] = true
			}
		}
	}

	// Show conflict resolution results.
	if conflicts == 0 {
		log.Infof("faces: found no ambiguous subjects")
	} else if !fix {
		log.Infof("faces: found %s", english.Plural(conflicts, "ambiguous subject", "ambiguous subjects"))
	} else {
		log.Infof("faces: found %s, %d resolved", english.Plural(conflicts, "ambiguous subject", "ambiguous subjects"), resolved)
	}

	// Show remaining issues.
	if markers, err := query.MarkersWithSubjectConflict(); err != nil {
		logErr("faces", "find marker conflicts", err)
	} else {
		for _, m := range markers {
			if m.FaceID == "" {
				log.Warnf("faces: marker %s has an empty face id - you may have found a bug", m.MarkerUID)
			} else if f, ok := faces[m.FaceID]; !ok {
				log.Warnf("faces: marker %s has invalid face %s of subject %s (%s)", m.MarkerUID, m.FaceID, entity.SubjNames.Log(f.SubjUID), f.SubjUID)
			} else if m.SubjUID != "" {
				log.Infof("faces: marker %s with %s subject %s (%s) conflicts with face %s (%s) of subject %s (%s)", m.MarkerUID, entity.SrcString(m.SubjSrc), entity.SubjNames.Log(m.SubjUID), m.SubjUID, m.FaceID, entity.SrcString(f.FaceSrc), entity.SubjNames.Log(f.SubjUID), f.SubjUID)
			} else if m.MarkerName != "" {
				log.Infof("faces: marker %s with %s subject name %s conflicts with face %s (%s) of subject %s (%s)", m.MarkerUID, entity.SrcString(m.SubjSrc), clean.Log(m.MarkerName), m.FaceID, entity.SrcString(f.FaceSrc), entity.SubjNames.Log(f.SubjUID), f.SubjUID)
			} else {
				log.Infof("faces: marker %s with unknown subject (%s) conflicts with face %s (%s) of subject %s (%s)", m.MarkerUID, entity.SrcString(m.SubjSrc), m.FaceID, entity.SrcString(f.FaceSrc), entity.SubjNames.Log(f.SubjUID), f.SubjUID)
			}

		}
	}

	// Find and fix orphan face clusters.
	if orphans, err := entity.OrphanFaces(); err != nil {
		log.Errorf("faces: %s while finding orphan face clusters", err)
	} else if l := len(orphans); l == 0 {
		log.Infof("faces: found no orphan face clusters")
	} else if !fix {
		log.Infof("faces: found %s", english.Plural(l, "orphan face cluster", "orphan face clusters"))
	} else if err := orphans.Delete(); err != nil {
		log.Errorf("faces: failed removing %s: %s", english.Plural(l, "orphan face cluster", "orphan face clusters"), err)
	} else {
		log.Infof("faces: removed %s", english.Plural(l, "orphan face cluster", "orphan face clusters"))
	}

	// Find and fix orphan people.
	if orphans, err := entity.OrphanPeople(); err != nil {
		log.Errorf("faces: %s while finding orphan people", err)
	} else if l := len(orphans); l == 0 {
		log.Infof("faces: found no orphan people")
	} else if !fix {
		log.Infof("faces: found %s", english.Plural(l, "orphan person", "orphan people"))
	} else if err := orphans.Delete(); err != nil {
		log.Errorf("faces: failed fixing %s: %s", english.Plural(l, "orphan person", "orphan people"), err)
	} else {
		log.Infof("faces: removed %s", english.Plural(l, "orphan person", "orphan people"))
	}

	return nil
}
