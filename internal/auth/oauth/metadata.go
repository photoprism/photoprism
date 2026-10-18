package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/photoprism/photoprism/pkg/http/safe"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Bounds of client metadata documents and their cache.
const (
	MetadataMaxBytes        = 16 * 1024
	MetadataTimeout         = 5 * time.Second
	MetadataMaxRedirects    = 20
	MetadataCacheMaxItems   = 256
	MetadataCacheMinAge     = 5 * time.Minute
	MetadataCacheMaxAge     = 24 * time.Hour
	MetadataCacheDefaultAge = time.Hour
	MetadataCacheStaleAge   = 24 * time.Hour
)

// Errors returned when a client metadata document cannot be used.
var (
	ErrInvalidMetadata     = errors.New("invalid client metadata document")
	ErrMetadataUnavailable = errors.New("client metadata document unavailable")
	ErrMetadataRateLimited = errors.New("client metadata document fetch rate limit exceeded")
)

// ClientMetadata represents the fields of a client metadata document that are used.
type ClientMetadata struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
}

// FetchMetadata fetches a client metadata document with metadataFetchOptions; tests may replace it.
var FetchMetadata = func(clientID string) ([]byte, http.Header, error) {
	return safe.Fetch(clientID, metadataFetchOptions())
}

// metadataFetchOptions returns the fetch options for client metadata documents, which follow no redirects
// and reach no private network addresses.
func metadataFetchOptions() *safe.Options {
	return &safe.Options{
		Timeout:      MetadataTimeout,
		MaxSizeBytes: MetadataMaxBytes,
		AllowPrivate: false,
		Accept:       "application/json",
		NoRedirects:  true,
	}
}

// metadataEntry is a cached client metadata document.
type metadataEntry struct {
	client  *Client
	expires time.Time
}

// metadataCache holds the clients of fetched metadata documents by client ID.
var metadataCache = struct {
	sync.Mutex
	items map[string]metadataEntry
}{items: make(map[string]metadataEntry)}

// FindMetadataClient returns the app client described by the metadata document at the client ID URL,
// from the cache if possible. The caller must check that the client ID may be fetched. Before a fetch,
// lookup.AllowFetch is asked, so the caller can charge a rate limit only when the document is not cached.
// With lookup.AllowStale, an expired entry is used for up to MetadataCacheStaleAge while it cannot be fetched.
func FindMetadataClient(clientID string, lookup ClientLookup) (*Client, error) {
	now := time.Now()

	metadataCache.Lock()
	entry, found := metadataCache.items[clientID]
	metadataCache.Unlock()

	stale := lookup.AllowStale && found && now.Before(entry.expires.Add(MetadataCacheStaleAge))

	if found && now.Before(entry.expires) {
		return entry.client, nil
	} else if lookup.AllowFetch != nil && !lookup.AllowFetch() {
		if stale {
			return entry.client, nil
		}

		return nil, ErrMetadataRateLimited
	}

	body, header, err := FetchMetadata(clientID)

	if err != nil {
		if withdrawnMetadata(err) {
			uncacheMetadataClient(clientID)
		} else if stale {
			return entry.client, nil
		}

		return nil, fmt.Errorf("%w: %s", ErrMetadataUnavailable, err)
	}

	client, err := ParseClientMetadata(clientID, body)

	if mediaType, _, _ := mime.ParseMediaType(header.Get("Content-Type")); err == nil && mediaType != "application/json" && !strings.HasSuffix(mediaType, "+json") {
		err = ErrInvalidMetadata
	}

	// A document that is no longer valid replaces the cached one.
	if err != nil {
		uncacheMetadataClient(clientID)
		return nil, err
	}

	cacheMetadataClient(clientID, client, now.Add(metadataCacheAge(header)))

	return client, nil
}

// ParseClientMetadata validates a client metadata document fetched from the client ID URL and returns the
// app client it describes.
func ParseClientMetadata(clientID string, body []byte) (*Client, error) {
	var doc ClientMetadata

	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, ErrInvalidMetadata
	}

	switch {
	case doc.ClientID != clientID:
		return nil, ErrInvalidMetadata
	case len(doc.RedirectURIs) == 0 || len(doc.RedirectURIs) > MetadataMaxRedirects:
		return nil, ErrInvalidMetadata
	case doc.TokenEndpointAuthMethod != "" && doc.TokenEndpointAuthMethod != "none":
		return nil, ErrInvalidClient
	case len(doc.GrantTypes) > 0 && !slices.Contains(doc.GrantTypes, "authorization_code"):
		return nil, ErrInvalidMetadata
	case len(doc.ResponseTypes) > 0 && !slices.Contains(doc.ResponseTypes, "code"):
		return nil, ErrInvalidMetadata
	}

	for _, uri := range doc.RedirectURIs {
		if !ValidValue(uri, MaxRedirectURILength) || !ValidRedirectURI(uri) {
			return nil, ErrInvalidMetadata
		}
	}

	return &Client{
		ID:           clientID,
		Name:         metadataClientName(doc.ClientName, clientID),
		Scope:        metadataClientScope(doc.Scope),
		RedirectURIs: doc.RedirectURIs,
		Metadata:     true,
	}, nil
}

