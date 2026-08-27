package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/event"

	"github.com/photoprism/photoprism/internal/form"
)

func TestSubject_TableName(t *testing.T) {
	m := &Subject{}
	assert.Contains(t, m.TableName(), "subjects")
}

func TestNewSubject(t *testing.T) {
	t.Run("JensMander", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjPerson, SrcAuto)
		assert.Equal(t, "Jens Mander", m.SubjName)
		assert.Equal(t, "jens-mander", m.SubjSlug)
		assert.Equal(t, "person", m.SubjType)
	})
	t.Run("SubjectTypeEmpty", func(t *testing.T) {
		m := NewSubject("Anna Mander", "", SrcAuto)
		assert.Equal(t, "Anna Mander", m.SubjName)
		assert.Equal(t, "anna-mander", m.SubjSlug)
		assert.Equal(t, "person", m.SubjType)
	})
	t.Run("SubjectNameEmpty", func(t *testing.T) {
		m := NewSubject("", "", SrcAuto)
		assert.Nil(t, m)
	})
}

func TestSubject_SetName(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjPerson, SrcAuto)

		assert.Equal(t, "Jens Mander", m.SubjName)
		assert.Equal(t, "jens-mander", m.SubjSlug)

		if err := m.SetName("Foo McBar"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo McBar", m.SubjName)
		assert.Equal(t, "foo-mcbar", m.SubjSlug)
	})
	t.Run("Empty", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjPerson, SrcAuto)

		assert.Equal(t, "Jens Mander", m.SubjName)
		assert.Equal(t, "jens-mander", m.SubjSlug)

		err := m.SetName("")

		if err == nil {
			t.Fatal(err)
		}

		assert.Equal(t, "name must not be empty", err.Error())
		assert.Equal(t, "Jens Mander", m.SubjName)
	})
	t.Run("NoChange", func(t *testing.T) {
		m := NewSubject("Anna Mander", SubjPerson, SrcAuto)

		assert.Equal(t, "Anna Mander", m.SubjName)

		if err := m.SetName("Anna Mander"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Anna Mander", m.SubjName)
	})
}

func TestFirstOrCreatePerson(t *testing.T) {
	t.Run("NotYetExistingPerson", func(t *testing.T) {
		m := NewSubject("Create Me", SubjPerson, SrcAuto)
		result := FirstOrCreateSubject(m)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, "Create Me", m.SubjName)
		assert.Equal(t, "create-me", m.SubjSlug)
	})
	t.Run("ExistingPerson", func(t *testing.T) {
		m := SubjectFixtures.Pointer("john-doe")
		result := FirstOrCreateSubject(m)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, "John Doe", m.SubjName)
		assert.Equal(t, "john-doe", m.SubjSlug)
		assert.Equal(t, "Short Note", m.SubjNotes)
	})
}

func TestSubject_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewSubject("Save Me", SubjPerson, SrcAuto)
		initialDate := m.UpdatedAt
		err := m.Save()

		if err != nil {
			t.Fatal(err)
		}

		afterDate := m.UpdatedAt

		assert.True(t, afterDate.After(initialDate))

	})
}

func TestSubject_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjPerson, SrcAuto)
		err := m.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, m.Deleted())

		var subj Subjects

		if err := Db().Where("subj_name = ?", m.SubjName).Find(&subj).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, subj, 1)

		err = m.Delete()
		if err != nil {
			t.Fatal(err)
		}

		if err := Db().Where("subj_name = ?", m.SubjName).Find(&subj).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, subj, 0)
	})
	t.Run("AlreadyDeleted", func(t *testing.T) {
		m := NewSubject("Jens Doe", SubjPerson, SrcAuto)

		err := m.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, m.Deleted())

		time := Now()
		m.DeletedAt = &time

		assert.True(t, m.Deleted())

		assert.Nil(t, m.Delete())
	})
}

