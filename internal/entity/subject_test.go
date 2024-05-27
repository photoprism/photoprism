package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestSubject_TableName(t *testing.T) {
	m := &Subject{}
	assert.Contains(t, m.TableName(), "subjects")
}

func TestNewSubject(t *testing.T) {
	t.Run("Jens_Mander", func(t *testing.T) {
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
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "Create Me", m.SubjName)
		assert.Equal(t, "create-me", m.SubjSlug)
	})
	t.Run("ExistingPerson", func(t *testing.T) {
		m := SubjectFixtures.Pointer("john-doe")
		result := FirstOrCreateSubject(m)

		if result == nil {
			t.Fatal("result should not be nil")
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

		if err := m.Updates(Subject{SubjName: "UpdatedName", SubjType: "UpdatedType"}); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "UpdatedName", m.SubjName)
			assert.Equal(t, "UpdatedType", m.SubjType)
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
