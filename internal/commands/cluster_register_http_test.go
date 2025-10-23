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
	"github.com/photoprism/photoprism/internal/service/cluster"
)

func TestClusterRegister_HTTPHappyPath(t *testing.T) {
	// Fake Portal register endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n1",
				Name:      "pp-node-02",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwd",
				DSN:       "user:pwd@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			Secrets: &cluster.RegisterSecrets{
				ClientSecret: cluster.ExampleClientSecret,
				RotatedAt:    "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  false,
			AlreadyProvisioned: false,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-02", "--role", "instance", "--portal-url", ts.URL, "--join-token", cluster.ExampleJoinToken, "--json",
	})
	assert.NoError(t, err)
	// Parse JSON
	assert.Equal(t, "pp-node-02", gjson.Get(out, "Node.Name").String())
	assert.Equal(t, cluster.ExampleClientSecret, gjson.Get(out, "Secrets.ClientSecret").String())
	assert.Equal(t, "pwd", gjson.Get(out, "Database.Password").String())
	dsn := gjson.Get(out, "Database.DSN").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "user", parsed.User)
	assert.Equal(t, "pwd", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
}

func TestClusterRegister_SiteURLFlag(t *testing.T) {
	conf := get.Config()
	prev := conf.Options().SiteUrl
	conf.Options().SiteUrl = ""
	defer func() { conf.Options().SiteUrl = prev }()

	const site = "https://public.example.test/"
	const advertise = "https://internal.example.test/"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, site, gjson.GetBytes(body, "SiteUrl").String())
		assert.Equal(t, advertise, gjson.GetBytes(body, "AdvertiseUrl").String())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n-site",
				Name:      "neon",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
				SiteUrl:   site,
			},
			Secrets: &cluster.RegisterSecrets{ClientSecret: cluster.ExampleClientSecret, RotatedAt: "2025-09-15T00:00:00Z"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register",
		"--name", "neon",
		"--advertise-url", advertise,
		"--site-url", site,
		"--portal-url", ts.URL,
		"--join-token", cluster.ExampleJoinToken,
		"--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, site, gjson.Get(out, "Node.SiteUrl").String())
}

func TestClusterNodesRotate_HTTPHappyPath(t *testing.T) {
	// Fake Portal register endpoint for rotation
	secret := cluster.ExampleClientSecret
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n1",
				Name:      "pp-node-03",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwd2",
				DSN:       "user:pwd2@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			Secrets: &cluster.RegisterSecrets{
				ClientSecret: secret,
				RotatedAt:    "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: false,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", cluster.ExampleJoinToken)
	_ = os.Setenv("PHOTOPRISM_CLI", "noninteractive")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	defer os.Unsetenv("PHOTOPRISM_CLI")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--portal-url=" + ts.URL, "--join-token=" + cluster.ExampleJoinToken, "--db", "--secret", "--yes", "pp-node-03",
	})
	assert.NoError(t, err)
	assert.Contains(t, out, "pp-node-03")
	assert.Contains(t, out, "Node Client Secret")
	assert.Contains(t, out, "DB Password")
}

func TestClusterNodesRotate_HTTPJson(t *testing.T) {
	// Fake Portal register endpoint for rotation in JSON mode
	secret := cluster.ExampleClientSecret
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n2",
				Name:      "pp-node-04",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwd3",
				DSN:       "user:pwd3@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			Secrets: &cluster.RegisterSecrets{
				ClientSecret: secret,
				RotatedAt:    "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", cluster.ExampleJoinToken)
	_ = os.Setenv("PHOTOPRISM_CLI", "noninteractive")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	defer os.Unsetenv("PHOTOPRISM_CLI")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--db", "--secret", "--yes", "pp-node-04",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-04", gjson.Get(out, "Node.Name").String())
	assert.Equal(t, secret, gjson.Get(out, "Secrets.ClientSecret").String())
	assert.Equal(t, "pwd3", gjson.Get(out, "Database.Password").String())
	dsn := gjson.Get(out, "Database.DSN").String()
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
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Read payload to assert rotate flags
		b, _ := io.ReadAll(r.Body)
		rotate := gjson.GetBytes(b, "RotateDatabase").Bool()
		rotateSecret := gjson.GetBytes(b, "RotateSecret").Bool()
		// Expect DB rotation only
		if !rotate || rotateSecret {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n3",
				Name:      "pp-node-05",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwd4",
				DSN:       "pp_user:pwd4@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", cluster.ExampleJoinToken)
	_ = os.Setenv("PHOTOPRISM_YES", "true")
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	defer os.Unsetenv("PHOTOPRISM_YES")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--db", "--yes", "pp-node-05",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-05", gjson.Get(out, "Node.Name").String())
	assert.Equal(t, "pwd4", gjson.Get(out, "Database.Password").String())
	dsn := gjson.Get(out, "Database.DSN").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "pp_user", parsed.User)
	assert.Equal(t, "pwd4", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
	assert.Equal(t, "", gjson.Get(out, "Secrets.ClientSecret").String())
}

