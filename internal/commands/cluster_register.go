package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	clusternode "github.com/photoprism/photoprism/internal/service/cluster/node"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/service/http/header"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// Supported cluster node register flags.
var (
	regNameFlag       = &cli.StringFlag{Name: "name", Usage: "node `NAME` (lowercase letters, digits, hyphens)"}
	regRoleFlag       = &cli.StringFlag{Name: "role", Usage: "node `ROLE` (instance, service)", Value: "instance"}
	regIntUrlFlag     = &cli.StringFlag{Name: "advertise-url", Usage: "internal service `URL`"}
	regSiteUrlFlag    = &cli.StringFlag{Name: "site-url", Usage: "public site `URL` (https://...)"}
	regLabelFlag      = &cli.StringSliceFlag{Name: "label", Usage: "`k=v` label (repeatable)"}
	regRotateDatabase = &cli.BoolFlag{Name: "rotate", Usage: "rotates the node's database password"}
	regRotateSec      = &cli.BoolFlag{Name: "rotate-secret", Usage: "rotates the node's secret used for JWT"}
	regPortalURL      = &cli.StringFlag{Name: "portal-url", Usage: "Portal base `URL` (defaults to config)"}
	regPortalTok      = &cli.StringFlag{Name: "join-token", Usage: "Portal access `TOKEN` (defaults to config)"}
	regWriteConf      = &cli.BoolFlag{Name: "write-config", Usage: "persists returned secrets and DB settings to local config"}
	regForceFlag      = &cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "confirm actions that may overwrite/replace local data (e.g., --write-config)"}
	regDryRun         = DryRunFlag("print derived values and payload without performing registration")
)

// ClusterRegisterCommand registers a node with the Portal via HTTP.
// ClusterRegisterCommand wires the `cluster register` CLI entrypoint.
var ClusterRegisterCommand = &cli.Command{
	Name:  "register",
	Usage: "Registers a node or updates its credentials within a cluster",
	Flags: append(append([]cli.Flag{
		regDryRun,
		regNameFlag,
		regRoleFlag,
		regPortalURL,
		regPortalTok,
		regIntUrlFlag,
		regSiteUrlFlag,
		regLabelFlag,
		regRotateDatabase,
		regRotateSec,
		regWriteConf,
		regForceFlag,
	}, report.CliFlags...)),
	Action: clusterRegisterAction,
}

