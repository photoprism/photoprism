package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPerson_TableName(t *testing.T) {
	m := &Person{}
	assert.Contains(t, m.TableName(), "people")
}

func TestNewPerson(t *testing.T) {
	t.Run("Jens_Mander", func(t *testing.T) {
		m := NewPerson("Jens Mander", SrcAuto, 0)
		assert.Equal(t, "Jens Mander", m.PersonName)
		assert.Equal(t, "jens-mander", m.PersonSlug)
	})
}

func TestPerson_SetName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewPerson("Jens Mander", SrcAuto, 0)

		assert.Equal(t, "Jens Mander", m.PersonName)
		assert.Equal(t, "jens-mander", m.PersonSlug)

		m.SetName("Foo McBar")

		assert.Equal(t, "Foo McBar", m.PersonName)
		assert.Equal(t, "foo-mcbar", m.PersonSlug)
	})
}

func TestFirstOrCreatePerson(t *testing.T) {
	m := NewPerson("Create Me", SrcAuto, 0)
	result := FirstOrCreatePerson(m)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	assert.Equal(t, "Create Me", m.PersonName)
	assert.Equal(t, "create-me", m.PersonSlug)
}

func TestPerson_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewPerson("Save Me", SrcAuto, 0)
		initialDate := m.UpdatedAt
		err := m.Save()

		if err != nil {
			t.Fatal(err)
		}

		afterDate := m.UpdatedAt

		assert.True(t, afterDate.After(initialDate))

	})
}

func TestPerson_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewPerson("Jens Mander", SrcAuto, 0)
		err := m.Save()
		assert.False(t, m.Deleted())

		var people People

		if err := Db().Where("person_name = ?", m.PersonName).Find(&people).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, people, 1)

		err = m.Delete()
		if err != nil {
			t.Fatal(err)
		}

		if err := Db().Where("person_name = ?", m.PersonName).Find(&people).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, people, 0)
	})
}

func TestPerson_Restore(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var deleteTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		m := &Person{DeletedAt: &deleteTime, PersonName: "ToBeRestored"}
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
	t.Run("person not deleted", func(t *testing.T) {
		m := &Person{DeletedAt: nil, PersonName: "NotDeleted1234"}
		err := m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, m.Deleted())
	})
}

func TestFindPerson(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewPerson("Find Me", SrcAuto, 0)
		err := m.Save()
		if err != nil {
			t.Fatal(err)
		}
		found := FindPerson("find me")
		assert.Equal(t, "Find Me", found.PersonName)
	})
	t.Run("nil", func(t *testing.T) {
		r := FindPerson("XXX")
		assert.Nil(t, r)
	})

}

func TestPerson_Links(t *testing.T) {
	t.Run("no-result", func(t *testing.T) {
		m := UnknownPerson
		links := m.Links()
		assert.Empty(t, links)
	})
}

func TestPerson_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewPerson("Update Me", SrcAuto, 0)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		if err := m.Update("PersonName", "Updated Name"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Updated Name", m.PersonName)
		}
	})

}
