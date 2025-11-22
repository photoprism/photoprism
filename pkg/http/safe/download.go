package safe

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
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

	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return ErrSchemeNotAllowed
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

	// Optional SSRF block
	if !o.AllowPrivate {
		if ip := net.ParseIP(u.Hostname()); ip != nil {
			if isPrivateOrDisallowedIP(ip) {
				return ErrPrivateIP
			}
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			addrs, lookErr := net.DefaultResolver.LookupIPAddr(ctx, u.Hostname())
			if lookErr != nil {
				return lookErr
			}
			for _, a := range addrs {
				if isPrivateOrDisallowedIP(a.IP) {
					return ErrPrivateIP
				}
			}
		}
	}

	// Enforce redirect validation when private networks are disallowed.
	client := &http.Client{
		Timeout: o.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !o.AllowPrivate {
				h := req.URL.Hostname()
				if ip := net.ParseIP(h); ip != nil {
					if isPrivateOrDisallowedIP(ip) {
						return ErrPrivateIP
					}
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					addrs, lookErr := net.DefaultResolver.LookupIPAddr(ctx, h)
					if lookErr != nil {
						return lookErr
					}
					for _, a := range addrs {
						if isPrivateOrDisallowedIP(a.IP) {
							return ErrPrivateIP
						}
					}
				}
			}
			// Propagate Accept header from the first request.
			if len(via) > 0 {
				if v := via[0].Header.Get("Accept"); v != "" {
					req.Header.Set("Accept", v)
				}
			}
			return nil
		},
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

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	// Validate the connected peer address when private ranges are disallowed.
	if !o.AllowPrivate && finalIP != nil && isPrivateOrDisallowedIP(finalIP) {
		return ErrPrivateIP
	}

	if resp.ContentLength > 0 && o.MaxSizeBytes > 0 && resp.ContentLength > o.MaxSizeBytes {
		return ErrSizeExceeded
	}

	tmp := destPath + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
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
	if err = os.Rename(tmp, destPath); err != nil {
		return err
	}
	return nil
}

func isPrivateOrDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 10 {
			return true
		}
		if v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31 {
			return true
		}
		if v4[0] == 192 && v4[1] == 168 {
			return true
		}
		if v4[0] == 169 && v4[1] == 254 {
			return true
		}
		return false
	}
	// IPv6 ULA fc00::/7
	if ip.To16() != nil {
		if ip[0]&0xFE == 0xFC {
			return true
		}
	}
	return false
}
