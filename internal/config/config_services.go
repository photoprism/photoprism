package config

import "strings"

// ServicesCIDR returns the configured CIDR allowlist for outbound service requests.
func (c *Config) ServicesCIDR() string {
	return strings.TrimSpace(c.options.ServicesCIDR)
}
