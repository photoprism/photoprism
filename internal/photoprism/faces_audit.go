package photoprism

import (
	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/sanitize"
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
		log.Infof("found no subjects")
	} else {
		log.Infof("found %s", english.Plural(n, "subject", "subjects"))
	}

	// Fix non-existent marker subjects references?
	if n := len(invalidSubj); n == 0 {
		log.Infof("found no invalid marker subjects")
	} else if !fix {
		log.Infof("%s with non-existent subjects", english.Plural(n, "marker", "markers"))
	} else if removed, err := query.RemoveNonExistentMarkerSubjects(); err != nil {
		log.Errorf("faces: %s (remove orphan subjects)", err)
	} else if removed > 0 {
		log.Infof("removed %d / %d markers with non-existent subjects", removed, n)
	}

	// Fix non-existent marker face references?
	if n := len(invalidFaces); n == 0 {
		log.Infof("found no invalid marker faces")
	} else if !fix {
		log.Infof("%s with non-existent faces", english.Plural(n, "marker", "markers"))
	} else if removed, err := query.RemoveNonExistentMarkerFaces(); err != nil {
		log.Errorf("faces: %s (remove orphan embeddings)", err)
	} else if removed > 0 {
		log.Infof("removed %d / %d markers with non-existent faces", removed, n)
	}

	conflicts := 0
	resolved := 0

	faces, err := query.Faces(true, false, false)

	if err != nil {
		return err
	}

	faceMap := make(map[string]entity.Face)

	for _, f1 := range faces {
		faceMap[f1.ID] = f1

		for _, f2 := range faces {
			if matched, dist := f1.Match(face.Embeddings{f2.Embedding()}); matched {
				if f1.SubjUID == f2.SubjUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.MatchDist

				log.Infof("face %s: ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjUID != "" {
					log.Infof("face %s: subject %s (%s %s)", f1.ID, entity.SubjNames.Log(f1.SubjUID), f1.SubjUID, entity.SrcString(f1.FaceSrc))
				} else {
					log.Infof("face %s: has no subject (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjUID != "" {
					log.Infof("face %s: subject %s (%s %s)", f2.ID, entity.SubjNames.Log(f2.SubjUID), f2.SubjUID, entity.SrcString(f2.FaceSrc))
				} else {
					log.Infof("face %s: has no subject (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				if !fix {
					// Do nothing.
				} else if ok, err := f1.ResolveCollision(face.Embeddings{f2.Embedding()}); err != nil {
					log.Errorf("conflict resolution for %s failed, face id %s has collisions with other persons (%s)", entity.SubjNames.Log(f1.SubjUID), f1.ID, err)
				} else if ok {
					log.Infof("successful conflict resolution for %s, face id %s had collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
					resolved++
				} else {
					log.Infof("conflict resolution for %s not successful, face id %s still has collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
				}
			}
		}
	}

	if conflicts == 0 {
		log.Infof("found no ambiguous subjects")
	} else if !fix {
		log.Infof("%s", english.Plural(conflicts, "ambiguous subject", "ambiguous subjects"))
	} else {
		log.Infof("%s, %d resolved", english.Plural(conflicts, "ambiguous subject", "ambiguous subjects"), resolved)
	}

	if markers, err := query.MarkersWithSubjectConflict(); err != nil {
		log.Errorf("faces: %s (find marker conflicts)", err)
	} else {
		for _, m := range markers {
			log.Infof("marker %s: %s subject %s conflicts with face %s subject %s", m.MarkerUID, entity.SrcString(m.SubjSrc), sanitize.Log(subj[m.SubjUID].SubjName), m.FaceID, sanitize.Log(subj[faceMap[m.FaceID].SubjUID].SubjName))
		}
	}

	// Find and fix orphan face clusters.
	if orphans, err := entity.OrphanFaces(); err != nil {
		log.Errorf("%s while finding orphan face clusters", err)
	} else if l := len(orphans); l == 0 {
		log.Infof("found no orphan face clusters")
	} else if !fix {
		log.Infof("found %s", english.Plural(l, "orphan face cluster", "orphan face clusters"))
	} else if err := orphans.Delete(); err != nil {
		log.Errorf("failed removing %s: %s", english.Plural(l, "orphan face cluster", "orphan face clusters"), err)
	} else {
		log.Infof("removed %s", english.Plural(l, "orphan face cluster", "orphan face clusters"))
	}

	// Find and fix orphan people.
	if orphans, err := entity.OrphanPeople(); err != nil {
		log.Errorf("%s while finding orphan people", err)
	} else if l := len(orphans); l == 0 {
		log.Infof("found no orphan people")
	} else if !fix {
		log.Infof("found %s", english.Plural(l, "orphan person", "orphan people"))
	} else if err := orphans.Delete(); err != nil {
		log.Errorf("failed fixing %s: %s", english.Plural(l, "orphan person", "orphan people"), err)
	} else {
		log.Infof("removed %s", english.Plural(l, "orphan person", "orphan people"))
	}

	return nil
}
