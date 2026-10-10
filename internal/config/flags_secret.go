package config

import (
	"slices"
	"strings"
)

// SecretFlagPatterns lists the name fragments that mark a flag as carrying credential material.
var SecretFlagPatterns = []string{"password", "token", "secret", "key", "salt", "jwks", "dsn"}

// SecretFlagExceptions lists flags whose names match SecretFlagPatterns but that name a location,
// a lifetime, or a length rather than a credential.
var SecretFlagExceptions = []string{
	"download-token-maxage",
	"jwks-cache-ttl",
	"jwks-url",
	"password-length",
	"tls-key",
}

// FlagsWithoutSecret returns the names of flags that match SecretFlagPatterns but do not set Secret,
// skipping SecretFlagExceptions and the specified exceptions, and the number of flags that matched.
func FlagsWithoutSecret(flags CliFlags, exceptions ...string) (missing []string, matched int) {
	for _, flag := range flags {
		name := flag.Name()

		if slices.Contains(SecretFlagExceptions, name) || slices.Contains(exceptions, name) {
			continue
		}

		lowerName, lowerEnv := strings.ToLower(name), strings.ToLower(flag.EnvVar())

		if !slices.ContainsFunc(SecretFlagPatterns, func(p string) bool {
			return strings.Contains(lowerName, p) || strings.Contains(lowerEnv, p)
		}) {
			continue
		}

		matched++

		if !flag.Secret {
			missing = append(missing, name)
		}
	}

	return missing, matched
}
