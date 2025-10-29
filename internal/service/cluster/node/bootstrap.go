package node

import (
	"bytes"
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

	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/dns"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var log = event.Log

// init registers the cluster node bootstrap extension so it runs before the
// database connection is established.
func init() {
	// Register early so this can adjust DB settings before connectDb().
	config.Register(config.StageBoot, "cluster-node", InitConfig, nil)
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
		log.Debugf("config: skipping cluster bootstrap for %s", clean.Log(role))
		return nil
	}

	portalURL := strings.TrimSpace(c.PortalUrl())
	joinToken := strings.TrimSpace(c.JoinToken())

	if portalURL == "" || joinToken == "" {
		log.Debugf("config: no cluster bootstrap configuration found")
		return nil
	}

	log.Debugf("config: attempting to join the configured cluster")

	u, err := url.Parse(portalURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		log.Warnf("config: invalid cluster portal URL %s", clean.Log(portalURL))
		return nil
	}

	// Enforce TLS for non-local URLs.
	if u.Scheme != "https" && !dns.IsLoopbackHost(u.Hostname()) {
		log.Warnf("config: refusing non-TLS portal URL %s on non-local host", clean.Log(portalURL))
		return nil
	}

	// Register with retry policy.
	var registerResp *cluster.RegisterResponse
	if cluster.BootstrapAutoJoinEnabled {
		if registerResp, err = registerWithPortal(c, u, joinToken); err != nil {
			log.Warnf("config: failed to join the configured cluster (%s)", clean.Error(err))
			if isAuthError(err) {
				log.Infof("config: refreshing node credentials after authentication failure")
			}
			if isAuthError(err) && refreshNodeCredentials(c, u) {
				if registerResp, err = registerWithPortal(c, u, joinToken); err != nil {
					log.Warnf("config: retry join attempt failed (%s)", clean.Error(err))
				}
			}
		}
	}

	// Pull theme if missing or outdated, and activate it when present.
	if cluster.BootstrapAutoThemeEnabled {
		if err = syncNodeTheme(c, u, registerResp); err != nil {
			// Theme install failures are non-critical; log at debug to avoid noise.
			log.Debugf("config: portal theme download skipped (%s)", clean.Error(err))
		}
		activateNodeThemeIfPresent(c)
	}

	// Log cluster UUID.
	if uuid := c.ClusterUUID(); uuid != "" {
		log.Infof("cluster: UUID %s", clean.Log(uuid))
	}

	return nil
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
func registerWithPortal(c *config.Config, portal *url.URL, token string) (*cluster.RegisterResponse, error) {
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
		AppName:      clean.TypeUnicode(c.About()),
		AppVersion:   clean.TypeUnicode(c.Version()),
		Theme:        clean.TypeUnicode(c.NodeThemeVersion()),
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
				log.Debugf("config: join attempt %d/%d failed with network error: %s", attempt, maxAttempts, clean.Error(err))
				time.Sleep(delay)
				continue
			}
			return nil, err
		}

		// Ensure body is closed after handling the response.
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			var r cluster.RegisterResponse
			dec := json.NewDecoder(resp.Body)
			if err = dec.Decode(&r); err != nil {
				return nil, err
			}
			if err = persistRegistration(c, &r, wantRotateDatabase); err != nil {
				return nil, err
			}
			primeJWKS(c, r.JWKSUrl)
			if resp.StatusCode == http.StatusCreated {
				log.Infof("config: successfully joined cluster as node %s (%d)", clean.LogQuote(r.Node.Name), resp.StatusCode)
			} else {
				log.Infof("config: cluster membership confirmed (%d)", resp.StatusCode)
			}
			return &r, nil
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			// Terminal errors (no retry). 404 likely indicates a Portal without cluster endpoints.
			return nil, errors.New(resp.Status)
		case http.StatusTooManyRequests:
			if attempt < maxAttempts {
				log.Debugf("config: join attempt %d/%d rate limited by portal", attempt, maxAttempts)
				time.Sleep(delay)
				continue
			}
			return nil, errors.New(resp.Status)
		case http.StatusConflict, http.StatusBadRequest:
			// Do not retry on 400/409 per spec intent.
			return nil, errors.New(resp.Status)
		default:
			if attempt < maxAttempts {
				log.Debugf("config: join attempt %d/%d received portal status %s", attempt, maxAttempts, resp.Status)
				// TODO: Consider exponential backoff with jitter instead of constant delay.
				time.Sleep(delay)
				continue
			}
			return nil, errors.New(resp.Status)
		}
	}
	return nil, nil
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
	if dns.IsLoopbackHost(host) {
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
		if _, err := c.SaveNodeClientSecret(r.Secrets.ClientSecret); err != nil {
			return fmt.Errorf("failed to persist node client secret: %w", err)
		}
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
			log.Infof("config: applied portal database settings; restart required to connect with new credentials")
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
		log.Debugf("config: jwks prime skipped (%s)", clean.Error(err))
	}
}

