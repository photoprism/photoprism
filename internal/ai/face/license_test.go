package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingModel_LicenseGated(t *testing.T) {
	t.Run("ResearchOnly", func(t *testing.T) {
		assert.True(t, FindEmbeddingModel(ModelArcFaceR50).LicenseGated())
		assert.True(t, FindEmbeddingModel(ModelArcFaceMBF).LicenseGated())
	})
	t.Run("Permissive", func(t *testing.T) {
		assert.False(t, FindEmbeddingModel(ModelSFace).LicenseGated())
		assert.False(t, FindEmbeddingModel(ModelAuraFace).LicenseGated())
	})
	t.Run("UnknownLicense", func(t *testing.T) {
		// FaceNet's weight provenance was never verified, which is not the same as gated.
		assert.False(t, FindEmbeddingModel(ModelFaceNet).LicenseGated())
	})
	t.Run("NilModel", func(t *testing.T) {
		assert.False(t, (*EmbeddingModel)(nil).LicenseGated())
	})
}

func TestLicenseAccepted(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		t.Setenv(LicenseAcceptanceVar, "")
		assert.False(t, LicenseAccepted())
	})
	t.Run("Accepted", func(t *testing.T) {
		t.Setenv(LicenseAcceptanceVar, "1")
		assert.True(t, LicenseAccepted())
	})
	t.Run("Refused", func(t *testing.T) {
		t.Setenv(LicenseAcceptanceVar, "0")
		assert.False(t, LicenseAccepted())
	})
}

func TestLicenseEligibleEdition(t *testing.T) {
	t.Run("Eligible", func(t *testing.T) {
		assert.True(t, LicenseEligibleEdition("ce"))
		assert.True(t, LicenseEligibleEdition(" Plus "))
	})
	t.Run("Ineligible", func(t *testing.T) {
		assert.False(t, LicenseEligibleEdition("pro"))
		assert.False(t, LicenseEligibleEdition("portal"))
	})
	t.Run("Unknown", func(t *testing.T) {
		// An edition this predicate does not know about must not enable gated weights.
		assert.False(t, LicenseEligibleEdition(""))
		assert.False(t, LicenseEligibleEdition("enterprise"))
	})
}

func TestLicenseRefused(t *testing.T) {
	t.Run("Ungated", func(t *testing.T) {
		assert.NoError(t, LicenseRefused(ModelSFace, "pro"))
		assert.NoError(t, LicenseRefused(ModelFaceNet, "portal"))
	})
	t.Run("UnknownModel", func(t *testing.T) {
		assert.NoError(t, LicenseRefused("dlib", "ce"))
	})
	t.Run("NotAccepted", func(t *testing.T) {
		t.Setenv(LicenseAcceptanceVar, "")
		err := LicenseRefused(ModelArcFaceR50, "ce")

		require.Error(t, err)
		assert.Contains(t, err.Error(), LicenseAcceptanceVar)
	})
	t.Run("AcceptedAndEligible", func(t *testing.T) {
		t.Setenv(LicenseAcceptanceVar, "1")
		assert.NoError(t, LicenseRefused(ModelArcFaceR50, "ce"))
		assert.NoError(t, LicenseRefused(ModelArcFaceMBF, "plus"))
	})
	t.Run("AcceptedButIneligible", func(t *testing.T) {
		t.Setenv(LicenseAcceptanceVar, "1")
		err := LicenseRefused(ModelArcFaceR50, "pro")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "pro edition")
	})
}
