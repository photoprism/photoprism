package node

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var log = event.Log

// Values is an shorthand alias for map[string]interface{}.
type Values = map[string]interface{}

func init() {
	// Register early so this can adjust DB settings before connectDb().
	config.RegisterEarly("cluster-node", InitConfig, nil)
}

// InitConfig performs node bootstrap: optional registration with the Portal
// and theme installation. Runs early during config.Init().
func InitConfig(c *config.Config) error {
	if !cluster.BootstrapAutoJoinEnabled && !cluster.BootstrapAutoThemeEnabled {
		return nil
	}

	role := c.NodeRole()
	// Skip on portal nodes and unknown node types.
	if c.Portal() || (role != cluster.RoleInstance && role != cluster.RoleService) {
		return nil
	}

	portalURL := strings.TrimSpace(c.PortalUrl())
	joinToken := strings.TrimSpace(c.JoinToken())
	if portalURL == "" || joinToken == "" {
		return nil
	}

	u, err := url.Parse(portalURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		log.Warnf("cluster: invalid portal url %s", clean.Log(portalURL))
		return nil
	}

	// Enforce TLS for non-local URLs.
	if u.Scheme != "https" && !isLocalHost(u.Hostname()) {
		log.Warnf("cluster: refusing non-TLS portal url %s on non-local host", clean.Log(portalURL))
		return nil
	}

	// Register with retry policy.
	if cluster.BootstrapAutoJoinEnabled {
		if err := registerWithPortal(c, u, joinToken); err != nil {
			// Registration errors are expected when the Portal is temporarily unavailable
			// or not configured with cluster endpoints (404). Keep as warn to signal
			// exhaustion/terminal errors; per-attempt details are logged at debug level.
			log.Warnf("cluster: register failed (%s)", clean.Error(err))
		}
	}

	// Pull theme if missing.
	if cluster.BootstrapAutoThemeEnabled {
		if err := installThemeIfMissing(c, u, joinToken); err != nil {
			// Theme install failures are non-critical; log at debug to avoid noise.
			log.Debugf("cluster: theme install skipped/failed (%s)", clean.Error(err))
		}
	}

	return nil
}

func isLocalHost(h string) bool {
	switch strings.ToLower(h) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		// TODO: Consider treating RFC1918/link-local hosts as local for TLS enforcement
		// if the operator explicitly opts in (e.g., via a policy var). Keep simple for now.
		return false
	}
}

func newHTTPClient(timeout time.Duration) *http.Client {
	// TODO: Consider reusing a shared *http.Transport with sane defaults and enabling
	// proxy support explicitly if required. For now, rely on net/http defaults and
	// the HTTPS_PROXY set in config.Init().
	return &http.Client{Timeout: timeout}
}

