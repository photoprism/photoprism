package safe

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
)

// Fetch reads the response body of a URL into memory with the same address checks as Download, and returns
// it with the response headers. The timeout covers the whole request, including name resolution. With
// NoRedirects, a redirect is an error. Unlike Download, it ignores the PHOTOPRISM_HTTP_* overrides.
func Fetch(rawURL string, opt *Options) ([]byte, http.Header, error) {
	u, err := URL(rawURL)

	if err != nil {
		return nil, nil, err
	}

	o := Options{Timeout: defaultTimeout, MaxSizeBytes: defaultMaxSize, AllowPrivate: true, Accept: "*/*"}

	if opt != nil {
		if opt.Timeout > 0 {
			o.Timeout = opt.Timeout
		}
		if opt.MaxSizeBytes > 0 {
			o.MaxSizeBytes = opt.MaxSizeBytes
		}
		o.AllowPrivate = opt.AllowPrivate
		o.NoRedirects = opt.NoRedirects
		if strings.TrimSpace(opt.Accept) != "" {
			o.Accept = opt.Accept
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()

	// Check the target address when private networks are disallowed.
	if !o.AllowPrivate {
		if err = checkHostContext(ctx, u.Hostname()); err != nil {
			return nil, nil, err
		}
	}

	client := &http.Client{
		Timeout:       o.Timeout,
		Transport:     newTransport(o.AllowPrivate),
		CheckRedirect: checkRedirect(o.AllowPrivate),
	}

	if o.NoRedirects {
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return ErrRedirect }
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", o.Accept)

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
		return nil, nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, &StatusError{Code: resp.StatusCode}
	}

	// Validate the connected peer address when private ranges are disallowed; an unknown peer is refused.
	if !o.AllowPrivate && disallowedPeer(finalIP) {
		return nil, nil, ErrPrivateIP
	}

	if resp.ContentLength > 0 && resp.ContentLength > o.MaxSizeBytes {
		return nil, nil, ErrSizeExceeded
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, o.MaxSizeBytes+1))

	if err != nil {
		return nil, nil, err
	} else if int64(len(body)) > o.MaxSizeBytes {
		return nil, nil, ErrSizeExceeded
	}

	return body, resp.Header, nil
}
