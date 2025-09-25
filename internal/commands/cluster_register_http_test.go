package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestClusterRegister_HTTPHappyPath(t *testing.T) {
	// Fake Portal register endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n1", "name": "pp-node-02", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwd", "dsn": "user:pwd@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			"secrets":            map[string]any{"clientSecret": "secret", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  false,
			"alreadyProvisioned": false,
		})
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-02", "--role", "instance", "--portal-url", ts.URL, "--join-token", "test-token", "--json",
	})
	assert.NoError(t, err)
	// Parse JSON
	assert.Equal(t, "pp-node-02", gjson.Get(out, "node.name").String())
	assert.Equal(t, "secret", gjson.Get(out, "secrets.clientSecret").String())
	assert.Equal(t, "pwd", gjson.Get(out, "database.password").String())
	dsn := gjson.Get(out, "database.dsn").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "user", parsed.User)
	assert.Equal(t, "pwd", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
}

func TestClusterNodesRotate_HTTPHappyPath(t *testing.T) {
	// Fake Portal register endpoint for rotation
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n1", "name": "pp-node-03", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwd2", "dsn": "user:pwd2@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			"secrets":            map[string]any{"clientSecret": "secret2", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", "test-token")
	_ = os.Setenv("PHOTOPRISM_CLI", "noninteractive")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	defer os.Unsetenv("PHOTOPRISM_CLI")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--portal-url=" + ts.URL, "--join-token=test-token", "--db", "--secret", "--yes", "pp-node-03",
	})
	assert.NoError(t, err)
	assert.Contains(t, out, "pp-node-03")
	assert.Contains(t, out, "Node Client Secret")
	assert.Contains(t, out, "DB Password")
}

func TestClusterNodesRotate_HTTPJson(t *testing.T) {
	// Fake Portal register endpoint for rotation in JSON mode
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n2", "name": "pp-node-04", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwd3", "dsn": "user:pwd3@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			"secrets":            map[string]any{"clientSecret": "secret3", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", "test-token")
	_ = os.Setenv("PHOTOPRISM_CLI", "noninteractive")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	defer os.Unsetenv("PHOTOPRISM_CLI")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--db", "--secret", "--yes", "pp-node-04",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-04", gjson.Get(out, "node.name").String())
	assert.Equal(t, "secret3", gjson.Get(out, "secrets.clientSecret").String())
	assert.Equal(t, "pwd3", gjson.Get(out, "database.password").String())
	dsn := gjson.Get(out, "database.dsn").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "user", parsed.User)
	assert.Equal(t, "pwd3", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
}

func TestClusterNodesRotate_DBOnly_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Read payload to assert rotate flags
		b, _ := io.ReadAll(r.Body)
		rotate := gjson.GetBytes(b, "rotate").Bool()
		rotateSecret := gjson.GetBytes(b, "rotateSecret").Bool()
		// Expect DB rotation only
		if !rotate || rotateSecret {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":     map[string]any{"id": "n3", "name": "pp-node-05", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database": map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwd4", "dsn": "pp_user:pwd4@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			// secrets omitted on DB-only rotate
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", "test-token")
	_ = os.Setenv("PHOTOPRISM_YES", "true")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	defer os.Unsetenv("PHOTOPRISM_YES")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--db", "--yes", "pp-node-05",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-05", gjson.Get(out, "node.name").String())
	assert.Equal(t, "pwd4", gjson.Get(out, "database.password").String())
	dsn := gjson.Get(out, "database.dsn").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "pp_user", parsed.User)
	assert.Equal(t, "pwd4", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
	assert.Equal(t, "", gjson.Get(out, "secrets.clientSecret").String())
}

func TestClusterNodesRotate_SecretOnly_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		rotate := gjson.GetBytes(b, "rotate").Bool()
		rotateSecret := gjson.GetBytes(b, "rotateSecret").Bool()
		// Expect secret-only rotation
		if rotate || !rotateSecret {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n4", "name": "pp-node-06", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "rotatedAt": "2025-09-15T00:00:00Z"},
			"secrets":            map[string]any{"clientSecret": "secret4", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", "test-token")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--secret", "--yes", "pp-node-06",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-06", gjson.Get(out, "node.name").String())
	assert.Equal(t, "secret4", gjson.Get(out, "secrets.clientSecret").String())
	assert.Equal(t, "", gjson.Get(out, "database.password").String())
}

func TestClusterRegister_HTTPUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-unauth", "--role", "instance", "--portal-url", ts.URL, "--join-token", "wrong", "--json",
	})
	if ec, ok := err.(cli.ExitCoder); ok {
		assert.Equal(t, 4, ec.ExitCode())
	} else {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
}

