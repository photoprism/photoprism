package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_Report(t *testing.T) {
	m := FindSessionByRefID("sessxkkcabcd")

	r, _ := m.Report(false)
	assert.GreaterOrEqual(t, len(r), 1)

	r2, _ := m.Report(true)
	assert.GreaterOrEqual(t, len(r2), 1)
}

func TestSession_Report_Data(t *testing.T) {
	m := &Session{ID: "example", DataJSON: []byte(`{"tokens":["example"]}`)}

	rows, _ := m.Report(false)

	for _, row := range rows {
		assert.NotEqual(t, "DataJSON", row[0])
	}
}

func TestFindSessionByRefID_Stored(t *testing.T) {
	s := NewSession(3600, 0)
	s.SetClientName("find-ref-stored")
	require.NoError(t, s.Save())
	t.Cleanup(func() { _ = s.Delete() })

	found := FindSessionByRefID(s.RefID)
	require.NotNil(t, found)
	assert.True(t, found.stored)

	require.NoError(t, UnscopedDb().Exec("DELETE FROM auth_sessions WHERE id = ?", s.ID).Error)
	assert.ErrorIs(t, found.Save(), ErrSessionNotFound)
	assert.Equal(t, 0, countSessions(t, s.ID))
}