func TestSubject_Restore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var deleteTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		m := &Subject{DeletedAt: &deleteTime, SubjType: SubjPerson, SubjName: "ToBeRestored"}
		err := m.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, m.Deleted())

		err = m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, m.Deleted())
	})
	t.Run("SubjectNotDeleted", func(t *testing.T) {
		m := &Subject{DeletedAt: nil, SubjType: SubjPerson, SubjName: "NotDeleted1234"}
		err := m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, m.Deleted())
	})
}

func TestFindSubject(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewSubject("Find Me", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		if s := FindSubject(m.SubjName); s != nil {
			t.Fatal("result must be nil")
		}

		if s := FindSubject(m.SubjUID); s != nil {
			assert.Equal(t, "Find Me", s.SubjName)
		} else {
			t.Fatal("result must not be nil")
		}
	})
	t.Run("Nil", func(t *testing.T) {
		r := FindSubject("XXX")
		assert.Nil(t, r)
	})
	t.Run("EmptyUid", func(t *testing.T) {
		r := FindSubject("")
		assert.Nil(t, r)
	})
}

func TestFindSubjectByName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := FindSubjectByName("John Doe", false)
		assert.Equal(t, "John Doe", r.SubjName)
	})
	t.Run("NameEmpty", func(t *testing.T) {
		assert.Nil(t, FindSubjectByName("", false))
	})
	t.Run("RestoreDeleted", func(t *testing.T) {
		m := NewSubject("Jim Doe", SubjPerson, SrcAuto)

		time := Now()
		m.DeletedAt = &time

		err := m.Save()
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, m.Deleted())

		r := FindSubjectByName("Jim Doe", false)
		assert.Equal(t, "Jim Doe", r.SubjName)
		assert.True(t, r.Deleted())

		r = FindSubjectByName("Jim Doe", true)
		assert.Equal(t, "Jim Doe", r.SubjName)
		assert.False(t, r.Deleted())
	})
}

func TestSubject_Links(t *testing.T) {
	t.Run("NoResult", func(t *testing.T) {
		m := SubjectFixtures.Pointer("john-doe")
		links := m.Links()
		assert.Empty(t, links)
	})
}

func TestSubject_String(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var m *Subject
		assert.Equal(t, "Subject<nil>", m.String())
		assert.Equal(t, "Subject<nil>", fmt.Sprintf("%s", m))
	})
	t.Run("New", func(t *testing.T) {
		m := &Subject{}
		assert.Equal(t, "*Subject", m.String())
		assert.Equal(t, "*Subject", fmt.Sprintf("%s", m))
	})
	t.Run("JohnDoe", func(t *testing.T) {
		m := SubjectFixtures.Pointer("john-doe")
		assert.Equal(t, "John Doe", m.String())
	})
}

func TestSubject_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewSubject("Update Me", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		if err := m.Update("SubjName", "Updated Name"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Updated Name", m.SubjName)
		}
	})

}

// TODO fails on mariadb
func TestSubject_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewSubject("Update Me", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		if err := m.Updates(Subject{SubjName: "UpdatedName", SubjType: "newtype"}); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "UpdatedName", m.SubjName)
			assert.Equal(t, "newtype", m.SubjType)
		}
	})

}

func TestSubject_Visible(t *testing.T) {
	t.Run("Hidden", func(t *testing.T) {
		subj := NewSubject("Jens Mander", SubjPerson, SrcManual)
		assert.True(t, subj.Visible())
		subj.SubjHidden = true
		assert.False(t, subj.Visible())
	})
	t.Run("Private", func(t *testing.T) {
		subj := NewSubject("Jens Mander", SubjPerson, SrcManual)
		assert.True(t, subj.Visible())
		subj.SubjPrivate = true
		assert.False(t, subj.Visible())
	})
	t.Run("Excluded", func(t *testing.T) {
		subj := NewSubject("Jens Mander", SubjPerson, SrcManual)
		assert.True(t, subj.Visible())
		subj.SubjExcluded = true
		assert.False(t, subj.Visible())
	})
}