// clusterRegisterAction resolves CLI flags, builds the registration payload,
// and calls the Portal's register endpoint with retry/backoff handling.
func clusterRegisterAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		// Resolve inputs
		name := clean.DNSLabel(ctx.String("name"))
		derivedName := false

		if name == "" { // default from config if set
			name = clean.DNSLabel(conf.NodeName())
			if name != "" {
				derivedName = true
			}
		}

		if name == "" {
			return cli.Exit(fmt.Errorf("node name is required (use --name or set node-name)"), 2)
		}

		nodeRole := clean.TypeLowerDash(ctx.String("role"))
		switch nodeRole {
		case "instance", "service":
		default:
			return cli.Exit(fmt.Errorf("invalid --role (must be instance or service)"), 2)
		}

		portalURL := ctx.String("portal-url")
		derivedPortal := false
		if portalURL == "" {
			portalURL = conf.PortalUrl()
			if portalURL != "" {
				derivedPortal = true
			}
		}

		// Derive advertise/site URLs when omitted.
		advertise := ctx.String("advertise-url")
		if advertise == "" {
			advertise = conf.AdvertiseUrl()
		}
		site := strings.TrimSpace(ctx.String("site-url"))
		if site == "" {
			site = conf.SiteUrl()
		}

		payload := cluster.RegisterRequest{
			NodeName:       name,
			NodeRole:       nodeRole,
			Labels:         parseLabelSlice(ctx.StringSlice("label")),
			AdvertiseUrl:   advertise,
			RotateDatabase: ctx.Bool("rotate"),
			RotateSecret:   ctx.Bool("rotate-secret"),
		}

		// If auto detection is allowed, rotate database only when the current node lacks configured credentials.
		if !payload.RotateDatabase && conf.ShouldAutoRotateDatabase() {
			payload.RotateDatabase = true
		}

		// If we already have client credentials for this node (e.g., re-registering the
		// same instance), include them so the portal can verify UUID/name changes. Avoid
		// sending the portal's own credentials when registering a different node.
		if id, secret := strings.TrimSpace(conf.NodeClientID()), strings.TrimSpace(conf.NodeClientSecret()); id != "" && secret != "" && strings.EqualFold(conf.NodeName(), name) {
			payload.ClientID = id
			payload.ClientSecret = secret
		}

		if site != "" {
			payload.SiteUrl = site
		}
		b, _ := json.Marshal(payload)

		// In dry-run, we allow empty portalURL (will print derived/empty values).
		if ctx.Bool("dry-run") {
			if ctx.Bool("json") {
				out := struct {
					PortalURL string                  `json:"PortalUrl"`
					Payload   cluster.RegisterRequest `json:"Payload"`
				}{PortalURL: portalURL, Payload: payload}
				jb, _ := json.Marshal(out)
				fmt.Println(string(jb))
			} else {
				fmt.Printf("Portal URL: %s\n", portalURL)
				fmt.Printf("Node Name:  %s\n", name)
				if derivedPortal || derivedName || advertise == conf.AdvertiseUrl() {
					fmt.Println("(derived defaults were used where flags were omitted)")
				}
				fmt.Printf("Advertise:  %s\n", advertise)
				if payload.SiteUrl != "" {
					fmt.Printf("Site URL:   %s\n", payload.SiteUrl)
				}
				// Warn if non-HTTPS on public host; server will enforce too.
				if warnInsecurePublicURL(advertise) {
					fmt.Println("Warning: advertise-url is http for a public host; server may reject it (HTTPS required).")
				}
				if payload.SiteUrl != "" && warnInsecurePublicURL(payload.SiteUrl) {
					fmt.Println("Warning: site-url is http for a public host; server may reject it (HTTPS required).")
				}
				// Single-line summary for quick operator scan
				if payload.SiteUrl != "" {
					fmt.Printf("Derived: portal=%s advertise=%s site=%s\n", portalURL, advertise, payload.SiteUrl)
				} else {
					fmt.Printf("Derived: portal=%s advertise=%s\n", portalURL, advertise)
				}
			}
			return nil
		}

		// For actual registration, require portal URL and token.
		if portalURL == "" {
			return cli.Exit(fmt.Errorf("portal URL is required (use --portal-url or set portal-url)"), 2)
		}

		token := ctx.String("join-token")

		if token == "" {
			token = conf.JoinToken()
		}

		if token == "" {
			return cli.Exit(fmt.Errorf("portal token is required (use --join-token or set join-token)"), 2)
		}

		// POST with bounded backoff on 429
		endpointUrl := stringsTrimRightSlash(portalURL) + "/api/v1/cluster/nodes/register"

		var resp cluster.RegisterResponse
		if err := postWithBackoff(endpointUrl, token, b, &resp); err != nil {
			var httpErr *httpError
			if errors.As(err, &httpErr) && httpErr.Status == http.StatusTooManyRequests {
				return cli.Exit(fmt.Errorf("portal rate-limited registration attempts"), 6)
			}
			// Map common errors
			if errors.As(err, &httpErr) {
				switch httpErr.Status {
				case http.StatusUnauthorized, http.StatusForbidden:
					return cli.Exit(fmt.Errorf("%s", httpErr.Error()), 4)
				case http.StatusConflict:
					return cli.Exit(fmt.Errorf("%s", httpErr.Error()), 5)
				case http.StatusBadRequest:
					return cli.Exit(fmt.Errorf("%s", httpErr.Error()), 2)
				case http.StatusNotFound:
					return cli.Exit(fmt.Errorf("%s", httpErr.Error()), 3)
				}
			}
			return cli.Exit(err, 1)
		}

		// Output
		if ctx.Bool("json") {
			jb, _ := json.Marshal(resp)
			fmt.Println(string(jb))
		} else {
			// Human-readable: node row and credentials if present (UUID first as primary identifier)
			cols := []string{"UUID", "ClientID", "Name", "Role"}
			row := []string{resp.Node.UUID, resp.Node.ClientID, resp.Node.Name, resp.Node.Role}

			if resp.Database.Driver != "" {
				cols = append(cols, "DB Driver")
				row = append(row, resp.Database.Driver)
			}
			if resp.Database.Name != "" {
				cols = append(cols, "DB Name")
				row = append(row, resp.Database.Name)
			}
			if resp.Database.User != "" {
				cols = append(cols, "DB User")
				row = append(row, resp.Database.User)
			}
			if resp.Database.Host != "" {
				cols = append(cols, "Host")
				row = append(row, resp.Database.Host)
			}
			if resp.Database.Port > 0 {
				cols = append(cols, "Port")
				row = append(row, strconv.Itoa(resp.Database.Port))
			}

			rows := [][]string{row}
			out, _ := report.RenderFormat(rows, cols, report.CliFormat(ctx))
			fmt.Printf("\n%s\n", out)

			// Secrets/credentials block if any
			// Show secrets in up to two tables, then print DSN if present
			if (resp.Secrets != nil && resp.Secrets.ClientSecret != "") || resp.Database.Password != "" {
				fmt.Println("PLEASE WRITE DOWN THE FOLLOWING CREDENTIALS; THEY WILL NOT BE SHOWN AGAIN:")
				if resp.Secrets != nil && resp.Secrets.ClientSecret != "" {
					secretTable, _ := report.RenderFormat([][]string{{resp.Secrets.ClientSecret}}, []string{"Node Client Secret"}, report.CliFormat(ctx))
					fmt.Printf("\n%s\n", secretTable)
				}
				if resp.Database.Password != "" {
					dbTable, _ := report.RenderFormat([][]string{{resp.Database.User, resp.Database.Password}}, []string{"DB User", "DB Password"}, report.CliFormat(ctx))
					fmt.Printf("\n%s\n", dbTable)
				}
				if resp.Database.DSN != "" {
					fmt.Printf("DSN: %s\n", resp.Database.DSN)
				}
			}
		}

		// Optional persistence
		if ctx.Bool("write-config") {
			if err := persistRegisterResponse(conf, &resp); err != nil {
				return err
			}
		}

		return nil
	})
}