// syncNodeTheme downloads or refreshes the Portal-provided theme in the node-specific
// theme directory when the local version is missing or differs from the portal version.
func syncNodeTheme(c *config.Config, portal *url.URL, registerResp *cluster.RegisterResponse) error {
	themeDir := c.NodeThemePath()
	localVersion := strings.TrimSpace(c.NodeThemeVersion())
	hasAppJS := fs.FileExists(filepath.Join(themeDir, fs.AppJsFile))

	portalVersion := ""
	if registerResp != nil {
		portalVersion = clean.TypeUnicode(registerResp.Theme)
	}

	shouldProbe := registerResp == nil

	needsDownload := false
	requiresOverwrite := false

	switch {
	case portalVersion != "":
		if !hasAppJS {
			log.Infof("config: portal theme %s not installed yet; scheduling download", clean.Log(portalVersion))
			needsDownload = true
		} else if localVersion != portalVersion {
			log.Infof("config: portal theme update detected (local %s, portal %s); scheduling download", clean.Log(localVersion), clean.Log(portalVersion))
			needsDownload = true
			requiresOverwrite = true
		} else {
			log.Infof("config: portal theme version %s already installed", clean.Log(localVersion))
		}
	case shouldProbe:
		// Registration failed or was skipped; attempt to obtain the theme when missing.
		needsDownload = !hasAppJS || localVersion == ""
		if needsDownload {
			log.Infof("config: probing portal for theme because local bundle is missing")
		}
	default:
		// Portal responded but has no theme configured; keep existing node theme.
		log.Infof("config: portal did not advertise a theme; skipping download")
		return nil
	}

	if !needsDownload {
		return nil
	}

	// Acquire OAuth bearer via client credentials; skip when credentials are unavailable.
	bearer := ""
	var tokenErr error
	if id, secret := strings.TrimSpace(c.NodeClientID()), strings.TrimSpace(c.NodeClientSecret()); id != "" && secret != "" {
		if t, err := oauthAccessToken(c, portal, id, secret); err != nil {
			tokenErr = err
			log.Infof("config: portal access token request failed (%s)", clean.Error(err))
		} else {
			bearer = t
		}
	}

	if bearer == "" {
		shouldRefresh := false
		if tokenErr != nil && isAuthError(tokenErr) {
			shouldRefresh = true
		} else if registerResp != nil && (strings.TrimSpace(c.NodeClientID()) == "" || strings.TrimSpace(c.NodeClientSecret()) == "") {
			shouldRefresh = true
		}

		if shouldRefresh {
			log.Infof("config: refreshing node credentials for portal theme download")
		}
		if shouldRefresh && refreshNodeCredentials(c, portal) {
			if id, secret := strings.TrimSpace(c.NodeClientID()), strings.TrimSpace(c.NodeClientSecret()); id != "" && secret != "" {
				if t, err := oauthAccessToken(c, portal, id, secret); err == nil {
					bearer = t
				} else {
					log.Infof("config: portal access token retry failed (%s)", clean.Error(err))
				}
			}
		}
	}

	if bearer == "" {
		log.Infof("config: theme sync skipped because no portal credentials are available yet")
		return nil
	}

	endpoint := *portal
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/cluster/theme"

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
		if err = fs.MkdirAll(c.TempPath()); err != nil {
			return err
		}

		zipName := filepath.Join(c.TempPath(), "cluster-theme.zip")
		var out *os.File

		if out, err = os.Create(zipName); err != nil {
			return err
		}

		if _, err = io.Copy(out, resp.Body); err != nil {
			_ = out.Close()
			return err
		}

		_ = out.Close()

		if requiresOverwrite && fs.PathExists(themeDir) {
			if err = os.RemoveAll(themeDir); err != nil {
				return err
			}
		}

		if err = fs.MkdirAll(themeDir); err != nil {
			return err
		}

		_, _, unzipErr := fs.Unzip(zipName, themeDir, 32*fs.MB, 512*fs.MB)
		return unzipErr
	case http.StatusNotFound:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return errors.New(resp.Status)
	default:
		return errors.New(resp.Status)
	}
}