func TestClusterRegister_HTTPConflict(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer ts.Close()

	_, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-conflict", "--role", "instance", "--portal-url", ts.URL, "--join-token", "test-token", "--json",
	})
	if ec, ok := err.(cli.ExitCoder); ok {
		assert.Equal(t, 5, ec.ExitCode())
	} else {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
}

func TestClusterRegister_DryRun_JSON(t *testing.T) {
	// No server needed; dry-run avoids HTTP
	get.Config().Options().PortalUrl = cfg.DefaultPortalUrl
	get.Config().Options().ClusterDomain = "cluster.dev"
	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--dry-run", "--json",
	})
	// Should not fail; output must include portalUrl and payload
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.NotEmpty(t, gjson.Get(out, "portalUrl").String())
	assert.Equal(t, "instance", gjson.Get(out, "payload.nodeRole").String())
	// nodeName may be derived; ensure non-empty
	assert.NotEmpty(t, gjson.Get(out, "payload.nodeName").String())
}

func TestClusterRegister_DryRun_Text(t *testing.T) {
	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--dry-run",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Contains(t, out, "Portal URL:")
	assert.Contains(t, out, "Node Name:")
}

func TestClusterRegister_HTTPBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	_, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp node invalid", "--role", "instance", "--portal-url", ts.URL, "--join-token", "test-token", "--json",
	})
	if ec, ok := err.(cli.ExitCoder); ok {
		assert.Equal(t, 2, ec.ExitCode())
	} else {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
}

func TestClusterRegister_HTTPRateLimitOnceThenOK(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n7", "name": "pp-node-rl", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwdrl", "dsn": "pp_user:pwdrl@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-rl", "--role", "instance", "--portal-url", ts.URL, "--join-token", "test-token", "--rotate", "--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-rl", gjson.Get(out, "node.name").String())
}

func TestClusterNodesRotate_HTTPUnauthorized_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=wrong", "--db", "--yes", "pp-node-x",
	})
	if ec, ok := err.(cli.ExitCoder); ok {
		assert.Equal(t, 4, ec.ExitCode())
	} else {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
}

func TestClusterNodesRotate_HTTPConflict_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer ts.Close()

	_, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=test-token", "--db", "--yes", "pp-node-x",
	})
	if ec, ok := err.(cli.ExitCoder); ok {
		assert.Equal(t, 5, ec.ExitCode())
	} else {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
}

func TestClusterNodesRotate_HTTPBadRequest_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	_, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=test-token", "--db", "--yes", "pp node invalid",
	})
	if ec, ok := err.(cli.ExitCoder); ok {
		assert.Equal(t, 2, ec.ExitCode())
	} else {
		t.Fatalf("expected ExitCoder, got %T", err)
	}
}

func TestClusterNodesRotate_HTTPRateLimitOnceThenOK_JSON(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n8", "name": "pp-node-rl2", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwdrl2", "dsn": "pp_user:pwdrl2@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=test-token", "--db", "--yes", "pp-node-rl2",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-rl2", gjson.Get(out, "node.name").String())
}

func TestClusterRegister_RotateDatabase_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if !gjson.GetBytes(b, "rotate").Bool() || gjson.GetBytes(b, "rotateSecret").Bool() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n5", "name": "pp-node-07", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "password": "pwd7", "dsn": "pp_user:pwd7@tcp(db:3306)/pp_db?parseTime=true", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-07", "--role", "instance", "--portal-url", ts.URL, "--join-token", "test-token", "--rotate", "--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-07", gjson.Get(out, "node.name").String())
	assert.Equal(t, "pwd7", gjson.Get(out, "database.password").String())
	dsn := gjson.Get(out, "database.dsn").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "pp_user", parsed.User)
	assert.Equal(t, "pwd7", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
}

func TestClusterRegister_RotateSecret_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if gjson.GetBytes(b, "rotate").Bool() || !gjson.GetBytes(b, "rotateSecret").Bool() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":               map[string]any{"id": "n6", "name": "pp-node-08", "role": "instance", "createdAt": "2025-09-15T00:00:00Z", "updatedAt": "2025-09-15T00:00:00Z"},
			"database":           map[string]any{"host": "database", "port": 3306, "name": "pp_db", "user": "pp_user", "rotatedAt": "2025-09-15T00:00:00Z"},
			"secrets":            map[string]any{"clientSecret": "pwd8secret", "rotatedAt": "2025-09-15T00:00:00Z"},
			"alreadyRegistered":  true,
			"alreadyProvisioned": true,
		})
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-08", "--role", "instance", "--portal-url", ts.URL, "--join-token", "test-token", "--rotate-secret", "--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-08", gjson.Get(out, "node.name").String())
	assert.Equal(t, "pwd8secret", gjson.Get(out, "secrets.clientSecret").String())
	assert.Equal(t, "", gjson.Get(out, "database.password").String())
}
