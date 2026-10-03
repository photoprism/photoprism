package commands

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
)

// setNodeCredentials configures local node credentials for the duration of a test.
func setNodeCredentials(t *testing.T, conf *config.Config, name, id, secret string) {
	t.Helper()

	prevName, prevID, prevSecret := conf.Options().NodeName, conf.Options().NodeClientID, conf.Options().NodeClientSecret

	t.Cleanup(func() {
		conf.Options().NodeName = prevName
		conf.Options().NodeClientID = prevID
		conf.Options().NodeClientSecret = prevSecret
	})

	conf.Options().NodeName, conf.Options().NodeClientID, conf.Options().NodeClientSecret = name, id, secret
}

func TestClusterRegisterToken(t *testing.T) {
	var gotBasic, gotScope string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/oauth/token" {
			http.NotFound(w, r)
			return
		}

		gotBasic = r.Header.Get("Authorization")
		_ = r.ParseForm()
		gotScope = r.PostForm.Get("scope")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "node-access-token", "token_type": "Bearer"})
	}))
	defer ts.Close()

	conf := get.Config()

	t.Run("OwnRegistration", func(t *testing.T) {
		setNodeCredentials(t, conf, "pp-token-node", "cs5cpu17token123", cluster.ExampleClientSecret)

		token, err := clusterRegisterToken(conf, ts.URL, cluster.ExampleJoinToken, "pp-token-node")
		assert.NoError(t, err)

		// A node mutating its own registration is authorized by its own access token.
		assert.Equal(t, "node-access-token", token)
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("cs5cpu17token123:"+cluster.ExampleClientSecret)), gotBasic)
		assert.Equal(t, "cluster", gotScope)
	})
	t.Run("FirstJoin", func(t *testing.T) {
		setNodeCredentials(t, conf, "pp-token-node", "", "")

		// Without node credentials there is nothing to exchange, so the join token applies.
		token, err := clusterRegisterToken(conf, ts.URL, cluster.ExampleJoinToken, "pp-token-node")
		assert.NoError(t, err)
		assert.Equal(t, cluster.ExampleJoinToken, token)
	})
	t.Run("DifferentNode", func(t *testing.T) {
		setNodeCredentials(t, conf, "pp-token-node", "cs5cpu17token123", cluster.ExampleClientSecret)

		// Local credentials do not authorize a registration for another node.
		token, err := clusterRegisterToken(conf, ts.URL, cluster.ExampleJoinToken, "pp-other-node")
		assert.NoError(t, err)
		assert.Equal(t, cluster.ExampleJoinToken, token)
	})
	t.Run("MissingJoinToken", func(t *testing.T) {
		setNodeCredentials(t, conf, "pp-token-node", "", "")

		_, err := clusterRegisterToken(conf, ts.URL, "", "pp-token-node")
		assert.Error(t, err)
	})
	t.Run("InvalidPortalUrl", func(t *testing.T) {
		setNodeCredentials(t, conf, "pp-token-node", "cs5cpu17token123", cluster.ExampleClientSecret)

		_, err := clusterRegisterToken(conf, "not a url", cluster.ExampleJoinToken, "pp-token-node")
		assert.Error(t, err)
	})
	t.Run("StaleCredentialsRejoin", func(t *testing.T) {
		denied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer denied.Close()

		setNodeCredentials(t, conf, "pp-token-node", "cs5cpu17token123", cluster.ExampleClientSecret)

		// Credentials the Portal no longer honors fall back to the join token, so an
		// operator can rejoin once the stale registration is removed.
		token, err := clusterRegisterToken(conf, denied.URL, cluster.ExampleJoinToken, "pp-token-node")
		assert.NoError(t, err)
		assert.Equal(t, cluster.ExampleJoinToken, token)
	})
	t.Run("StaleCredentialsWithoutJoinToken", func(t *testing.T) {
		denied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer denied.Close()

		setNodeCredentials(t, conf, "pp-token-node", "cs5cpu17token123", cluster.ExampleClientSecret)

		_, err := clusterRegisterToken(conf, denied.URL, "", "pp-token-node")

		if assert.Error(t, err) {
			// An unusable credential is an authentication failure, not a usage error.
			assert.Equal(t, 4, clusterTokenExitCode(err))
		}
	})
	t.Run("MissingTokenExitCode", func(t *testing.T) {
		setNodeCredentials(t, conf, "pp-token-node", "", "")

		_, err := clusterRegisterToken(conf, ts.URL, "", "pp-token-node")

		if assert.Error(t, err) {
			assert.ErrorIs(t, err, ErrMissingPortalToken)
			assert.Equal(t, 2, clusterTokenExitCode(err))
		}
	})
}
