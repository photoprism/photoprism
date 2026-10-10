package safe

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Download fetches a URL to a destination file with timeouts, size limits, and optional SSRF protection.
func Download(destPath, rawURL string, opt *Options) error {
	if destPath == "" {
		return errors.New("invalid destination path")
	}
	// Prepare destination directory.
	if dir := filepath.Dir(destPath); dir == "" || dir == "/" || dir == "." || dir == ".." {
		return errors.New("invalid destination directory")
	} else if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	u, err := URL(rawURL)
	if err != nil {
		return err
	}

	// Defaults w/ env overrides
	maxSize := defaultMaxSize
	if n := envInt64("PHOTOPRISM_HTTP_MAX_DOWNLOAD"); n > 0 {
		maxSize = n
	}
	timeout := defaultTimeout
	if d := envDuration("PHOTOPRISM_HTTP_TIMEOUT"); d > 0 {
		timeout = d
	}

	o := Options{Timeout: timeout, MaxSizeBytes: maxSize, AllowPrivate: true, Accept: "*/*"}
	if opt != nil {
		if opt.Timeout > 0 {
			o.Timeout = opt.Timeout
		}
		if opt.MaxSizeBytes > 0 {
			o.MaxSizeBytes = opt.MaxSizeBytes
		}
		o.AllowPrivate = opt.AllowPrivate
		if strings.TrimSpace(opt.Accept) != "" {
			o.Accept = opt.Accept
		}
	}

	// Check the target address when private networks are disallowed.
	if !o.AllowPrivate {
		if err = checkHost(u.Hostname()); err != nil {
			return err
		}
	}

	// Enforce redirect validation when private networks are disallowed.
	client := &http.Client{
		Timeout:       o.Timeout,
		Transport:     newTransport(o.AllowPrivate),
		CheckRedirect: checkRedirect(o.AllowPrivate),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	if o.Accept != "" {
		req.Header.Set("Accept", o.Accept)
	}
	// Capture the final remote IP used for the connection.
	var finalIP net.IP
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			if addr := info.Conn.RemoteAddr(); addr != nil {
				host, _, _ := net.SplitHostPort(addr.String())
				if ip := net.ParseIP(host); ip != nil {
					finalIP = ip
				}
			}
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// #nosec G704 URL is parsed and validated (scheme, host, optional private-IP checks).
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	// Validate the connected peer address when private ranges are disallowed; an unknown peer is refused.
	if !o.AllowPrivate && disallowedPeer(finalIP) {
		return ErrPrivateIP
	}

	if resp.ContentLength > 0 && o.MaxSizeBytes > 0 && resp.ContentLength > o.MaxSizeBytes {
		return ErrSizeExceeded
	}

	tmp := destPath + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:gosec // destPath validated by caller; temp file
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(tmp)
		}
	}()

	var r io.Reader = resp.Body
	if o.MaxSizeBytes > 0 {
		r = io.LimitReader(resp.Body, o.MaxSizeBytes+1)
	}
	n, copyErr := io.Copy(f, r)
	if copyErr != nil {
		err = copyErr
		return err
	}
	if o.MaxSizeBytes > 0 && n > o.MaxSizeBytes {
		err = ErrSizeExceeded
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, destPath)
}

// maxRedirects is the number of requests after which a download stops following redirects.
const maxRedirects = 10

// checkRedirect returns the redirect policy of a download, which checks each target address when
// private networks are disallowed and keeps the Accept header of the first request.
func checkRedirect(allowPrivate bool) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		} else if !allowPrivate {
			if err := checkHost(req.URL.Hostname()); err != nil {
				return err
			}
		}

		if len(via) > 0 {
			if v := via[0].Header.Get("Accept"); v != "" {
				req.Header.Set("Accept", v)
			}
		}

		return nil
	}
}

// newTransport returns the default transport or, when private networks are disallowed, a copy of it
// without keep-alives that refuses a connection to a disallowed address before it is opened.
func newTransport(allowPrivate bool) http.RoundTripper {
	if allowPrivate {
		return http.DefaultTransport
	}

	base, ok := http.DefaultTransport.(*http.Transport)

	if !ok {
		base = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}

	transport := base.Clone()
	transport.DisableKeepAlives = true
	transport.DialTLS = nil //nolint:staticcheck // SA1019: a deprecated dialer that is still used if set
	transport.DialTLSContext = nil
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second, Control: dialControlFunc}
	transport.DialContext = dialer.DialContext

	return transport
}