func TestClusterNodesRotate_SecretOnly_JSON(t *testing.T) {
	secret := cluster.ExampleClientSecret
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		rotate := gjson.GetBytes(b, "RotateDatabase").Bool()
		rotateSecret := gjson.GetBytes(b, "RotateSecret").Bool()
		// Expect secret-only rotation
		if rotate || !rotateSecret {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n4",
				Name:      "pp-node-06",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			Secrets: &cluster.RegisterSecrets{
				ClientSecret: secret,
				RotatedAt:    "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	_ = os.Setenv("PHOTOPRISM_PORTAL_URL", ts.URL)
	_ = os.Setenv("PHOTOPRISM_JOIN_TOKEN", cluster.ExampleJoinToken)
	defer os.Unsetenv("PHOTOPRISM_PORTAL_URL")
	defer os.Unsetenv("PHOTOPRISM_JOIN_TOKEN")
	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--secret", "--yes", "pp-node-06",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-06", gjson.Get(out, "Node.Name").String())
	assert.Equal(t, secret, gjson.Get(out, "Secrets.ClientSecret").String())
	assert.Equal(t, "", gjson.Get(out, "Database.Password").String())
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
		"register", "--name", "pp-node-conflict", "--role", "instance", "--portal-url", ts.URL, "--join-token", cluster.ExampleJoinToken, "--json",
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
	// Should not fail; output must include PortalUrl and payload
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.NotEmpty(t, gjson.Get(out, "PortalUrl").String())
	assert.Equal(t, "instance", gjson.Get(out, "Payload.NodeRole").String())
	// NodeName may be derived; ensure non-empty
	assert.NotEmpty(t, gjson.Get(out, "Payload.NodeName").String())
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
		"register", "--name", "pp node invalid", "--role", "instance", "--portal-url", ts.URL, "--join-token", cluster.ExampleJoinToken, "--json",
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
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n7",
				Name:      "pp-node-rl",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwdrl",
				DSN:       "pp_user:pwdrl@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-rl", "--role", "instance", "--portal-url", ts.URL, "--join-token", cluster.ExampleJoinToken, "--rotate", "--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-rl", gjson.Get(out, "Node.Name").String())
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
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=" + cluster.ExampleJoinToken, "--db", "--yes", "pp-node-x",
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
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=" + cluster.ExampleJoinToken, "--db", "--yes", "pp node invalid",
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
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n8",
				Name:      "pp-node-rl2",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwdrl2",
				DSN:       "pp_user:pwdrl2@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterNodesRotateCommand, []string{
		"rotate", "--json", "--portal-url=" + ts.URL, "--join-token=" + cluster.ExampleJoinToken, "--db", "--yes", "pp-node-rl2",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-rl2", gjson.Get(out, "Node.Name").String())
}

func TestClusterRegister_RotateDatabase_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if !gjson.GetBytes(b, "RotateDatabase").Bool() || gjson.GetBytes(b, "RotateSecret").Bool() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n5",
				Name:      "pp-node-07",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				Password:  "pwd7",
				DSN:       "pp_user:pwd7@tcp(db:3306)/pp_db?parseTime=true",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-07", "--role", "instance", "--portal-url", ts.URL, "--join-token", cluster.ExampleJoinToken, "--rotate", "--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-07", gjson.Get(out, "Node.Name").String())
	assert.Equal(t, "pwd7", gjson.Get(out, "Database.Password").String())
	dsn := gjson.Get(out, "Database.DSN").String()
	parsed := cfg.NewDSN(dsn)
	assert.Equal(t, "pp_user", parsed.User)
	assert.Equal(t, "pwd7", parsed.Password)
	assert.Equal(t, "tcp", parsed.Net)
	assert.Equal(t, "db:3306", parsed.Server)
	assert.Equal(t, "pp_db", parsed.Name)
}

func TestClusterRegister_RotateSecret_JSON(t *testing.T) {
	secret := cluster.ExampleClientSecret
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cluster/nodes/register" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+cluster.ExampleJoinToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if gjson.GetBytes(b, "RotateDatabase").Bool() || !gjson.GetBytes(b, "RotateSecret").Bool() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := cluster.RegisterResponse{
			Node: cluster.Node{
				UUID:      "n6",
				Name:      "pp-node-08",
				Role:      "instance",
				CreatedAt: "2025-09-15T00:00:00Z",
				UpdatedAt: "2025-09-15T00:00:00Z",
			},
			Database: cluster.RegisterDatabase{
				Host:      "database",
				Port:      3306,
				Name:      "pp_db",
				User:      "pp_user",
				RotatedAt: "2025-09-15T00:00:00Z",
			},
			Secrets: &cluster.RegisterSecrets{
				ClientSecret: secret,
				RotatedAt:    "2025-09-15T00:00:00Z",
			},
			AlreadyRegistered:  true,
			AlreadyProvisioned: true,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	out, err := RunWithTestContext(ClusterRegisterCommand, []string{
		"register", "--name", "pp-node-08", "--role", "instance", "--portal-url", ts.URL, "--join-token", cluster.ExampleJoinToken, "--rotate-secret", "--json",
	})
	assert.NoError(t, err)
	assert.Equal(t, "pp-node-08", gjson.Get(out, "Node.Name").String())
	assert.Equal(t, secret, gjson.Get(out, "Secrets.ClientSecret").String())
	assert.Equal(t, "", gjson.Get(out, "Database.Password").String())
}
