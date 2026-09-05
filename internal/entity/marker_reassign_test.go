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
// of the photo editor.
func TestMarkerSaveForm_Reassign(t *testing.T) {
	ValidateFixtures(t)
	t.Run("MarkerLinkedToOtherSubject", func(t *testing.T) {
		// Two existing people, each with a marker of their own.
		subjA := FirstOrCreateSubject(NewSubject("Reassign Person A", SubjPerson, SrcManual))
		subjB := FirstOrCreateSubject(NewSubject("Reassign Person B", SubjPerson, SrcManual))

		if subjA == nil || subjB == nil {
			t.Fatal("failed creating test subjects")
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(subjA).Error)
			assert.NoError(t, UnscopedDb().Delete(subjB).Error)
		})
		// Own photo and file, so the added markers cannot affect other tests.
		photo := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoName: "reassign-test", PhotoPath: "2790/07"}

		if err := photo.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID).Error)
			assert.NoError(t, UnscopedDb().Delete(&photo).Error)
		})

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
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&file).Error) })
		markerA := NewMarker(file, crop.Area{Name: "face", X: 0.31, Y: 0.31, W: 0.05, H: 0.05}, subjA.SubjUID, SrcImage, MarkerFace, 100, 65)
		markerA.MarkerName = subjA.SubjName
		markerA.SubjSrc = SrcManual

		if err := markerA.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(markerA).Error) })

		markerB := NewMarker(file, crop.Area{Name: "face", X: 0.61, Y: 0.61, W: 0.05, H: 0.05}, subjB.SubjUID, SrcImage, MarkerFace, 100, 65)
		markerB.MarkerName = subjB.SubjName
		markerB.SubjSrc = SrcManual

		if err := markerB.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(markerB).Error) })

		// A third marker, linked to Person A, is assigned to Person B: this is what the
		// people tab sends when its state says the marker has no subject yet.
		markerC := NewMarker(file, crop.Area{Name: "face", X: 0.11, Y: 0.11, W: 0.05, H: 0.05}, subjA.SubjUID, SrcImage, MarkerFace, 100, 65)
		markerC.MarkerName = subjA.SubjName
		markerC.SubjSrc = SrcManual

		if err := markerC.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(markerC).Error) })

		frm, err := form.NewMarker(*markerC)

		if err != nil {
			t.Fatal(err)
		}

		frm.MarkerName = subjB.SubjName
		frm.SubjSrc = SrcManual

		changed, err := markerC.SaveForm(frm)

		assert.NoError(t, err)
		assert.True(t, changed)

		// Only the edited marker moves; both people survive under their own names.
		foundA := FindSubject(subjA.SubjUID)
		foundB := FindSubject(subjB.SubjUID)

		if assert.NotNil(t, foundA, "person A must still exist") {
			assert.False(t, foundA.Deleted(), "person A must not be flagged as missing")
			assert.Equal(t, subjA.SubjName, foundA.SubjName, "person A must keep their name")
		}

		if assert.NotNil(t, foundB, "person B must still exist") {
			assert.Equal(t, subjB.SubjName, foundB.SubjName, "person B must keep their name")
		}

		markers := map[string]Marker{}

		for _, uid := range []string{markerA.MarkerUID, markerB.MarkerUID, markerC.MarkerUID} {
			found := Marker{}
			if err = UnscopedDb().First(&found, "marker_uid = ?", uid).Error; err != nil {
				t.Fatal(err)
			}
			markers[uid] = found
		}

		assert.Equal(t, subjB.SubjUID, markers[markerC.MarkerUID].SubjUID, "edited marker must be reassigned to person B")
		assert.Equal(t, subjB.SubjName, markers[markerC.MarkerUID].MarkerName, "edited marker must carry person B's name")
		assert.Equal(t, subjA.SubjUID, markers[markerA.MarkerUID].SubjUID, "marker of person A must not be reassigned")
		assert.Equal(t, subjB.SubjUID, markers[markerB.MarkerUID].SubjUID, "marker of person B must not be reassigned")
	})
	t.Run("MarkerLinkedToUnusedName", func(t *testing.T) {
		// Naming a linked marker something nobody else owns is an ordinary rename.
		subj := FirstOrCreateSubject(NewSubject("Rename Person Before", SubjPerson, SrcManual))

		if subj == nil {
			t.Fatal("failed creating test subject")
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(subj).Error) })
		photo := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoName: "rename-test", PhotoPath: "2790/08"}

		if err := photo.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&photo).Error) })

		file := File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     rnd.GenerateUID(FileUID),
			FileName:    "2790/08/rename-test.jpg",
			FileHash:    "2cad9168fa6acc5c5c2965ddf6ec465ca42fd811",
			FileType:    "jpg",
			FileWidth:   720,
			FileHeight:  480,
			FilePrimary: true,
		}

		if err := file.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID).Error)
			assert.NoError(t, UnscopedDb().Delete(&file).Error)
		})

		marker := NewMarker(file, crop.Area{Name: "face", X: 0.41, Y: 0.41, W: 0.05, H: 0.05}, subj.SubjUID, SrcImage, MarkerFace, 100, 65)
		marker.MarkerName = subj.SubjName
		marker.SubjSrc = SrcManual

		if err := marker.Create(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(marker).Error) })

		frm, err := form.NewMarker(*marker)

		if err != nil {
			t.Fatal(err)
		}

		frm.MarkerName = "Rename Person After"
		frm.SubjSrc = SrcManual

		changed, err := marker.SaveForm(frm)

		assert.NoError(t, err)
		assert.True(t, changed)

		found := FindSubject(subj.SubjUID)

		if assert.NotNil(t, found, "renamed person must still exist") {
			assert.Equal(t, "Rename Person After", found.SubjName, "person must be renamed in place")
		}

		assert.Equal(t, subj.SubjUID, marker.SubjUID, "marker must stay linked to the same person")
	})
}