func TestSubject_SaveForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		subj := NewSubject("Save Form Test", SubjPerson, SrcManual)

		assert.Equal(t, "Save Form Test", subj.SubjName)
		assert.Equal(t, "save-form-test", subj.SubjSlug)
		assert.Equal(t, false, subj.SubjHidden)
		assert.Equal(t, true, subj.IsPerson())

		if err := subj.Create(); err != nil {
			t.Fatal(err)
		}

		subjForm, err := form.NewSubject(subj)

		if err != nil {
			t.Fatal(err)
		}

		subjForm.SubjName = "Bill Gates III"
		subjForm.SubjHidden = true
		subjForm.SubjFavorite = true

		t.Logf("Subject Form: %#v", subjForm)

		if changed, err := subj.SaveForm(subjForm); err != nil {
			t.Fatal(err)
		} else if !changed {
			t.Fatal("subject must be changed")
		}

		assert.Equal(t, "Bill Gates III", subj.SubjName)
		assert.Equal(t, "bill-gates-iii", subj.SubjSlug)
		assert.Equal(t, true, subj.SubjHidden)
		assert.Equal(t, true, subj.SubjFavorite)
		assert.Equal(t, true, subj.IsPerson())

		if err := subj.Delete(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NoUid", func(t *testing.T) {
		subj := NewSubject("No Uid", SubjPerson, SrcManual)

		assert.Equal(t, true, subj.IsPerson())

		if err := subj.Create(); err != nil {
			t.Fatal(err)
		}

		subjForm, err := form.NewSubject(subj)

		if err != nil {
			t.Fatal(err)
		}

		subj.SubjUID = ""

		changed, err := subj.SaveForm(subjForm)

		assert.Contains(t, err.Error(), "no uid")
		assert.False(t, changed)
	})
	t.Run("ManualThumb", func(t *testing.T) {
		subj := NewSubject("Cover Person", SubjPerson, SrcAuto)
		subj.SubjFavorite = true
		if err := subj.Save(); err != nil {
			t.Fatal(err)
		}

		subjForm, err := form.NewSubject(subj)
		if err != nil {
			t.Fatal(err)
		}

		hash := "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd"
		subjForm.Thumb = hash
		subjForm.ThumbSrc = "manual"

		changed, err := subj.SaveForm(subjForm)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, changed)
		assert.Equal(t, hash, subj.Thumb)
		assert.Equal(t, SrcManual, subj.ThumbSrc)

		if err := subj.Delete(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("DefaultThumbSrc", func(t *testing.T) {
		subj := NewSubject("Auto Src", SubjPerson, SrcAuto)
		if err := subj.Save(); err != nil {
			t.Fatal(err)
		}

		subjForm, err := form.NewSubject(subj)
		if err != nil {
			t.Fatal(err)
		}

		subjForm.Thumb = "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c"
		subjForm.ThumbSrc = ""

		changed, err := subj.SaveForm(subjForm)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid thumb")
		assert.False(t, changed)
		assert.Equal(t, "", subj.Thumb)
		assert.Equal(t, "", subj.ThumbSrc)

		if err := subj.Delete(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InvalidThumbSrc", func(t *testing.T) {
		subj := NewSubject("Invalid Src", SubjPerson, SrcAuto)
		if err := subj.Save(); err != nil {
			t.Fatal(err)
		}

		subjForm, err := form.NewSubject(subj)
		if err != nil {
			t.Fatal(err)
		}

		subjForm.Thumb = "6f6cbaa6ae8ead9da7ee99ab66aca1ae7eed8d5c-0910162fd2fd"
		subjForm.ThumbSrc = "invalid-src"

		_, err = subj.SaveForm(subjForm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid thumb source")

		if err := subj.Delete(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSubject_UpdateName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewSubject("Test Person", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Test Person", m.SubjName)
		assert.Equal(t, "test-person", m.SubjSlug)

		if s, err := m.UpdateName("New New"); err != nil {
			t.Fatal(err)
		} else if s == nil {
			t.Fatal("subject is nil")
		} else {
			assert.Equal(t, "New New", m.SubjName)
			assert.Equal(t, "new-new", m.SubjSlug)
			assert.Equal(t, "New New", s.SubjName)
			assert.Equal(t, "new-new", s.SubjSlug)
		}
	})
	t.Run("PublishesUidOnlyUpdatedEvents", func(t *testing.T) {
		m := NewSubject("Uid Only Person", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		sub := event.Subscribe("subjects.updated", "people.updated")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		if _, err := m.UpdateName("Uid Only Renamed"); err != nil {
			t.Fatal(err)
		}

		// A rename publishes one subjects.updated and one people.updated event,
		// both carrying only the subject UID.
		for _, expected := range []string{"subjects.updated", "people.updated"} {
			select {
			case msg := <-sub.Receiver:
				assert.Equal(t, expected, msg.Name)
				uids, ok := msg.Fields["entities"].([]string)
				assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
				assert.Equal(t, []string{m.SubjUID}, uids)
			case <-time.After(2 * time.Second):
				t.Fatalf("expected one %s event", expected)
			}
		}
	})
	t.Run("SubjNameEmpty", func(t *testing.T) {
		m := NewSubject("Empty", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		m.SubjName = ""

		assert.Equal(t, "", m.SubjName)

		s, err := m.UpdateName("hans")
		assert.Equal(t, "", s.SubjName)
		assert.Error(t, err)
	})
	t.Run("SubjUidEmpty", func(t *testing.T) {
		m := NewSubject("Janet", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		m.SubjUID = ""

		assert.Equal(t, "", m.SubjUID)

		s, err := m.UpdateName("hans")
		assert.Equal(t, "", s.SubjUID)
		assert.Error(t, err)
	})
	t.Run("EmptyName", func(t *testing.T) {
		m := NewSubject("Test Person2", SubjPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Test Person2", m.SubjName)
		assert.Equal(t, "test-person2", m.SubjSlug)

		if s, err := m.UpdateName(""); err == nil {
			t.Error("error expected")
		} else if s == nil {
			t.Fatal("subject is nil")
		} else {
			assert.Equal(t, "Test Person2", m.SubjName)
			assert.Equal(t, "test-person2", m.SubjSlug)
			assert.Equal(t, "Test Person2", s.SubjName)
			assert.Equal(t, "test-person2", s.SubjSlug)
		}
	})
}

func TestSubject_RefreshPhotos(t *testing.T) {
	subj := SubjectFixtures.Get("john-doe")

	if err := subj.RefreshPhotos(); err != nil {
		t.Fatal(err)
	}
}

func TestSubject_DeletePermanently(t *testing.T) {
	m := NewSubject("Tim Doe", SubjPerson, SrcAuto)

	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Tim Doe", m.SubjName)
	assert.Empty(t, m.DeletedAt)
	assert.NotEmpty(t, FindSubject(m.SubjUID))

	assert.Nil(t, m.DeletePermanently())

	time := Now()
	m.DeletedAt = &time

	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, m.DeletedAt)
	assert.NotEmpty(t, FindSubject(m.SubjUID))

	if err := m.DeletePermanently(); err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, m.DeletedAt)
	assert.Empty(t, FindSubject(m.SubjUID))
}

func TestReassignSubject(t *testing.T) {
	t.Run("OtherPersonOwnsName", func(t *testing.T) {
		subj := FirstOrCreateSubject(NewSubject("Reassign Lookup Source", SubjPerson, SrcManual))
		other := FirstOrCreateSubject(NewSubject("Reassign Lookup Target", SubjPerson, SrcManual))

		if subj == nil || other == nil {
			t.Fatal("failed creating test subjects")
		}

		found := ReassignSubject(subj, "Reassign Lookup Target")

		if assert.NotNil(t, found) {
			assert.Equal(t, other.SubjUID, found.SubjUID)
		}
	})
	t.Run("NameIsUnused", func(t *testing.T) {
		subj := FirstOrCreateSubject(NewSubject("Reassign Lookup Unused", SubjPerson, SrcManual))

		if subj == nil {
			t.Fatal("failed creating test subject")
		}

		assert.Nil(t, ReassignSubject(subj, "Reassign Lookup Nobody Has This"))
	})
	t.Run("SamePerson", func(t *testing.T) {
		subj := FirstOrCreateSubject(NewSubject("Reassign Lookup Self", SubjPerson, SrcManual))

		if subj == nil {
			t.Fatal("failed creating test subject")
		}

		assert.Nil(t, ReassignSubject(subj, "Reassign Lookup Self"))
	})
	t.Run("EmptyName", func(t *testing.T) {
		subj := FirstOrCreateSubject(NewSubject("Reassign Lookup Empty", SubjPerson, SrcManual))

		if subj == nil {
			t.Fatal("failed creating test subject")
		}

		assert.Nil(t, ReassignSubject(subj, ""))
		assert.Nil(t, ReassignSubject(subj, "   "))
	})
	t.Run("NilSubject", func(t *testing.T) {
		assert.Nil(t, ReassignSubject(nil, "Reassign Lookup Target"))
	})
	t.Run("DeletedPersonOwnsName", func(t *testing.T) {
		subj := FirstOrCreateSubject(NewSubject("Reassign Lookup Live", SubjPerson, SrcManual))
		gone := FirstOrCreateSubject(NewSubject("Reassign Lookup Gone", SubjPerson, SrcManual))

		if subj == nil || gone == nil {
			t.Fatal("failed creating test subjects")
		}

		if err := gone.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, ReassignSubject(subj, "Reassign Lookup Gone"))
	})
}

// TestSubject_MergeWith_ClearsCollisions pins that stating two subjects are one person also
// retracts the geometry that treating them as two produced.
//
// A collision narrows a cluster's accept distance and nothing else widens it again, so without
// this the clusters stay gated against faces that the merge just established do belong to them.
func TestSubject_MergeWith_ClearsCollisions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		typo := NewSubject("Merge Collision Typo", SubjPerson, SrcManual)
		require.NotNil(t, typo)
		require.NoError(t, typo.Create())

		keep := NewSubject("Merge Collision Keep", SubjPerson, SrcManual)
		require.NotNil(t, keep)
		require.NoError(t, keep.Create())

		// One narrowed cluster per subject: the merge has to reach both, because the collision was
		// recorded on each side of the same false premise.
		typoFace := &Face{
			ID: "MERGECOLLISION0000000000000000C1", SubjUID: typo.SubjUID, FaceSrc: SrcManual,
			SampleRadius: 0.3, Samples: 4, Collisions: 1, CollisionRadius: 0.64,
			FaceKind: int(face.AmbiguousFace),
		}
		keepFace := &Face{
			ID: "MERGECOLLISION0000000000000000C2", SubjUID: keep.SubjUID, FaceSrc: SrcManual,
			SampleRadius: 0.3, Samples: 4, Collisions: 2, CollisionRadius: 0.802,
		}

		require.NoError(t, Db().Create(typoFace).Error)
		require.NoError(t, Db().Create(keepFace).Error)

		t.Cleanup(func() {
			UnscopedDb().Delete(&Face{}, "id IN (?)", []string{typoFace.ID, keepFace.ID})
			UnscopedDb().Delete(&Subject{}, "subj_uid IN (?)", []string{typo.SubjUID, keep.SubjUID})
		})

		require.NoError(t, typo.MergeWith(keep))

		var merged, kept Face

		require.NoError(t, UnscopedDb().Where("id = ?", typoFace.ID).First(&merged).Error)
		assert.Equal(t, keep.SubjUID, merged.SubjUID, "the cluster moves to the surviving subject")
		assert.Zero(t, merged.Collisions)
		assert.Zero(t, merged.CollisionRadius)
		assert.Equal(t, int(face.RegularFace), merged.FaceKind, "and takes part in matching again")
		assert.Nil(t, merged.MatchedAt, "so the markers it refused are compared against it again")

		require.NoError(t, UnscopedDb().Where("id = ?", keepFace.ID).First(&kept).Error)
		assert.Zero(t, kept.Collisions, "the surviving subject's own clusters are cleared too")
		assert.Zero(t, kept.CollisionRadius)
	})
}
