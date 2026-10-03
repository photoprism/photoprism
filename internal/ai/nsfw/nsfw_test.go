package nsfw

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultModel verifies the bundled detector and its calibrated fallback threshold.
func TestDefaultModel(t *testing.T) {
	description := FindModel(DefaultModelName())

	require.NotNil(t, description)
	assert.Equal(t, ModelYahoo, description.Name)
	assert.InDelta(t, 0.327, description.DefaultThreshold, 1e-6)
}

// TestRegisteredModelThresholds verifies the corpus-calibrated detector thresholds.
func TestRegisteredModelThresholds(t *testing.T) {
	expected := map[ModelName]float32{
		ModelAdamCoddFP32: 0.718,
		ModelAdamCoddINT8: 0.760,
		ModelFalconsai:    0.529,
		ModelFreepik:      0.992,
		ModelYahoo:        0.327,
	}

	for name, threshold := range expected {
		description := FindModel(name)
		require.NotNil(t, description)
		assert.InDelta(t, threshold, description.DefaultThreshold, 1e-6, name)
	}
}

// TestResultZeroValue verifies that an unfilled result is never read as a clearance.
func TestResultZeroValue(t *testing.T) {
	var result Result

	assert.False(t, result.IsSafe())
	assert.False(t, result.IsUnsafe())
	assert.True(t, result.IsUnavailable())
	assert.False(t, result.HasScores())
}

// TestUnavailable verifies that an undecided result records why.
func TestUnavailable(t *testing.T) {
	result := Unavailable("model is missing")

	assert.True(t, result.IsUnavailable())
	assert.False(t, result.IsSafe())
	assert.Equal(t, "model is missing", result.Reason)
	assert.Equal(t, float32(0), result.Score)
}

// TestNewResult verifies that the threshold comparison is inclusive of the boundary.
func TestNewResult(t *testing.T) {
	t.Run("Unsafe", func(t *testing.T) {
		result := NewResult(0.9, 0.75)
		assert.True(t, result.IsUnsafe())
		assert.False(t, result.IsSafe())
		assert.InDelta(t, 0.75, result.Threshold, 1e-6)
	})
	t.Run("Safe", func(t *testing.T) {
		result := NewResult(0.7, 0.75)
		assert.True(t, result.IsSafe())
		assert.False(t, result.IsUnsafe())
	})
	t.Run("AtThreshold", func(t *testing.T) {
		assert.True(t, NewResult(0.75, 0.75).IsUnsafe())
	})
	t.Run("ZeroScoreIsDecided", func(t *testing.T) {
		result := NewResult(0, 0.75)
		assert.True(t, result.IsSafe())
		assert.False(t, result.IsUnavailable())
	})
	t.Run("NonFiniteIsUnavailable", func(t *testing.T) {
		result := NewResult(float32(math.NaN()), 0.75)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
		assert.NotEmpty(t, result.Reason)
	})
}