// metadataClientName returns the client name without control, format, filler or line separator characters,
// clipped to 200 characters, or the client ID host if the name has no letter or digit or contains a URL.
func metadataClientName(name, clientID string) string {
	name = strings.Join(strings.FieldsFunc(name, func(r rune) bool { return invalidRune(r) || fillerRune(r) || unicode.IsSpace(r) }), " ")

	if name = txt.Clip(name, 200); strings.IndexFunc(name, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) >= 0 && !strings.Contains(name, "://") {
		return name
	}

	return strings.TrimSuffix(strings.TrimPrefix(scheme.OriginURL(clientID), "https://"), "/")
}

// fillerRune reports whether r is a filler or blank character that renders as empty space.
func fillerRune(r rune) bool {
	switch r {
	case '\u115F', '\u1160', '\u2800', '\u3164', '\uFFA0':
		return true
	}

	return false
}

// metadataClientScope returns the scopes a client metadata document allows: those of AppScopes it names,
// or all of them if it names none.
func metadataClientScope(scope string) string {
	if scope = strings.TrimSpace(scope); scope == "" {
		return JoinScopes(AppScopes)
	}

	result := make([]string, 0, len(AppScopes))

	for _, s := range strings.Fields(scope) {
		if slices.Contains(AppScopes, s) && !slices.Contains(result, s) {
			result = append(result, s)
		}
	}

	return JoinScopes(result)
}

// metadataCacheAge returns how long a document may be cached based on its Cache-Control max-age,
// bounded by MetadataCacheMinAge and MetadataCacheMaxAge.
func metadataCacheAge(header http.Header) time.Duration {
	age := MetadataCacheDefaultAge

	noStore := false

	for _, directive := range strings.Split(header.Get("Cache-Control"), ",") {
		directive = strings.ToLower(strings.TrimSpace(directive))

		if directive == "no-store" || directive == "no-cache" {
			noStore = true
		} else if v, found := strings.CutPrefix(directive, "max-age="); found {
			if seconds, err := strconv.Atoi(v); err == nil {
				age = time.Duration(seconds) * time.Second
			}
		}
	}

	if noStore {
		return MetadataCacheMinAge
	}

	return min(max(age, MetadataCacheMinAge), MetadataCacheMaxAge)
}

// cacheMetadataClient stores a client in the bounded cache, removing entries past their stale age and
// then the one that expires first if the cache is full.
func cacheMetadataClient(clientID string, client *Client, expires time.Time) {
	metadataCache.Lock()
	defer metadataCache.Unlock()

	now := time.Now()

	for id, entry := range metadataCache.items {
		if !now.Before(entry.expires.Add(MetadataCacheStaleAge)) {
			delete(metadataCache.items, id)
		}
	}

	for len(metadataCache.items) >= MetadataCacheMaxItems {
		var first string

		for id, entry := range metadataCache.items {
			if first == "" || entry.expires.Before(metadataCache.items[first].expires) {
				first = id
			}
		}

		delete(metadataCache.items, first)
	}

	metadataCache.items[clientID] = metadataEntry{client: client, expires: expires}
}

// withdrawnMetadata reports whether a fetch error means the document was removed rather than unavailable,
// i.e. the server answered with a client error other than 408 or 429.
func withdrawnMetadata(err error) bool {
	var statusErr *safe.StatusError

	if !errors.As(err, &statusErr) {
		return false
	}

	return statusErr.Code >= 400 && statusErr.Code < 500 && statusErr.Code != http.StatusRequestTimeout && statusErr.Code != http.StatusTooManyRequests
}

// uncacheMetadataClient removes a client from the cache.
func uncacheMetadataClient(clientID string) {
	metadataCache.Lock()
	delete(metadataCache.items, clientID)
	metadataCache.Unlock()
}

// FlushMetadataCache removes all cached client metadata documents.
func FlushMetadataCache() {
	metadataCache.Lock()
	clear(metadataCache.items)
	metadataCache.Unlock()
}
