package photoprism

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestFaces_LogsMarkerSubjects pins that a run reports names it resolved to people and markers it
// linked to existing people on separate lines, and counts both as subjects.
func TestFaces_LogsMarkerSubjects(t *testing.T) {
	w := isolatedTestFaces(t, "faceslogsubjects")

	// Settles what the fixtures leave, so the counts below are this test's alone.
	_, _, err := query.CreateMarkerSubjects()
	require.NoError(t, err)

	newMarker := func(t *testing.T, name, src string) {
		t.Helper()

		m := entity.Marker{
			MarkerUID:  rnd.GenerateUID('m'),
			FileUID:    consensusTestFileUID,
			MarkerType: entity.MarkerFace,
			MarkerName: name,
			SubjSrc:    src,
			W:          0.1,
			H:          0.1,
		}

		require.NoError(t, entity.UnscopedDb().Create(&m).Error)
	}

	run := func(t *testing.T) (facesRunResult, string) {
		t.Helper()

		hook := captureLog(t)
		result, err := w.start(FacesOptions{})
		require.NoError(t, err)

		return result, strings.Join(loggedMessages(hook, logrus.InfoLevel), "\n")
	}

	t.Run("BothPaths", func(t *testing.T) {
		bob := consensusTestSubject(t, "Log Subjects Bob")
		dora := consensusTestSubject(t, "Log Subjects Dora")
		newMarker(t, bob.SubjName, entity.SrcXmp)
		newMarker(t, dora.SubjName, entity.SrcXmp)
		newMarker(t, "Log Subjects Carl", entity.SrcManual)

		result, logged := run(t)
		assert.Contains(t, logged, "markers: resolved 1 name to a person [")
		assert.Contains(t, logged, "markers: linked 2 markers to existing people [")
		assert.Equal(t, 3, result.Subjects)
	})
	t.Run("XmpOnly", func(t *testing.T) {
		erin := consensusTestSubject(t, "Log Subjects Erin")
		newMarker(t, erin.SubjName, entity.SrcXmp)

		result, logged := run(t)
		assert.Contains(t, logged, "markers: linked 1 marker to existing people [")
		assert.NotContains(t, logged, "markers: resolved")
		assert.Equal(t, 1, result.Subjects)
		assert.True(t, result.Moved())
	})
	t.Run("ManualOnly", func(t *testing.T) {
		newMarker(t, "Log Subjects Finn", entity.SrcManual)

		result, logged := run(t)
		assert.Contains(t, logged, "markers: resolved 1 name to a person [")
		assert.NotContains(t, logged, "markers: linked")
		assert.Equal(t, 1, result.Subjects)
	})
	t.Run("Nothing", func(t *testing.T) {
		result, logged := run(t)
		assert.NotContains(t, logged, "markers: resolved")
		assert.NotContains(t, logged, "markers: linked")
		assert.Zero(t, result.Subjects)
	})
}
