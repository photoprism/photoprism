package authn

// Authentication providers.
const (
	ProviderDefault = ""
	ProviderNone    = "none"
	ProviderToken   = "token"
	ProviderLocal   = "local"
	ProviderLDAP    = "ldap"
)

// ProviderString returns the provider name as a string for use in logs and reports.
func ProviderString(s string) string {
	if s == ProviderDefault {
		return "default"
	}

	return s
}
