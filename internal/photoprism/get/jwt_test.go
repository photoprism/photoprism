package get

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

func TestJWTVerifierReuse(t *testing.T) {
	verifier1 := JWTVerifier()
	require.NotNil(t, verifier1)

	verifier2 := JWTVerifier()
	require.NotNil(t, verifier2)

	assert.Same(t, verifier1, verifier2)
}

func TestJWTVerifierResetOnConfigChange(t *testing.T) {
	orig := Config()
	verifier1 := JWTVerifier()
	require.NotNil(t, verifier1)

	tempConf := config.NewMinimalTestConfigWithDbTTest("jwt-reset", t.TempDir(), t)
	SetConfig(tempConf)
	t.Cleanup(func() {
		SetConfig(orig)
		orig.RegisterDb()
	})

	verifier2 := JWTVerifier()
	require.NotNil(t, verifier2)

	assert.NotSame(t, verifier1, verifier2)
}