// activateNodeThemeIfPresent switches the active theme path to the node-specific
// directory when a valid cluster-managed theme bundle is available.
func activateNodeThemeIfPresent(c *config.Config) {
	if c == nil {
		return
	}

	// If NodeThemePath() does not exist or does not contain an app.js file,
	// NodeThemeVersion() returns an empty string. No additional checks required.
	if c.NodeThemeVersion() == "" {
		return
	}

	// nodeDir is already clean, because filepath.Join() returns it that way.
	nodeDir := c.NodeThemePath()

	// Return is theme is already activated.
	if filepath.Clean(c.ThemePath()) == nodeDir {
		return
	}

	// Activate cluster theme.
	c.SetThemePath(nodeDir)

	// Report activation.
	log.Debugf("config: activated portal theme from %s", clean.Log(nodeDir))
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

// refreshNodeCredentials rotates the node OAuth client secret using the join token
// and persists the new client ID / secret pair. It returns true when credentials
// were refreshed successfully.
func refreshNodeCredentials(c *config.Config, portal *url.URL) bool {
	if c == nil || portal == nil {
		return false
	}

	joinToken := strings.TrimSpace(c.JoinToken())
	if joinToken == "" {
		log.Infof("config: cannot refresh node credentials without a join token")
		return false
	}

	id, secret, err := obtainNodeCredentialsViaRegister(c, portal, joinToken)
	if err != nil {
		log.Infof("config: failed to refresh node credentials (%s)", clean.Error(err))
		return false
	}

	if _, err = c.SaveNodeClientSecret(secret); err != nil {
		log.Warnf("config: failed to persist node client secret (%s)", clean.Error(err))
		return false
	}

	c.Options().NodeClientID = id
	updates := cluster.OptionsUpdate{}
	if id != "" {
		updates.SetNodeClientID(id)
	}

	if wrote, err := ApplyOptionsUpdate(c, updates); err != nil {
		log.Warnf("config: failed to persist node client id (%s)", clean.Error(err))
	} else if wrote {
		if loadErr := c.Options().Load(c.OptionsYaml()); loadErr != nil {
			log.Warnf("config: failed to reload options after credential refresh (%s)", clean.Error(loadErr))
		}
	}

	return true
}

// isAuthError reports whether the error indicates an authentication failure
// (HTTP 401 or 403) so callers can decide when to refresh credentials.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "401") || strings.Contains(msg, "403")
}

// obtainNodeCredentialsViaRegister calls the portal registration endpoint to
// rotate the node secret and returns the new client ID and secret.
func obtainNodeCredentialsViaRegister(c *config.Config, portal *url.URL, joinToken string) (string, string, error) {
	if portal == nil {
		return "", "", fmt.Errorf("invalid portal url")
	}

	endpoint := *portal
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/cluster/nodes/register"

	payload := cluster.RegisterRequest{
		NodeName:     c.NodeName(),
		NodeRole:     c.NodeRole(),
		RotateSecret: true,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+joinToken)

	resp, err := newHTTPClient(cluster.BootstrapRegisterTimeout).Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusConflict:
		var regResp cluster.RegisterResponse
		if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
			return "", "", err
		}
		id := regResp.Node.ClientID
		secret := ""
		if regResp.Secrets != nil {
			secret = regResp.Secrets.ClientSecret
		}
		if id == "" || secret == "" {
			return "", "", fmt.Errorf("missing client credentials in response")
		}
		return id, secret, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return "", "", fmt.Errorf("%s", resp.Status)
	default:
		return "", "", fmt.Errorf("%s", resp.Status)
	}
}