// TestResultDecide verifies that class scores reduce to a decision.
func TestResultDecide(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Result{}.Decide(0.75)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
		assert.Equal(t, "no scores", result.Reason)
	})
	t.Run("Unsafe", func(t *testing.T) {
		result := Result{Hentai: 0.98}.Decide(0.75)
		assert.True(t, result.IsUnsafe())
		assert.InDelta(t, 0.98, result.Score, 1e-6)
		assert.InDelta(t, 0.98, result.Hentai, 1e-6)
	})
	t.Run("Safe", func(t *testing.T) {
		result := Result{Neutral: 0.9, Sexy: 0.05}.Decide(0.75)
		assert.True(t, result.IsSafe())
		assert.InDelta(t, 0.05, result.Score, 1e-6)
	})
	t.Run("DrawingIsNotUnsafe", func(t *testing.T) {
		assert.True(t, Result{Drawing: 0.999}.Decide(0.75).IsSafe())
	})
	t.Run("MaxOfUnsafeClasses", func(t *testing.T) {
		result := Result{Hentai: 0.2, Porn: 0.3, Sexy: 0.8}.Decide(0.75)
		assert.True(t, result.IsUnsafe())
		assert.InDelta(t, 0.8, result.Score, 1e-6)
	})
	t.Run("CurrentResultUsesScore", func(t *testing.T) {
		result := NewResult(0.85, 0.8).Decide(0.9)
		assert.True(t, result.IsSafe())
		assert.InDelta(t, 0.85, result.Score, 1e-6)
		assert.InDelta(t, 0.9, result.Threshold, 1e-6)
	})
	t.Run("InvalidCurrentScoreIsUnavailable", func(t *testing.T) {
		result := Result{Status: StatusSafe, Score: 2}.Decide(0.75)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
		assert.NotEmpty(t, result.Reason)
	})
	t.Run("ExplicitUnsafeWithoutScore", func(t *testing.T) {
		var result Result
		require.NoError(t, json.Unmarshal([]byte(`{"status":"unsafe"}`), &result))
		assert.True(t, result.Decide(0.75).IsUnsafe())
	})
	t.Run("ExplicitUnsafeUsesClassScores", func(t *testing.T) {
		var result Result
		require.NoError(t, json.Unmarshal([]byte(`{"status":"unsafe","Porn":0.99}`), &result))
		decided := result.Decide(0.75)
		assert.True(t, decided.IsUnsafe())
		assert.InDelta(t, 0.99, decided.Score, 1e-6)
	})
	t.Run("ExplicitZeroScoreIsReevaluated", func(t *testing.T) {
		var result Result
		require.NoError(t, json.Unmarshal([]byte(`{"status":"unsafe","score":0}`), &result))
		assert.True(t, result.Decide(0.75).IsSafe())
	})
}

// TestResultUnsafeScore verifies that only unsafe classes contribute to the reduced score.
func TestResultUnsafeScore(t *testing.T) {
	result := Result{Drawing: 0.99, Hentai: 0.2, Neutral: 0.98, Porn: 0.7, Sexy: 0.4}
	assert.InDelta(t, 0.7, result.UnsafeScore(), 1e-6)
}

// TestResultHasScores verifies that zero and populated class vectors are distinguishable.
func TestResultHasScores(t *testing.T) {
	assert.False(t, Result{}.HasScores())
	assert.True(t, Result{Neutral: 1}.HasScores())
}

// TestResultDecisionPredicates verifies the mutually exclusive decision helpers.
func TestResultDecisionPredicates(t *testing.T) {
	t.Run("Safe", func(t *testing.T) {
		result := Result{Status: StatusSafe}
		assert.True(t, result.IsSafe())
		assert.False(t, result.IsUnsafe())
		assert.False(t, result.IsUnavailable())
	})
	t.Run("Unsafe", func(t *testing.T) {
		result := Result{Status: StatusUnsafe}
		assert.False(t, result.IsSafe())
		assert.True(t, result.IsUnsafe())
		assert.False(t, result.IsUnavailable())
	})
	t.Run("Unavailable", func(t *testing.T) {
		result := Result{Status: StatusUnavailable}
		assert.False(t, result.IsSafe())
		assert.False(t, result.IsUnsafe())
		assert.True(t, result.IsUnavailable())
	})
}

// TestResultDecideNoNeutralVeto verifies that neutral scores do not suppress unsafe scores.
func TestResultDecideNoNeutralVeto(t *testing.T) {
	assert.True(t, Result{Neutral: 0.30, Porn: 0.80}.Decide(0.75).IsUnsafe())
	assert.True(t, Result{Neutral: 0.60, Porn: 0.40}.Decide(0.35).IsUnsafe())
}

// TestStatusString verifies that the zero value renders as "unavailable".
func TestStatusString(t *testing.T) {
	assert.Equal(t, "unavailable", StatusUnavailable.String())
	assert.Equal(t, "safe", StatusSafe.String())
	assert.Equal(t, "unsafe", StatusUnsafe.String())
}

