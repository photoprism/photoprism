package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestSubject_UpdateName_FormerName pins that a renamed person no longer answers to the former name.
func TestSubject_UpdateName_FormerName(t *testing.T) {
	suffix := rnd.Base36(6)
	former := "Former Anna " + suffix
	current := "Former Anne " + suffix

	m := NewSubject(former, SubjPerson, SrcManual)
	require.NoError(t, m.Save())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", m.SubjUID) })
	require.NotNil(t, FindSubjectByName(former, false))

	_, err := m.UpdateName(current)
	require.NoError(t, err)

	t.Run("FormerName", func(t *testing.T) {
		assert.Nil(t, FindSubjectByName(former, false))
		assert.Empty(t, SubjNames.Key(former))
	})
	t.Run("CurrentName", func(t *testing.T) {
		if found := FindSubjectByName(current, false); assert.NotNil(t, found) {
			assert.Equal(t, m.SubjUID, found.SubjUID)
		}
	})
	t.Run("RenameOtherIntoFormerName", func(t *testing.T) {
		other := NewSubject("Former Bella "+suffix, SubjPerson, SrcManual)
		require.NoError(t, other.Save())
		t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", other.SubjUID) })

		assert.Nil(t, ReassignSubject(other, former))

		renamed, err := other.UpdateName(former)
		require.NoError(t, err)
		assert.Equal(t, other.SubjUID, renamed.SubjUID, "the name is taken by nobody, so it renames")
		assert.False(t, FindSubject(other.SubjUID).Deleted())
		assert.Equal(t, current, FindSubject(m.SubjUID).SubjName)

		_, err = other.UpdateName("Former Bella " + suffix)
		require.NoError(t, err)
	})
	t.Run("OtherPersonWithFormerName", func(t *testing.T) {
		other := NewSubject(former, SubjPerson, SrcManual)
		require.NoError(t, other.Save())
		t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", other.SubjUID) })

		if found := FindSubjectByName(former, false); assert.NotNil(t, found) {
			assert.Equal(t, other.SubjUID, found.SubjUID)
		}
	})
}

// TestSetSubjName pins that a changed name is retracted from the reverse lookup of the previous one,
// and only for that uid.
func TestSetSubjName(t *testing.T) {
	newUID := func(t *testing.T) string {
		uid := rnd.GenerateUID('j')
		t.Cleanup(func() { SubjNames.Unset(uid) })
		return uid
	}

	t.Run("Rename", func(t *testing.T) {
		uid := newUID(t)
		setSubjName(uid, "Set Name Anna")
		setSubjName(uid, "Set Name Anne")
		assert.Equal(t, "Set Name Anne", SubjNames.Get(uid))
		assert.Equal(t, uid, SubjNames.Key("Set Name Anne"))
		assert.Empty(t, SubjNames.Key("Set Name Anna"))
	})
	t.Run("Unchanged", func(t *testing.T) {
		uid := newUID(t)
		setSubjName(uid, "Set Name Carl")
		setSubjName(uid, "Set Name Carl")
		assert.Equal(t, []string{uid}, SubjNames.Keys("Set Name Carl"))
	})
	t.Run("CaseOnly", func(t *testing.T) {
		uid := newUID(t)
		setSubjName(uid, "set name dora")
		setSubjName(uid, "Set Name Dora")
		assert.Equal(t, "Set Name Dora", SubjNames.Get(uid))
		assert.Equal(t, []string{uid}, SubjNames.Keys("set name dora"))
	})
	t.Run("SharedName", func(t *testing.T) {
		renamed, other := newUID(t), newUID(t)
		setSubjName(other, "set name emma")
		setSubjName(renamed, "Set Name Emma")
		setSubjName(renamed, "Set Name Emily")
		assert.Equal(t, other, SubjNames.Key("Set Name Emma"), "a person with the same name still resolves")
		assert.Equal(t, renamed, SubjNames.Key("Set Name Emily"))
	})
	t.Run("EmptyName", func(t *testing.T) {
		uid := newUID(t)
		setSubjName(uid, "Set Name Finn")
		setSubjName(uid, "")
		assert.False(t, SubjNames.Has(uid))
		assert.Empty(t, SubjNames.Key("Set Name Finn"))
	})
	t.Run("EmptyUID", func(t *testing.T) {
		setSubjName("", "Set Name Gina")
		assert.Empty(t, SubjNames.Key("Set Name Gina"))
	})
}

// TestSubject_AfterFind_FormerName pins that loading a person replaces a former name a stale lookup
// still holds, as when another process renamed them.
func TestSubject_AfterFind_FormerName(t *testing.T) {
	suffix := rnd.Base36(6)
	former := "Stale Anna " + suffix

	m := NewSubject("Stale Anne "+suffix, SubjPerson, SrcManual)
	require.NoError(t, m.Save())
	t.Cleanup(func() {
		UnscopedDb().Delete(Subject{}, "subj_uid = ?", m.SubjUID)
		SubjNames.Unset(m.SubjUID)
	})

	t.Run("FindSubject", func(t *testing.T) {
		SubjNames.Set(m.SubjUID, former)
		require.NotNil(t, FindSubject(m.SubjUID))
		assert.Empty(t, SubjNames.Key(former))
		assert.Equal(t, m.SubjUID, SubjNames.Key(m.SubjName))
	})
	t.Run("FindSubjectByName", func(t *testing.T) {
		SubjNames.Set(m.SubjUID, former)
		assert.Nil(t, FindSubjectByName(former, false), "a stale entry does not resolve the former name")

		if found := FindSubjectByName(m.SubjName, false); assert.NotNil(t, found) {
			assert.Equal(t, m.SubjUID, found.SubjUID)
		}
	})
	t.Run("Person", func(t *testing.T) {
		SubjNames.Set(m.SubjUID, former)
		p := Person{SubjUID: m.SubjUID, SubjName: m.SubjName}
		require.NoError(t, p.AfterFind())
		assert.Empty(t, SubjNames.Key(former))
	})
}
