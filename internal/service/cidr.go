package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var errServiceHostNotAllowed = errors.New("service host is not allowed by services-cidr")

// ParseCIDRs parses comma-separated CIDR ranges and IPs.
func ParseCIDRs(raw string) (result []*net.IPNet, err error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return result, nil
	}

	for _, token := range strings.Split(raw, ",") {
		token = strings.TrimSpace(token)

		if token == "" {
			continue
		}

		if ip := net.ParseIP(token); ip != nil {
			result = append(result, singleIPNet(ip))
			continue
		}

		if _, network, parseErr := net.ParseCIDR(token); parseErr != nil {
			return nil, fmt.Errorf("invalid services-cidr value %q", token)
		} else {
			result = append(result, network)
		}
	}

	return result, nil
}

// singleIPNet returns the narrowest CIDR network for a single IP.
func singleIPNet(ip net.IP) *net.IPNet {
	if ip == nil {
		return nil
	}

	if ip = ip.To4(); ip != nil {
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
	}

	if ip = ip.To16(); ip != nil {
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
	}

	return nil
}

// IPAllowed reports whether an IP is allowed by the configured CIDR rules.
func IPAllowed(ip net.IP, cidrs []*net.IPNet) bool {
	if ip == nil || ip.IsUnspecified() {
		return false
	}

	if len(cidrs) == 0 {
		return true
	}

	if ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	for _, network := range cidrs {
		if network != nil && network.Contains(ip) {
			return true
		}
	}

	return false
}

// ValidateURLHost validates a URL host against the configured CIDR allowlist.
func ValidateURLHost(u *url.URL, cidrs []*net.IPNet, timeout time.Duration) error {
	if u == nil {
		return errServiceHostNotAllowed
	}

	host := strings.TrimSpace(u.Hostname())

	if host == "" {
		return errServiceHostNotAllowed
	}

	if ip := net.ParseIP(host); ip != nil {
		if IPAllowed(ip, cidrs) {
			return nil
		}

		return errServiceHostNotAllowed
	}

	if len(cidrs) == 0 {
		return nil
	}

	ctx := context.Background()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)

	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if IPAllowed(addr.IP, cidrs) {
			return nil
		}
	}

	return errServiceHostNotAllowed
}

// NewHTTPClient creates an outbound HTTP client with optional CIDR restrictions.
func NewHTTPClient(timeout time.Duration, cidrs []*net.IPNet) *http.Client {
	dialer := &net.Dialer{}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				conn, err := dialer.DialContext(ctx, network, addr)

				if err != nil {
					return nil, err
				}

				host, _, splitErr := net.SplitHostPort(conn.RemoteAddr().String())

				if splitErr != nil || !IPAllowed(net.ParseIP(host), cidrs) {
					_ = conn.Close()
					return nil, errServiceHostNotAllowed
				}

				return conn, nil
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return ValidateURLHost(req.URL, cidrs, 5*time.Second)
		},
	}
}
