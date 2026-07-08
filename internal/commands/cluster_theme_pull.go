package commands

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// ClusterThemePullCommand downloads the Portal theme and installs it.
var ClusterThemePullCommand = &cli.Command{
	Name:  "theme",
	Usage: "Theme subcommands",
	Subcommands: []*cli.Command{
		{
			Name:  "pull",
			Usage: "Downloads the theme from a portal and installs it in config/theme or the dest path. If only a join token is provided, this command first registers the node to obtain client credentials, then downloads the theme (no extra command needed).",
			Flags: []cli.Flag{
				&cli.PathFlag{Name: "dest", Usage: "extract destination `PATH` (defaults to config/theme)", Value: ""},
				&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "replace existing files at destination"},
				&cli.StringFlag{Name: "portal-url", Usage: "Portal base `URL` (defaults to global config)"},
				&cli.StringFlag{Name: "join-token", Usage: "Portal access `TOKEN` (defaults to global config)"},
				&cli.StringFlag{Name: "client-id", Usage: "Node client `ID` (defaults to NodeClientID from config)"},
				&cli.StringFlag{Name: "client-secret", Usage: "Node client `SECRET` (defaults to NodeClientSecret from config)"},
				// JSON output supported via report.CliFlags on parent command where applicable
			},
			Action: clusterThemePullAction,
		},
	},
}

func clusterThemePullAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		portalURL := ""

		if ctx.IsSet("portal-url") && ctx.String("portal-url") != "" {
			portalURL = strings.TrimRight(ctx.String("portal-url"), "/")
		}

		if portalURL == "" {
			portalURL = strings.TrimRight(conf.PortalUrl(), "/")
		}

		if portalURL == "" {
			portalURL = strings.TrimRight(os.Getenv(config.EnvVar("portal-url")), "/")
		}

		if portalURL == "" {
			if domain := strings.TrimSpace(conf.ClusterDomain()); domain != "" {
				portalURL = fmt.Sprintf("https://portal.%s", domain)
			}
		}

		if portalURL == "" {
			return fmt.Errorf("portal-url not configured; set --portal-url or PHOTOPRISM_PORTAL_URL")
		}

		// Credentials: prefer OAuth client credentials (client-id/secret), fallback to join-token for compatibility.
		clientID := ctx.String("client-id")

		if clientID == "" {
			clientID = conf.NodeClientID()
		}

		clientSecret := ctx.String("client-secret")

		if clientSecret == "" {
			clientSecret = conf.NodeClientSecret()
		}

		token := ""

		if clientID != "" && clientSecret != "" {
			// OAuth client_credentials
			t, err := obtainOAuthToken(portalURL, clientID, clientSecret)
			if err != nil {
				log.Warnf("cluster: oauth token failed, falling back to join token (%s)", clean.Error(err))
			} else {
				token = t
			}
		}
		if token == "" {
			// Try join-token assisted path. If NodeClientID/NodeClientSecret not available, attempt register to obtain them, then OAuth.
			jt := ctx.String("join-token")

			if jt == "" {
				jt = conf.JoinToken()
			}

			if jt == "" {
				jt = os.Getenv(config.EnvVar("join-token"))
			}

			if jt != "" && (clientID == "" || clientSecret == "") {
				if id, sec, err := obtainClientCredentialsViaRegister(portalURL, jt, conf.NodeName()); err == nil {
					if t, err := obtainOAuthToken(portalURL, id, sec); err == nil {
						token = t
					}
				}
			}

			if token == "" {
				return fmt.Errorf("authentication required: provide --client-id/--client-secret or a join token to obtain credentials")
			}
		}

		dest := ctx.Path("dest")
		if dest == "" {
			dest = conf.ThemePath()
		}
		dest = fs.Abs(dest)

		// Destination must be a directory. Create if needed.
		if fi, err := os.Stat(dest); err == nil && !fi.IsDir() {
			return fmt.Errorf("destination is a file, expected a directory: %s", clean.Log(dest))
		} else if err != nil {
			if err = fs.MkdirAll(dest); err != nil {
				return err
			}
		}

		// If destination contains files and --force not set, refuse.
		if !ctx.Bool("force") {
			if nonEmpty, _ := dirNonEmpty(dest); nonEmpty {
				return fmt.Errorf("destination is not empty; use --force to replace existing files: %s", clean.Log(dest))
			}
		} else {
			// Clean destination contents, but keep the directory itself.
			if err := removeDirContents(dest); err != nil {
				return err
			}
		}

		// Download zip to a temp file.
		zipURL := portalURL + "/api/v1/cluster/theme"

		// TODO: Enforce TLS for non-local Portal URLs (similar to bootstrap) unless an explicit
		// insecure override is provided. Consider adding a --tls-only / --insecure flag.
		tmpFile, err := os.CreateTemp("", "photoprism-theme-*.zip")

		if err != nil {
			return err
		}

		defer func() {
			_ = os.Remove(tmpFile.Name()) //nolint:gosec // tmpFile path is generated by os.CreateTemp.
		}()

		// zipURL is built from the validated portal URL configured for this CLI command.
		req, err := http.NewRequest(http.MethodGet, zipURL, nil) //nolint:gosec

		if err != nil {
			return err
		}

		header.SetAuthorization(req, token)
		req.Header.Set(header.Accept, header.ContentTypeZip)

		// Use a short timeout for responsiveness; align with bootstrap defaults.
		client := &http.Client{Timeout: cluster.BootstrapRegisterTimeout}
		resp, err := client.Do(req) //nolint:gosec
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Debugf("cluster: %s (close theme response body)", clean.Error(closeErr))
			}
		}()
		if resp.StatusCode != http.StatusOK {
			// Map common codes to clearer messages
			switch resp.StatusCode {
			case http.StatusUnauthorized, http.StatusForbidden:
				return cli.Exit(fmt.Errorf("unauthorized; check portal token and permissions (%s)", resp.Status), 4)
			case http.StatusTooManyRequests:
				return cli.Exit(fmt.Errorf("rate limited by portal (%s)", resp.Status), 6)
			case http.StatusNotFound:
				return cli.Exit(fmt.Errorf("portal theme not found (%s)", resp.Status), 3)
			case http.StatusBadRequest:
				return cli.Exit(fmt.Errorf("bad request (%s)", resp.Status), 2)
			default:
				return cli.Exit(fmt.Errorf("download failed: %s", resp.Status), 1)
			}
		}
		if _, err = io.Copy(tmpFile, resp.Body); err != nil {
			return err
		}
		if err := tmpFile.Close(); err != nil {
			return err
		}

		// Extract safely into destination.
		if err := unzipSafe(tmpFile.Name(), dest); err != nil {
			return err
		}

		if ctx.Bool("json") {
			fmt.Printf("{\"installed\":\"%s\"}\n", clean.Log(dest))
		} else {
			log.Infof("installed theme files to %s", clean.Log(dest))
			fmt.Println(dest)
		}
		return nil
	})
}

