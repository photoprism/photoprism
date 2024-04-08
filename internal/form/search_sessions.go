package form

import "github.com/photoprism/photoprism/pkg/authn"

// SearchSessions represents a session search form.
type SearchSessions struct {
	Query    string `form:"q"`
	UID      string `form:"uid"`
	Provider string `form:"provider"`
	Method   string `form:"method"`
	Count    int    `form:"count" binding:"required" serialize:"-"`
	Offset   int    `form:"offset" serialize:"-"`
	Order    string `form:"order" serialize:"-"`
}

// AuthProviders returns the normalized authentication provider types.
func (f *SearchSessions) AuthProviders() []authn.ProviderType {
	return authn.Providers(f.Provider)
}

// AuthMethods returns the normalized authentication method types.
func (f *SearchSessions) AuthMethods() []authn.MethodType {
	return authn.Methods(f.Method)
}

// GetQuery returns the query string.
func (f *SearchSessions) GetQuery() string {
	return f.Query
}

// SetQuery sets the query string.
func (f *SearchSessions) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString parses the query string into form fields.
func (f *SearchSessions) ParseQueryString() error {
	return ParseQueryString(f)
}
