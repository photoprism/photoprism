package oidc

import (
	"strings"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// validPrompts lists the OpenID Connect "prompt" values PhotoPrism accepts on the
// authorization request. PromptNone ("none") is intentionally excluded: it tells the
// provider to never show UI and to error when interaction is required, which would
// break the interactive sign-in this option exists to unblock.
var validPrompts = map[string]bool{
	oidc.PromptLogin:         true,
	oidc.PromptConsent:       true,
	oidc.PromptSelectAccount: true,
}

// ParsePrompt splits a space-separated OIDC "prompt" configuration value into its
// recognized, lowercased values and the unsupported tokens that were ignored.
func ParsePrompt(s string) (valid, invalid []string) {
	for _, p := range strings.Fields(s) {
		if v := strings.ToLower(p); validPrompts[v] {
			valid = append(valid, v)
		} else {
			invalid = append(invalid, p)
		}
	}

	return valid, invalid
}
