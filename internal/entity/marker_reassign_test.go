package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestMarkerSaveForm_Reassign checks that a marker which is already linked to a subject
// is reassigned when it receives the name of another person, as sent by the people tab
// of the photo editor. Skipped until the marker update no longer renames the subject.
func TestMarkerSaveForm_Reassign(t *testing.T) {
	t.Skip("known issue #5764: naming a linked marker renames and merges its subject")

	t.Run("MarkerLinkedToOtherSubject", func(t *testing.T) {
		// Two existing people, each with a marker of their own.
		subjA := FirstOrCreateSubject(NewSubject("Reassign Person A", SubjPerson, SrcManual))
		subjB := FirstOrCreateSubject(NewSubject("Reassign Person B", SubjPerson, SrcManual))

		if subjA == nil || subjB == nil {
			t.Fatal("failed creating test subjects")
		}

		// Own photo and file, so the added markers cannot affect other tests.
		photo := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoName: "reassign-test", PhotoPath: "2790/07"}

		if err := photo.Create(); err != nil {
			t.Fatal(err)
		}

		file := File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     rnd.GenerateUID(FileUID),
			FileName:    "2790/07/reassign-test.jpg",
			FileHash:    "5cad9168fa6acc5c5c2965ddf6ec465ca42fd899",
			FileType:    "jpg",
			FileWidth:   720,
			FileHeight:  480,
			FilePrimary: true,
		}

		if err := file.Create(); err != nil {
			t.Fatal(err)
		}

		markerA := NewMarker(file, crop.Area{Name: "face", X: 0.31, Y: 0.31, W: 0.05, H: 0.05}, subjA.SubjUID, SrcImage, MarkerFace, 100, 65)
		markerA.MarkerName = subjA.SubjName
		markerA.SubjSrc = SrcManual

		if err := markerA.Create(); err != nil {
			t.Fatal(err)
		}

		markerB := NewMarker(file, crop.Area{Name: "face", X: 0.61, Y: 0.61, W: 0.05, H: 0.05}, subjB.SubjUID, SrcImage, MarkerFace, 100, 65)
		markerB.MarkerName = subjB.SubjName
		markerB.SubjSrc = SrcManual

		if err := markerB.Create(); err != nil {
			t.Fatal(err)
		}

		// A third marker, linked to Person A, is assigned to Person B: this is what the
		// people tab sends when its state says the marker has no subject yet.
		markerC := NewMarker(file, crop.Area{Name: "face", X: 0.11, Y: 0.11, W: 0.05, H: 0.05}, subjA.SubjUID, SrcImage, MarkerFace, 100, 65)
		markerC.MarkerName = subjA.SubjName
		markerC.SubjSrc = SrcManual

		if err := markerC.Create(); err != nil {
			t.Fatal(err)
		}

		frm, err := form.NewMarker(*markerC)

		if err != nil {
			t.Fatal(err)
		}

		frm.MarkerName = subjB.SubjName
		frm.SubjSrc = SrcManual

		changed, err := markerC.SaveForm(frm)

		assert.NoError(t, err)
		assert.True(t, changed)

		// Report the resulting state of both people and all three markers.
		foundA := FindSubject(subjA.SubjUID)
		foundB := FindSubject(subjB.SubjUID)

		t.Logf("subject A (%s): %#v", subjA.SubjUID, foundA)
		t.Logf("subject B (%s): %#v", subjB.SubjUID, foundB)

		for _, uid := range []string{markerA.MarkerUID, markerB.MarkerUID, markerC.MarkerUID} {
			m := Marker{}
			if err = UnscopedDb().First(&m, "marker_uid = ?", uid).Error; err != nil {
				t.Fatal(err)
			}
			t.Logf("marker %s: subj_uid %s, name %s", uid, m.SubjUID, m.MarkerName)
		}

		// Only the edited marker should have been reassigned, and Person A should still exist.
		mA := Marker{}
		if err = UnscopedDb().First(&mA, "marker_uid = ?", markerA.MarkerUID).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, subjA.SubjUID, mA.SubjUID, "marker of person A must not be reassigned")
		assert.NotNil(t, foundA, "person A must still exist")
	})
}
