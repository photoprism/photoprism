package commands

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Verifies OAuth path in cluster theme pull using client_id/client_secret.
func TestClusterThemePull_OAuth(t *testing.T) {
	// Build an in-memory zip with one file
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, _ := zw.Create("ok.txt")
	_, _ = f.Write([]byte("ok\n"))
	_ = zw.Close()

	// Fake portal server
	var gotBasic string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth/token":
			// Expect Basic auth for nodeid:secret
			gotBasic = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "token_type": "Bearer", "scope": "cluster vision"})
		case "/api/v1/cluster/theme":
			if r.Header.Get("Authorization") != "Bearer tok" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipBuf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// Prepare destination
	dest := t.TempDir()
	// Run CLI with OAuth creds
	out, err := RunWithTestContext(ClusterThemePullCommand.Subcommands[0], []string{
		"pull", "--dest", dest, "-f",
		"--portal-url=" + ts.URL,
		"--client-id=nodeid",
		"--client-secret=secret",
	})
	_ = out
	assert.NoError(t, err)
	// Verify file extracted
	assert.FileExists(t, filepath.Join(dest, "ok.txt"))
	// Verify Basic header format
	expect := "Basic " + base64.StdEncoding.EncodeToString([]byte("nodeid:secret"))
	assert.Equal(t, expect, gotBasic)
}

// Verifies that when only a join token is provided, the command obtains
// client credentials via the register endpoint, then uses OAuth to pull the theme.
func TestClusterThemePull_JoinTokenToOAuth(t *testing.T) {
	// Zip fixture
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	_, _ = zw.Create("ok2.txt")
	_ = zw.Close()

	// Fake portal server responds with register then token then theme
	var sawRotateSecret bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			// Must have Bearer join token
			if r.Header.Get("Authorization") != "Bearer jt" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// Read body to check rotateSecret flag
			var req struct {
				RotateSecret bool   `json:"rotateSecret"`
				NodeName     string `json:"nodeName"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			sawRotateSecret = req.RotateSecret
			w.Header().Set("Content-Type", "application/json")
			// Return NodeClientID and a fresh secret
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{
				UUID:    rnd.UUID(),
				Node:    cluster.Node{ClientID: "cs5gfen1bgxz7s9i", Name: "pp-node-01"},
				Secrets: &cluster.RegisterSecrets{ClientSecret: "s3cr3t"},
			})
		case "/api/v1/oauth/token":
			// Expect Basic for the returned creds
			if r.Header.Get("Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("cs5gfen1bgxz7s9i:s3cr3t")) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok2", "token_type": "Bearer"})
		case "/api/v1/cluster/theme":
			if r.Header.Get("Authorization") != "Bearer tok2" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipBuf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dest := t.TempDir()
	out, err := RunWithTestContext(ClusterThemePullCommand.Subcommands[0], []string{
		"pull", "--dest", dest, "-f",
		"--portal-url=" + ts.URL,
		"--join-token=jt",
	})
	_ = out
	assert.NoError(t, err)
	assert.True(t, sawRotateSecret)
}
