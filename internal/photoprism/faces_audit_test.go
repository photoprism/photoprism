package photoprism

import (
	"crypto/sha1" //nolint:gosec // SHA1 retained for legacy audit IDs.
	"encoding/base32"
	"math"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

func TestFaces_Audit(t *testing.T) {
	t.Run("FixEqualTrue", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		err := m.Audit(true, "")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("FixeEqualFalse", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		err := m.Audit(false, "")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("SubjectFilter", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		require.NoError(t, m.Audit(false, "jr0ncy131y7igds8"))
	})
}

func TestFaces_AuditNormalizesEmbeddings(t *testing.T) {
	c := config.NewMinimalTestConfigWithDbTTest("faces-audit-normalize", t.TempDir(), t)

	m := NewFaces(c)

	raw := make(face.Embedding, len(face.NullEmbedding))
	raw[0] = 2
	raw[1] = 1

	rawJSON := raw.JSON()

	//nolint:gosec // legacy ID compatibility relies on SHA1.
	original := sha1.Sum(rawJSON)
	oldID := base32.StdEncoding.EncodeToString(original[:])

	now := entity.Now()

	faceRow := &entity.Face{
		ID:            oldID,
		FaceSrc:       entity.SrcAuto,
		EmbeddingJSON: rawJSON,
		Samples:       5,
		SampleRadius:  0.12,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, entity.Db().Create(faceRow).Error)

	markerEmbJSON := (face.Embeddings{raw}).JSON()

	marker := &entity.Marker{
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcAuto,
		FaceID:         oldID,
		EmbeddingsJSON: markerEmbJSON,
		FaceDist:       0.5,
	}

	require.NoError(t, entity.Db().Create(marker).Error)

	//nolint:gosec // legacy ID compatibility relies on SHA1.
	hashNorm := sha1.Sum(normalizeEmbeddingCopy(raw).JSON())
	expectedID := base32.StdEncoding.EncodeToString(hashNorm[:])

	require.NoError(t, m.Audit(true, ""))

	var updated entity.Face
	require.NoError(t, entity.Db().Where("id = ?", expectedID).First(&updated).Error)
	require.NotEqual(t, oldID, updated.ID)

	updatedEmbedding := updated.Embedding()
	require.InDelta(t, 1.0, updatedEmbedding.Magnitude(), 1e-9)
	normalized := normalizeEmbeddingCopy(raw)
	require.InDelta(t, normalized[0], updatedEmbedding[0], 1e-9)

	var updatedMarker entity.Marker
	require.NoError(t, entity.Db().Where("marker_uid = ?", marker.MarkerUID).First(&updatedMarker).Error)

	expectedDist := minEmbeddingDistance(normalized, updatedMarker.Embeddings())
	require.InDelta(t, expectedDist, updatedMarker.FaceDist, 1e-9)
	require.Equal(t, expectedID, updatedMarker.FaceID)
}

func normalizeEmbeddingCopy(src face.Embedding) face.Embedding {
	copyEmb := make(face.Embedding, len(src))
	copy(copyEmb, src)

	var sum float64

	for _, v := range copyEmb {
		sum += v * v
	}

	length := math.Sqrt(sum)

	if length == 0 {
		return copyEmb
	}

	inv := 1 / length

	for i := range copyEmb {
		copyEmb[i] *= inv
	}

	return copyEmb
}

// captureLog records what the shared logger emits for the duration of a test, so the
// audit functions can be checked on what they report rather than only on not panicking.
func captureLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)

	hook := test.NewLocal(logger)
	t.Cleanup(hook.Reset)

	return hook
}

// loggedMessages returns the recorded messages at or above the given level.
func loggedMessages(hook *test.Hook, level logrus.Level) []string {
	var result []string

	for _, entry := range hook.AllEntries() {
		if entry.Level <= level {
			result = append(result, entry.Message)
		}
	}

	return result
}

func TestFaces_auditEmbeddingModels(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Cleanup(func() {
		restoreEmbedder(t)
	})

	t.Run("ConfiguredModel", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
			Name:  face.ModelFaceNet,
			Model: face.FindEmbeddingModel(face.ModelFaceNet),
		}))
		hook := captureLog(t)
		w.auditEmbeddingModels()

		assert.NotEmpty(t, hook.AllEntries(), "the audit reports what the library holds")
	})
	t.Run("StaleClusters", func(t *testing.T) {
		// A cluster from another model must be reported as stale.
		m := entity.NewFace("", entity.SrcAuto, face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())
		require.NotNil(t, m)
		m.EmbedModel = face.ModelArcFaceR50
		require.NoError(t, entity.Db().Create(m).Error)

		t.Cleanup(func() { entity.Db().Delete(m) })

		count, err := query.FacesFromOtherModels()
		require.NoError(t, err)
		require.Positive(t, count)

		hook := captureLog(t)
		w.auditEmbeddingModels()

		warnings := loggedMessages(hook, logrus.WarnLevel)
		require.NotEmpty(t, warnings, "a cluster from another model is a warning")
		assert.Contains(t, strings.Join(warnings, "\n"), face.ModelArcFaceR50)
		assert.Contains(t, strings.Join(warnings, "\n"), "faces migrate",
			"the operator is told how to resolve it")
	})
	t.Run("NoModelConfigured", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))

		hook := captureLog(t)
		w.auditEmbeddingModels()

		// Without a configured model nothing can be called incompatible with it, so the
		// counts are reported without telling the operator to migrate.
		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "faces migrate")
	})
}

func TestFaces_auditMarkerEmbeddingModels(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Run("StaleMarkers", func(t *testing.T) {
		m := &entity.Marker{
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			EmbedModel:     face.ModelArcFaceR50,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		}

		require.NoError(t, entity.Db().Create(m).Error)

		t.Cleanup(func() { entity.Db().Delete(m) })

		counts, err := query.MarkerEmbeddingModels()
		require.NoError(t, err)
		require.NotEmpty(t, counts)

		hook := captureLog(t)
		w.auditMarkerEmbeddingModels(face.ModelFaceNet)

		warnings := strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n")
		assert.Contains(t, warnings, face.ModelArcFaceR50)
		assert.Contains(t, warnings, face.ModelFaceNet, "the configured model is named alongside it")
	})
	t.Run("NoModelConfigured", func(t *testing.T) {
		hook := captureLog(t)
		w.auditMarkerEmbeddingModels("")

		// A blank stored model is legacy FaceNet, which nothing contradicts when no model
		// is configured, so those markers are counted rather than flagged.
		assert.NotContains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "not compatible")
	})
}