// obtainOAuthToken requests an access token via client_credentials using Basic auth.
func obtainOAuthToken(portalURL, clientID, clientSecret string) (string, error) {
	u, err := url.Parse(strings.TrimRight(portalURL, "/"))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid portal-url: %s", portalURL)
	}
	tokenURL := *u
	tokenURL.Path = strings.TrimRight(tokenURL.Path, "/") + "/api/v1/oauth/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	// tokenURL is derived from a parsed portal URL with explicit scheme/host validation above.
	req, _ := http.NewRequest(http.MethodPost, tokenURL.String(), strings.NewReader(form.Encode())) //nolint:gosec
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	basic := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+basic)

	client := &http.Client{Timeout: cluster.BootstrapRegisterTimeout}
	resp, err := client.Do(req) //nolint:gosec

	if err != nil {
		return "", err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Debugf("cluster: %s (close oauth response body)", clean.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", resp.Status)
	}

	var tok map[string]any

	if err = json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", err
	}

	accessToken, _ := tok["access_token"].(string)

	if accessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}

	return accessToken, nil
}

func dirNonEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)

	if err != nil {
		return false, err
	}

	for range entries {
		// Ignore typical dotfiles? Keep it simple: any entry counts
		return true, nil
	}

	return false, nil
}

func removeDirContents(dir string) error {
	entries, err := os.ReadDir(dir)

	if err != nil {
		return err
	}

	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}

	return nil
}

func unzipSafe(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)

	if err != nil {
		return err
	}

	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			log.Debugf("cluster: %s (close theme archive)", clean.Error(closeErr))
		}
	}()

	destAbs, err := filepath.Abs(dest)

	if err != nil {
		return err
	}

	// Empty theme archives are valid; install succeeds without files.
	for _, f := range r.File {
		// ZIP names are always slash-separated regardless of host OS.
		if strings.Contains(f.Name, `\`) {
			continue
		}

		name := path.Clean(strings.TrimSpace(f.Name))

		if name == "." || name == ".." || strings.HasPrefix(name, "../") || path.IsAbs(name) || strings.Contains(name, ":") {
			continue
		}

		// Skip entries that look like hidden files or directories.
		if fs.FileNameHidden(path.Base(name)) {
			continue
		}

		nameOS := filepath.FromSlash(name)
		target := filepath.Join(destAbs, nameOS)
		targetAbs, targetErr := filepath.Abs(target)

		if targetErr != nil {
			continue
		}

		rel, relErr := filepath.Rel(destAbs, targetAbs)

		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			if err = fs.MkdirAll(targetAbs); err != nil {
				return err
			}
			continue
		}

		// Ensure parent exists
		if err = fs.MkdirAll(filepath.Dir(targetAbs)); err != nil {
			return err
		}

		// Open for read.
		rc, err := f.Open()

		if err != nil {
			return err
		}

		// Create/truncate target.
		out, err := os.OpenFile(targetAbs, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode()) //nolint:gosec // paths derived from zip entries validated earlier

		if err != nil {
			_ = rc.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc) //nolint:gosec // zip entries size is bounded by upstream
		closeOutErr := out.Close()
		closeRcErr := rc.Close()

		if copyErr != nil {
			return copyErr
		}

		if closeOutErr != nil {
			return closeOutErr
		}

		if closeRcErr != nil {
			return closeRcErr
		}
	}
	return nil
}