// HTTP helpers and backoff

type httpError struct {
	Status int
	Body   string
}

func (e *httpError) Error() string { return fmt.Sprintf("http %d: %s", e.Status, e.Body) }

// postWithBackoff executes the register HTTP POST with bounded retries for 429 responses.
func postWithBackoff(url, token string, payload []byte, out any) error {
	// backoff: 500ms -> max ~8s, 6 attempts with jitter
	delay := 500 * time.Millisecond
	for attempt := 0; attempt < 6; attempt++ {
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		header.SetAuthorization(req, token)
		req.Header.Set(header.ContentType, "application/json")

		client := &http.Client{Timeout: cluster.BootstrapRegisterTimeout}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			// backoff and retry
			time.Sleep(jitter(delay, 0.25))
			if delay < 8*time.Second {
				delay *= 2
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(resp.Body)
			return &httpError{Status: resp.StatusCode, Body: string(b)}
		}
		dec := json.NewDecoder(resp.Body)
		return dec.Decode(out)
	}
	return &httpError{Status: http.StatusTooManyRequests, Body: "rate limited"}
}

// jitter applies +/- jitter to a duration to avoid retry stampedes.
func jitter(d time.Duration, frac float64) time.Duration {
	// simple +/- jitter
	n := time.Duration(float64(d) * (1 + (randFloat()*2-1)*frac))
	if n <= 0 {
		return d
	}
	return n
}

// randFloat returns a simple pseudo-random float in [0,1) without touching math/rand global state.
func randFloat() float64 { return float64(time.Now().UnixNano()%1000) / 1000.0 }

// stringsTrimRightSlash removes trailing slashes to build consistent endpoints.
func stringsTrimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// warnInsecurePublicURL returns true if the URL is HTTP on a non-local host.
func warnInsecurePublicURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	if parsed.Scheme != "http" {
		return false
	}
	h := parsed.Hostname()
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return false
	}
	return true
}

// parseLabelSlice converts repeated k=v CLI inputs into a map.
func parseLabelSlice(labels []string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, kv := range labels {
		if i := bytes.IndexByte([]byte(kv), '='); i > 0 && i < len(kv)-1 {
			k := kv[:i]
			v := kv[i+1:]
			m[k] = v
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// persistRegisterResponse writes returned secrets/DB details into local config when requested.
func persistRegisterResponse(conf *config.Config, resp *cluster.RegisterResponse) error {
	updates := cluster.OptionsUpdate{}

	if rnd.IsUUID(resp.UUID) {
		updates.SetClusterUUID(resp.UUID)
	}

	if cidr := strings.TrimSpace(resp.ClusterCIDR); cidr != "" {
		updates.SetClusterCIDR(cidr)
	}

	// Node client secret file
	if resp.Secrets != nil && resp.Secrets.ClientSecret != "" {
		// Prefer PHOTOPRISM_NODE_CLIENT_SECRET_FILE; otherwise config cluster path
		fileName := os.Getenv(config.FlagFileVar("NODE_CLIENT_SECRET"))
		if fileName == "" {
			fileName = filepath.Join(conf.PortalConfigPath(), "node-secret")
		}
		if err := fs.MkdirAll(filepath.Dir(fileName)); err != nil {
			return err
		}
		if err := os.WriteFile(fileName, []byte(resp.Secrets.ClientSecret), 0o600); err != nil {
			return err
		}
		log.Infof("wrote node client secret to %s", clean.Log(fileName))
	}

	// DB settings (MySQL/MariaDB only)
	if resp.Database.Name != "" && resp.Database.User != "" {
		updates.SetDatabaseDriver(config.MySQL)
		updates.SetDatabaseName(resp.Database.Name)
		updates.SetDatabaseServer(fmt.Sprintf("%s:%d", resp.Database.Host, resp.Database.Port))
		updates.SetDatabaseUser(resp.Database.User)
		updates.SetDatabasePassword(resp.Database.Password)
	}

	if !updates.IsZero() {
		if _, err := clusternode.ApplyOptionsUpdate(conf, updates); err != nil {
			return err
		}
		log.Infof("updated options.yml with cluster registration settings for node %s", clean.LogQuote(resp.Node.Name))
	}
	return nil
}
