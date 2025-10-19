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
	"unicode/utf8"

	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var log = event.Log

// init registers the cluster node bootstrap extension so it runs before the
// database connection is established.
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
		log.Debugf("cluster: skipped node bootstrap role=%s", role)
		return nil
	}

	portalURL := strings.TrimSpace(c.PortalUrl())
	joinToken := strings.TrimSpace(c.JoinToken())

	if portalURL == "" || joinToken == "" {
		log.Debugf("cluster: skipped node bootstrap portalUrl=%s joinToken=%s", portalURL, strings.Repeat("*", utf8.RuneCountInString(joinToken)))
		return nil
	}

	log.Debugf("cluster: starting node bootstrap")

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
		if err = registerWithPortal(c, u, joinToken); err != nil {
			// Registration errors are expected when the Portal is temporarily unavailable
			// or not configured with cluster endpoints (404). Keep as warn to signal
			// exhaustion/terminal errors; per-attempt details are logged at debug level.
			log.Warnf("cluster: could not join (%s)", clean.Error(err))
		}
	}

	// Pull theme if missing.
	if cluster.BootstrapAutoThemeEnabled {
		if err = installThemeIfMissing(c, u, joinToken); err != nil {
			// Theme install failures are non-critical; log at debug to avoid noise.
			log.Debugf("cluster: could not install theme (%s)", clean.Error(err))
		}
	}

	// Log cluster UUID.
	if uuid := c.ClusterUUID(); uuid != "" {
		log.Infof("cluster: UUID %s", clean.Log(uuid))
	}

	return nil
}

// isLocalHost reports whether the given host string refers to a local loopback
// address that may safely use plain HTTP during bootstrap.
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

// newHTTPClient returns a short-lived HTTP client configured with the provided
// timeout. It is intentionally lightweight to avoid leaking transports between
// bootstrap attempts.
func newHTTPClient(timeout time.Duration) *http.Client {
	// TODO: Consider reusing a shared *http.Transport with sane defaults and enabling
	// proxy support explicitly if required. For now, rely on net/http defaults and
	// the HTTPS_PROXY set in config.Init().
	return &http.Client{Timeout: timeout}
}

// registerWithPortal attempts to register the node with the Portal, retrying on
// transient errors up to the configured limits. Successful registrations update
// local configuration and prime JWKS credentials.
func registerWithPortal(c *config.Config, portal *url.URL, token string) error {
	maxAttempts := cluster.BootstrapRegisterMaxAttempts
	delay := cluster.BootstrapRegisterRetryDelay
	timeout := cluster.BootstrapRegisterTimeout

	endpoint := *portal
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/cluster/nodes/register"

	// Let the configuration decide if credentials are missing (MySQL with no effective name/user/password).
	wantRotateDatabase := c.ShouldAutoRotateDatabase()

	payload := cluster.RegisterRequest{
		NodeName:     c.NodeName(),
		NodeUUID:     c.NodeUUID(),
		NodeRole:     c.NodeRole(),
		AdvertiseUrl: c.AdvertiseUrl(),
	}

	// Auto-derive Advertise/Site URLs from node name and cluster domain when not configured.
	if domain := strings.TrimSpace(defaultClusterDomain(c)); domain != "" {
		if payload.NodeName == "" {
			payload.NodeName = c.NodeName()
		}

		if payload.AdvertiseUrl == "" {
			if u := defaultNodeURL(payload.NodeName, domain); u != "" {
				payload.AdvertiseUrl = u
			}
		}

		if payload.SiteUrl == "" && payload.AdvertiseUrl != "" {
			payload.SiteUrl = payload.AdvertiseUrl
		}
	}

	// Include client credentials when present so the Portal can verify re-registration
	// and authorize UUID/name changes.
	if id, secret := strings.TrimSpace(c.NodeClientID()), strings.TrimSpace(c.NodeClientSecret()); id != "" && secret != "" {
		payload.ClientID = id
		payload.ClientSecret = secret
	}

	// Include SiteUrl whenever configured; the server normalizes duplicates if needed.
	if su := c.SiteUrl(); su != "" {
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
				log.Debugf("cluster: join attempt %d/%d error: %s", attempt, maxAttempts, clean.Error(err))
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
			if err = dec.Decode(&r); err != nil {
				return err
			}
			if err = persistRegistration(c, &r, wantRotateDatabase); err != nil {
				return err
			}
			primeJWKS(c, r.JWKSUrl)
			if resp.StatusCode == http.StatusCreated {
				log.Infof("cluster: joined as %s (%d)", clean.LogQuote(r.Node.Name), resp.StatusCode)
			} else {
				log.Infof("cluster: registration ok (%d)", resp.StatusCode)
			}
			return nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			// Terminal errors (no retry). 404 likely indicates a Portal without cluster endpoints.
			return errors.New(resp.Status)
		case http.StatusTooManyRequests:
			if attempt < maxAttempts {
				log.Debugf("cluster: join attempt %d/%d rate limited", attempt, maxAttempts)
				time.Sleep(delay)
				continue
			}
			return errors.New(resp.Status)
		case http.StatusConflict, http.StatusBadRequest:
			// Do not retry on 400/409 per spec intent.
			return errors.New(resp.Status)
		default:
			if attempt < maxAttempts {
				log.Debugf("cluster: join attempt %d/%d server responded %s", attempt, maxAttempts, resp.Status)
				// TODO: Consider exponential backoff with jitter instead of constant delay.
				time.Sleep(delay)
				continue
			}
			return errors.New(resp.Status)
		}
	}
	return nil
}

