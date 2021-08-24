package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubject_TableName(t *testing.T) {
	m := &Subject{}
	assert.Contains(t, m.TableName(), "subjects")
}

func TestNewPerson(t *testing.T) {
	t.Run("Jens_Mander", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjectPerson, SrcAuto)
		assert.Equal(t, "Jens Mander", m.SubjectName)
		assert.Equal(t, "jens-mander", m.SubjectSlug)
	})
}

func TestSubject_SetName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjectPerson, SrcAuto)

		assert.Equal(t, "Jens Mander", m.SubjectName)
		assert.Equal(t, "jens-mander", m.SubjectSlug)

		if err := m.SetName("Foo McBar"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo McBar", m.SubjectName)
		assert.Equal(t, "foo-mcbar", m.SubjectSlug)
	})
}

func TestFirstOrCreatePerson(t *testing.T) {
	m := NewSubject("Create Me", SubjectPerson, SrcAuto)
	result := FirstOrCreateSubject(m)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	assert.Equal(t, "Create Me", m.SubjectName)
	assert.Equal(t, "create-me", m.SubjectSlug)
}

func TestSubject_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewSubject("Save Me", SubjectPerson, SrcAuto)
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
	t.Run("success", func(t *testing.T) {
		m := NewSubject("Jens Mander", SubjectPerson, SrcAuto)
		err := m.Save()
		assert.False(t, m.Deleted())

		var subj Subjects

		if err := Db().Where("subject_name = ?", m.SubjectName).Find(&subj).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, subj, 1)

		err = m.Delete()
		if err != nil {
			t.Fatal(err)
		}

		if err := Db().Where("subject_name = ?", m.SubjectName).Find(&subj).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, subj, 0)
	})
}

func TestSubject_Restore(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var deleteTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		m := &Subject{DeletedAt: &deleteTime, SubjectName: "ToBeRestored"}
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
	t.Run("subject not deleted", func(t *testing.T) {
		m := &Subject{DeletedAt: nil, SubjectName: "NotDeleted1234"}
		err := m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, m.Deleted())
	})
}

func TestFindSubject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewSubject("Find Me", SubjectPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		if s := FindSubject(m.SubjectName); s != nil {
			t.Fatal("result must be nil")
		}

		if s := FindSubject(m.SubjectUID); s != nil {
			assert.Equal(t, "Find Me", s.SubjectName)
		} else {
			t.Fatal("result must not be nil")
		}
	})
	t.Run("unknown person", func(t *testing.T) {
		if s := FindSubject(UnknownPerson.SubjectUID); s != nil {
			assert.Equal(t, "", s.SubjectName)
		} else {
			t.Fatal("result must not be nil")
		}
	})
	t.Run("nil", func(t *testing.T) {
		r := FindSubject("XXX")
		assert.Nil(t, r)
	})
}

func TestSubject_Links(t *testing.T) {
	t.Run("no-result", func(t *testing.T) {
		m := UnknownPerson
		links := m.Links()
		assert.Empty(t, links)
	})
}

func TestSubject_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewSubject("Update Me", SubjectPerson, SrcAuto)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		if err := m.Update("SubjectName", "Updated Name"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Updated Name", m.SubjectName)
		}
	})

}