// TestStatusJSON verifies that a status round-trips and that unknown names stay undecided.
func TestStatusJSON(t *testing.T) {
	t.Run("MarshalUnavailable", func(t *testing.T) {
		b, err := json.Marshal(StatusUnavailable)
		require.NoError(t, err)
		assert.JSONEq(t, `"unavailable"`, string(b))
	})
	t.Run("RoundTrip", func(t *testing.T) {
		for _, status := range []Status{StatusSafe, StatusUnsafe, StatusUnavailable} {
			b, err := json.Marshal(status)
			require.NoError(t, err)
			var parsed Status
			require.NoError(t, json.Unmarshal(b, &parsed))
			assert.Equal(t, status, parsed)
		}
	})
	t.Run("UnknownStaysUndecided", func(t *testing.T) {
		var parsed Status
		require.NoError(t, json.Unmarshal([]byte(`"probably-fine"`), &parsed))
		assert.Equal(t, StatusUnavailable, parsed)
	})
}

// TestResultLegacyWire verifies the established capitalized class fields remain stable.
func TestResultLegacyWire(t *testing.T) {
	b, err := json.Marshal(NewResult(0.9, 0.75))
	require.NoError(t, err)

	var fields map[string]any
	require.NoError(t, json.Unmarshal(b, &fields))

	for _, name := range []string{"Drawing", "Hentai", "Neutral", "Porn", "Sexy"} {
		assert.Contains(t, fields, name)
	}

	assert.Equal(t, "unsafe", fields["status"])
	assert.Contains(t, fields, "score")
	assert.Contains(t, fields, "threshold")
}

// TestUnmarshalLegacyResponse verifies that a response from a service older than Status carries
// no decision but can still be decided from its class scores.
func TestUnmarshalLegacyResponse(t *testing.T) {
	var result Result
	require.NoError(t, json.Unmarshal([]byte(`{"Drawing":0,"Hentai":0.98,"Neutral":0,"Porn":0,"Sexy":0}`), &result))

	assert.True(t, result.IsUnavailable())
	assert.True(t, result.HasScores())
	assert.True(t, result.Decide(0.75).IsUnsafe())
}

// TestResult_UnmarshalJSON verifies score presence remains distinct from an omitted score.
func TestResult_UnmarshalJSON(t *testing.T) {
	var withScore Result
	require.NoError(t, json.Unmarshal([]byte(`{"status":"unsafe","score":0}`), &withScore))
	assert.True(t, withScore.scoreSet)
	assert.True(t, withScore.Decide(0.75).IsSafe())

	var withoutScore Result
	require.NoError(t, json.Unmarshal([]byte(`{"status":"unsafe"}`), &withoutScore))
	assert.False(t, withoutScore.scoreSet)
	assert.True(t, withoutScore.Decide(0.75).IsUnsafe())

	var nullScore Result
	require.NoError(t, json.Unmarshal([]byte(`{"status":"unsafe","score":null,"threshold":0.5}`), &nullScore))
	assert.False(t, nullScore.scoreSet)
	assert.True(t, nullScore.Decide(0.75).IsUnsafe())

	var nullScoreWithClasses Result
	require.NoError(t, json.Unmarshal([]byte(`{"status":"safe","score":null,"threshold":0.5,"Porn":0.9}`), &nullScoreWithClasses))
	assert.False(t, nullScoreWithClasses.scoreSet)
	assert.True(t, nullScoreWithClasses.Decide(0.75).IsUnsafe())
}

// TestValidateScore verifies that only finite probabilities from 0 to 1 are accepted.
func TestValidateScore(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		require.NoError(t, ValidateScore(0))
		require.NoError(t, ValidateScore(0.5))
		require.NoError(t, ValidateScore(1))
	})
	t.Run("NaN", func(t *testing.T) {
		require.Error(t, ValidateScore(float32(math.NaN())))
	})
	t.Run("Inf", func(t *testing.T) {
		require.Error(t, ValidateScore(float32(math.Inf(1))))
	})
	t.Run("OutOfRange", func(t *testing.T) {
		require.Error(t, ValidateScore(-0.1))
		require.Error(t, ValidateScore(1.1))
	})
}