// defaultClusterDomain returns the configured cluster domain or, if absent,
// attempts to derive it from the Portal URL by stripping common prefixes.
func defaultClusterDomain(c *config.Config) string {
	if c == nil {
		return ""
	}

	domain := strings.TrimSpace(c.ClusterDomain())

	if domain != "" {
		return strings.Trim(domain, ".")
	}

	portalURL := strings.TrimSpace(c.PortalUrl())

	if portalURL == "" {
		return ""
	}

	u, err := url.Parse(portalURL)

	if err != nil {
		return ""
	}

	host := strings.Trim(u.Hostname(), ".")

	if host == "" {
		return ""
	}

	// Strip common prefixes like portal.<domain>.
	if isLocalHost(host) {
		return ""
	}

	if ip := net.ParseIP(host); ip != nil {
		// Prefer DNS domains over raw IP addresses; leave empty so caller can decide.
		return ""
	}

	if strings.HasPrefix(host, "portal.") && len(host) > len("portal.") {
		return strings.TrimPrefix(host, "portal.")
	}

	return host
}

// defaultNodeURL builds https://<name>.<domain> using sanitized labels.
func defaultNodeURL(name, domain string) string {
	name = clean.TypeLowerDash(strings.TrimSpace(name))
	domain = strings.Trim(strings.ToLower(domain), ".")

	if name == "" || domain == "" {
		return ""
	}

	return fmt.Sprintf("https://%s.%s", name, domain)
}

// isTemporary reports whether the given error represents a temporary network
// failure that merits another retry attempt.
func isTemporary(err error) bool {
	var nerr net.Error
	return errors.As(err, &nerr) && nerr.Timeout()
}

// persistRegistration merges registration responses into options.yml and, when
// necessary, reloads the in-memory configuration so future bootstrap steps use
// the updated values.
func persistRegistration(c *config.Config, r *cluster.RegisterResponse, wantRotateDatabase bool) error {
	updates := cluster.OptionsUpdate{}

	// Persist ClusterUUID from portal response if provided.
	if rnd.IsUUID(r.UUID) {
		updates.SetClusterUUID(r.UUID)
	}

	if cidr := strings.TrimSpace(r.ClusterCIDR); cidr != "" {
		updates.SetClusterCIDR(cidr)
	}

	// Always persist NodeClientID (client UID) from response for future OAuth token requests.
	if r.Node.ClientID != "" {
		updates.SetNodeClientID(r.Node.ClientID)
	}

	// Persist node client secret only if missing locally and provided by server.
	if r.Secrets != nil && r.Secrets.ClientSecret != "" && c.NodeClientSecret() == "" {
		updates.SetNodeClientSecret(r.Secrets.ClientSecret)
	}

	if jwksUrl := strings.TrimSpace(r.JWKSUrl); jwksUrl != "" {
		updates.SetJWKSUrl(jwksUrl)
		c.SetJWKSUrl(jwksUrl)
	}

	// Persist NodeUUID from portal response if provided and not set locally.
	if r.Node.UUID != "" && c.NodeUUID() == "" {
		updates.SetNodeUUID(r.Node.UUID)
	}

	// Persist DB settings only if rotation was requested and driver is MySQL/MariaDB
	// and local DB not configured (as checked before calling).
	if wantRotateDatabase {
		if r.Database.DSN != "" {
			updates.SetDatabaseDriver(r.Database.Driver)
			updates.SetDatabaseDSN(r.Database.DSN)
		} else if r.Database.Name != "" && r.Database.User != "" && r.Database.Password != "" {
			server := r.Database.Host
			if r.Database.Port > 0 {
				server = net.JoinHostPort(r.Database.Host, strconv.Itoa(r.Database.Port))
			}
			updates.SetDatabaseDriver(r.Database.Driver)
			updates.SetDatabaseServer(server)
			updates.SetDatabaseName(r.Database.Name)
			updates.SetDatabaseUser(r.Database.User)
			updates.SetDatabasePassword(r.Database.Password)
		}
	}

	if updates.IsZero() {
		return nil
	}

	wrote, err := ApplyOptionsUpdate(c, updates)

	if err != nil {
		return err
	}

	if wrote {
		// Reload into memory so later code paths see updated values during this run.
		_ = c.Options().Load(c.OptionsYaml())

		if updates.HasDatabaseUpdate() {
			log.Infof("cluster: database settings applied; restart required to take effect")
		}
	}

	return nil
}

// primeJWKS eagerly fetches the Portal JWKS so that subsequent token
// verification does not incur network latency during critical operations.
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
		if err = fs.MkdirAll(themeDir); err != nil {
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

	if err = dec.Decode(&tok); err != nil {
		return "", err
	}

	if tok.AccessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}

	return tok.AccessToken, nil
}