func registerWithPortal(c *config.Config, portal *url.URL, token string) error {
	maxAttempts := cluster.BootstrapRegisterMaxAttempts
	delay := cluster.BootstrapRegisterRetryDelay
	timeout := cluster.BootstrapRegisterTimeout

	endpoint := *portal
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/cluster/nodes/register"

	// Decide if DB rotation is desired as per spec: only if driver is MySQL/MariaDB
	// and no DSN/fields are set (raw options) and no password is provided via file.
	opts := c.Options()
	driver := c.DatabaseDriver()
	wantRotateDatabase := (driver == config.MySQL || driver == config.MariaDB) &&
		opts.DatabaseDSN == "" && opts.DatabaseName == "" && opts.DatabaseUser == "" && opts.DatabasePassword == "" &&
		c.DatabasePassword() == ""

	payload := cluster.RegisterRequest{
		NodeName:     c.NodeName(),
		NodeUUID:     c.NodeUUID(),
		NodeRole:     c.NodeRole(),
		AdvertiseUrl: c.AdvertiseUrl(),
	}
	// Include client credentials when present so the Portal can verify re-registration
	// and authorize UUID/name changes.
	if id, secret := strings.TrimSpace(c.NodeClientID()), strings.TrimSpace(c.NodeClientSecret()); id != "" && secret != "" {
		payload.ClientID = id
		payload.ClientSecret = secret
	}

	// Include siteUrl when it differs from advertiseUrl; server will validate/normalize.
	if su := c.SiteUrl(); su != "" && su != c.AdvertiseUrl() {
		payload.SiteUrl = su
	}

	if wantRotateDatabase {
		// Align with API: request database rotation/creation on (re)register.
		payload.RotateDatabase = true
	}

	bodyBytes, _ := json.Marshal(payload)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, _ := http.NewRequest(http.MethodPost, endpoint.String(), strings.NewReader(string(bodyBytes)))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := newHTTPClient(timeout).Do(req)
		if err != nil {
			if attempt < maxAttempts {
				log.Debugf("cluster: register attempt %d/%d error: %s", attempt, maxAttempts, clean.Error(err))
				time.Sleep(delay)
				continue
			}
			return err
		}

		// Ensure body is closed after handling the response.
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			var r cluster.RegisterResponse
			dec := json.NewDecoder(resp.Body)
			if err := dec.Decode(&r); err != nil {
				return err
			}
			if err := persistRegistration(c, &r, wantRotateDatabase); err != nil {
				return err
			}
			primeJWKS(c, r.JWKSUrl)
			if resp.StatusCode == http.StatusCreated {
				log.Infof("cluster: registered as %s (%d)", clean.LogQuote(r.Node.Name), resp.StatusCode)
			} else {
				log.Infof("cluster: registration ok (%d)", resp.StatusCode)
			}
			return nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			// Terminal errors (no retry). 404 likely indicates a Portal without cluster endpoints.
			return errors.New(resp.Status)
		case http.StatusTooManyRequests:
			if attempt < maxAttempts {
				log.Debugf("cluster: register attempt %d/%d rate limited", attempt, maxAttempts)
				time.Sleep(delay)
				continue
			}
			return errors.New(resp.Status)
		case http.StatusConflict, http.StatusBadRequest:
			// Do not retry on 400/409 per spec intent.
			return errors.New(resp.Status)
		default:
			if attempt < maxAttempts {
				log.Debugf("cluster: register attempt %d/%d server responded %s", attempt, maxAttempts, resp.Status)
				// TODO: Consider exponential backoff with jitter instead of constant delay.
				time.Sleep(delay)
				continue
			}
			return errors.New(resp.Status)
		}
	}
	return nil
}

func isTemporary(err error) bool {
	var nerr net.Error
	return errors.As(err, &nerr) && nerr.Timeout()
}

func persistRegistration(c *config.Config, r *cluster.RegisterResponse, wantRotateDatabase bool) error {
	updates := Values{}

	// Persist ClusterUUID from portal response if provided.
	if rnd.IsUUID(r.UUID) {
		updates["ClusterUUID"] = r.UUID
	}

	if cidr := strings.TrimSpace(r.ClusterCIDR); cidr != "" {
		updates["ClusterCIDR"] = cidr
	}

	// Always persist NodeClientID (client UID) from response for future OAuth token requests.
	if r.Node.ClientID != "" {
		updates["NodeClientID"] = r.Node.ClientID
	}

	// Persist node client secret only if missing locally and provided by server.
	if r.Secrets != nil && r.Secrets.ClientSecret != "" && c.NodeClientSecret() == "" {
		updates["NodeClientSecret"] = r.Secrets.ClientSecret
	}

	if jwksUrl := strings.TrimSpace(r.JWKSUrl); jwksUrl != "" {
		updates["JWKSUrl"] = jwksUrl
		c.SetJWKSUrl(jwksUrl)
	}

	// Persist NodeUUID from portal response if provided and not set locally.
	if r.Node.UUID != "" && c.NodeUUID() == "" {
		updates["NodeUUID"] = r.Node.UUID
	}

	// Persist DB settings only if rotation was requested and driver is MySQL/MariaDB
	// and local DB not configured (as checked before calling).
	if wantRotateDatabase {
		if r.Database.DSN != "" {
			updates["DatabaseDriver"] = r.Database.Driver
			updates["DatabaseDSN"] = r.Database.DSN
		} else if r.Database.Name != "" && r.Database.User != "" && r.Database.Password != "" {
			server := r.Database.Host
			if r.Database.Port > 0 {
				server = net.JoinHostPort(r.Database.Host, strconv.Itoa(r.Database.Port))
			}
			updates["DatabaseDriver"] = r.Database.Driver
			updates["DatabaseServer"] = server
			updates["DatabaseName"] = r.Database.Name
			updates["DatabaseUser"] = r.Database.User
			updates["DatabasePassword"] = r.Database.Password
		}
	}

	if len(updates) == 0 {
		return nil
	}

	if err := mergeOptionsYaml(c, updates); err != nil {
		return err
	}

	// Reload into memory so later code paths see updated values during this run.
	_ = c.Options().Load(c.OptionsYaml())

	if hasDBUpdate(updates) {
		log.Infof("cluster: database settings applied; restart required to take effect")
	}
	return nil
}

