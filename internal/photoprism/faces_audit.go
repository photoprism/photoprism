package photoprism

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Audit face clusters and subjects.
func (w *Faces) Audit(fix bool) (err error) {
	invalidFaces, invalidSubj, err := query.MarkersWithNonExistentReferences()

	if err != nil {
		return err
	}

	subj, err := query.SubjectMap()

	if err != nil {
		log.Error(err)
	}

	if n := len(subj); n == 0 {
		log.Infof("found no subjects")
	} else {
		log.Infof("%d known subjects", n)
	}

	// Fix non-existent marker subjects references?
	if n := len(invalidSubj); n == 0 {
		log.Infof("found no invalid marker subjects")
	} else if !fix {
		log.Infof("%d markers with non-existent subjects", n)
	} else if removed, err := query.RemoveNonExistentMarkerSubjects(); err != nil {
		log.Infof("removed %d / %d markers with non-existent subjects", removed, n)
	} else {
		log.Error(err)
	}

	// Fix non-existent marker face references?
	if n := len(invalidFaces); n == 0 {
		log.Infof("found no invalid marker faces")
	} else if !fix {
		log.Infof("%d markers with non-existent faces", n)
	} else if removed, err := query.RemoveNonExistentMarkerFaces(); err != nil {
		log.Infof("removed %d / %d markers with non-existent faces", removed, n)
	} else {
		log.Error(err)
	}

	conflicts := 0
	resolved := 0

	faces, err := query.Faces(true, false)

	if err != nil {
		return err
	}

	faceMap := make(map[string]entity.Face)

	for _, f1 := range faces {
		faceMap[f1.ID] = f1

		for _, f2 := range faces {
			if matched, dist := f1.Match(entity.Embeddings{f2.Embedding()}); matched {
				if f1.SubjectUID == f2.SubjectUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.ClusterRadius

				log.Infof("face %s: conflict at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjectUID != "" {
					log.Infof("face %s: subject %s (%s %s)", f1.ID, txt.Quote(subj[f1.SubjectUID].SubjectName), f1.SubjectUID, entity.SrcString(f1.FaceSrc))
				} else {
					log.Infof("face %s: no subject (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjectUID != "" {
					log.Infof("face %s: subject %s (%s %s)", f2.ID, txt.Quote(subj[f2.SubjectUID].SubjectName), f2.SubjectUID, entity.SrcString(f2.FaceSrc))
				} else {
					log.Infof("face %s: no subject (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				if !fix {
					// Do nothing.
				} else if ok, err := f1.ResolveCollision(entity.Embeddings{f2.Embedding()}); err != nil {
					log.Errorf("face %s: %s", f1.ID, err)
				} else if ok {
					log.Infof("face %s: collision has been resolved", f1.ID)
					resolved++
				} else {
					log.Infof("face %s: collision could not be resolved", f1.ID)
				}
			}
		}
	}

	if conflicts == 0 {
		log.Infof("found no conflicting face clusters")
	} else if !fix {
		log.Infof("%d conflicting face clusters", conflicts)
	} else {
		log.Infof("%d conflicting face clusters, %d resolved", conflicts, resolved)
	}

	if markers, err := query.MarkersWithSubjectConflict(); err != nil {
		log.Error(err)
	} else {
		for _, m := range markers {
			log.Infof("marker %d: %s subject %s conflicts with face %s subject %s", m.ID, entity.SrcString(m.SubjectSrc), txt.Quote(subj[m.SubjectUID].SubjectName), m.FaceID, txt.Quote(subj[faceMap[m.FaceID].SubjectUID].SubjectName))
		}
	}

	return nil
}
