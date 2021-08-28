package photoprism

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Audit face clusters and subjects.
func (w *Faces) Audit(fix bool) (err error) {
	invalidFaces, invalidSubj, err := query.MarkersWithInvalidReferences()

	if err != nil {
		return err
	}

	subj, err := query.SubjectMap()

	if err != nil {
		log.Error(err)
	}

	log.Infof("%d subjects indexed", len(subj))

	log.Infof("%d markers with non-existent subjects", len(invalidSubj))

	log.Infof("%d markers with non-existent faces", len(invalidFaces))

	conflicts := 0

	faces, err := query.Faces(true, "")

	if err != nil {
		return err
	}

	faceMap := make(map[string]entity.Face)

	for _, f1 := range faces {
		faceMap[f1.ID] = f1

		for _, f2 := range faces {
			if ok, dist := f1.Match(entity.Embeddings{f2.Embedding()}); ok {
				if f1.SubjectUID == f2.SubjectUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.ClusterRadius

				log.Infof("face %s: ambiguous at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

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
				} else if reported, err := f1.ReportCollision(entity.Embeddings{f2.Embedding()}); err != nil {
					log.Errorf("face %s: %s", f1.ID, err)
				} else if reported {
					log.Infof("face %s: collision has been reported", f1.ID)
				} else {
					log.Infof("face %s: collision has not been reported", f1.ID)
				}
			}
		}
	}

	log.Infof("%d ambiguous faces clusters", conflicts)

	if markers, err := query.MarkersWithSubjectConflict(); err != nil {
		log.Error(err)
	} else {
		for _, m := range markers {
			log.Infof("marker %d: %s subject %s conflicts with face %s subject %s", m.ID, entity.SrcString(m.SubjectSrc), txt.Quote(subj[m.SubjectUID].SubjectName), m.FaceID, txt.Quote(subj[faceMap[m.FaceID].SubjectUID].SubjectName))
		}
	}

	return nil
}