func primeJWKS(c *config.Config, url string) {
	if c == nil {
		return
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return
	}
	verifier := clusterjwt.NewVerifier(c)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := verifier.Prime(ctx, url); err != nil {
		log.Debugf("cluster: jwks prime skipped (%s)", clean.Error(err))
	}
}

func hasDBUpdate(m Values) bool {
	if _, ok := m["DatabaseDSN"]; ok {
		return true
	}
	if _, ok := m["DatabaseName"]; ok {
		return true
	}
	if _, ok := m["DatabaseUser"]; ok {
		return true
	}
	if _, ok := m["DatabasePassword"]; ok {
		return true
	}
	if _, ok := m["DatabaseServer"]; ok {
		return true
	}
	return false
}

func mergeOptionsYaml(c *config.Config, updates Values) error {
	if err := fs.MkdirAll(c.ConfigPath()); err != nil {
		return err
	}
	fileName := c.OptionsYaml()

	var m Values
	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &m)
		}
	}
	if m == nil {
		m = Values{}
	}
	for k, v := range updates {
		m[k] = v
	}

	b, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, b, fs.ModeFile)
}

// installThemeIfMissing downloads and installs the Portal-provided theme if the
// local theme directory is missing or lacks an app.js file.
func installThemeIfMissing(c *config.Config, portal *url.URL, token string) error {
	themeDir := c.ThemePath()

	need := !fs.PathExists(themeDir) ||
		(cluster.BootstrapThemeInstallOnlyIfMissingJS && !fs.FileExists(filepath.Join(themeDir, fs.AppJsFile)))

	if !need && !cluster.BootstrapAllowThemeOverwrite {
		return nil
	}

	endpoint := *portal
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/cluster/theme"

	// Prefer OAuth client-credentials using NodeClientID/NodeClientSecret if available; fallback to join token.
	bearer := ""
	if id, secret := strings.TrimSpace(c.NodeClientID()), strings.TrimSpace(c.NodeClientSecret()); id != "" && secret != "" {
		if t, err := oauthAccessToken(c, portal, id, secret); err != nil {
			log.Debugf("cluster: oauth token request failed (%s)", clean.Error(err))
		} else {
			bearer = t
		}
	}
	// If we do not have a bearer token, skip theme install for this run (no insecure fallback).
	if bearer == "" {
		log.Debugf("cluster: theme install skipped (missing OAuth credentials)")
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/zip")

	resp, err := newHTTPClient(cluster.BootstrapRegisterTimeout).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Save to temp zip.
		if err := fs.MkdirAll(c.TempPath()); err != nil {
			return err
		}
		zipName := filepath.Join(c.TempPath(), "cluster-theme.zip")
		out, err := os.Create(zipName)
		if err != nil {
			return err
		}
		if _, err = io.Copy(out, resp.Body); err != nil {
			_ = out.Close()
			return err
		}
		_ = out.Close()

		// Extract with moderate limits.
		if err := fs.MkdirAll(themeDir); err != nil {
			return err
		}
		_, _, unzipErr := fs.Unzip(zipName, themeDir, 32*fs.MB, 512*fs.MB)
		return unzipErr
	case http.StatusNotFound:
		// No theme configured at Portal.
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return errors.New(resp.Status)
	default:
		return errors.New(resp.Status)
	}
}

// oauthAccessToken requests an OAuth access token via client_credentials using Basic auth.
func oauthAccessToken(c *config.Config, portal *url.URL, clientID, clientSecret string) (string, error) {
	if portal == nil {
		return "", fmt.Errorf("invalid portal url")
	}
	tokenURL := *portal
	tokenURL.Path = strings.TrimRight(tokenURL.Path, "/") + "/api/v1/oauth/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, _ := http.NewRequest(http.MethodPost, tokenURL.String(), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	// Basic auth for client credentials
	basic := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+basic)

	resp, err := newHTTPClient(cluster.BootstrapRegisterTimeout).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", resp.Status)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}
	return tok.AccessToken, nil
}
