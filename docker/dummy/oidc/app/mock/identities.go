package mock

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// Identity is a test user the dummy provider can authenticate. A client selects one by sending
// its key as login_hint in the authorization request; without a hint, the default identity is used.
type Identity struct {
	Subject           string
	Name              string
	PreferredUsername string
	Nickname          string
	Email             string
	EmailVerified     bool
	PhoneNumber       string
	Groups            []string
	Claims            map[string]any
}

// DefaultSubject is the subject of the identity used when a request has no login_hint.
const DefaultSubject = "sub00000001"

// identities maps each login_hint to the test user it selects. Keys and usernames avoid the
// account fixtures of the main repository, except "alice", which deliberately collides with one.
var identities = map[string]Identity{
	"": {
		Subject:           DefaultSubject,
		Name:              "Test",
		PreferredUsername: "prefname",
		Nickname:          "testnick",
		Email:             "test@example.com",
		EmailVerified:     true,
		PhoneNumber:       "0791234567",
	},
	// olivia is a member of the admin group.
	"olivia": {
		Subject:           "sub00000002",
		Name:              "Olivia Admin",
		PreferredUsername: "olivia",
		Email:             "olivia@example.com",
		EmailVerified:     true,
		Groups:            []string{"photoprism-admin", "staff"},
	},
	// oscar is a member of the user group.
	"oscar": {
		Subject:           "sub00000003",
		Name:              "Oscar User",
		PreferredUsername: "oscar",
		Email:             "oscar@example.com",
		EmailVerified:     true,
		Groups:            []string{"photoprism-user"},
	},
	// otto belongs to no group and has an unverified email address.
	"otto": {
		Subject:           "sub00000004",
		Name:              "Otto Unverified",
		PreferredUsername: "otto",
		Email:             "otto@example.com",
		EmailVerified:     false,
	},
	// ola has a verified email address in another domain.
	"ola": {
		Subject:           "sub00000005",
		Name:              "Ola Other",
		PreferredUsername: "ola",
		Email:             "ola@example.org",
		EmailVerified:     true,
	},
	// odette has too many groups to list, so the provider refers to another claim source.
	"odette": {
		Subject:           "sub00000006",
		Name:              "Odette Overage",
		PreferredUsername: "odette",
		Email:             "odette@example.com",
		EmailVerified:     true,
		Claims: map[string]any{
			"_claim_names":   map[string]any{"groups": "src1"},
			"_claim_sources": map[string]any{"src1": map[string]any{"endpoint": "https://graph.example.com/groups"}},
		},
	},
	// petra carries the role a PhotoPrism Portal grants as an identity provider.
	"petra": {
		Subject:           "sub00000007",
		Name:              "Petra Portal",
		PreferredUsername: "petra",
		Email:             "petra@example.com",
		EmailVerified:     true,
		Claims:            map[string]any{"pp_issuer_kind": "portal", "pp_role": "manager"},
	},
	// alice has the name of a local account in the main repository's fixtures.
	"alice": {
		Subject:           "sub00000008",
		Name:              "Alice Conflict",
		PreferredUsername: "alice",
		Email:             "alice.oidc@example.com",
		EmailVerified:     true,
	},
	// nobody has no claim a username can be derived from.
	"nobody": {
		Subject:       "sub00000009",
		Email:         "nobody@example.com",
		EmailVerified: false,
	},
}

// IdentityByHint returns the identity selected by a login_hint, or an error if the hint is unknown.
func IdentityByHint(hint string) (Identity, error) {
	if id, ok := identities[strings.ToLower(strings.TrimSpace(hint))]; ok {
		return id, nil
	}

	return Identity{}, fmt.Errorf("unknown login_hint %q, use one of %s", hint, strings.Join(LoginHints(), ", "))
}

// IdentityBySubject returns the identity with the specified subject, or the default identity.
func IdentityBySubject(subject string) Identity {
	for _, id := range identities {
		if id.Subject == subject {
			return id
		}
	}

	return identities[""]
}

// LoginHints returns the sorted login_hint values that select a non-default identity.
func LoginHints() []string {
	hints := make([]string, 0, len(identities))

	for hint := range identities {
		if hint != "" {
			hints = append(hints, hint)
		}
	}

	sort.Strings(hints)

	return hints
}

// SetUserInfo populates a userinfo response with the claims of the identity.
func (id Identity) SetUserInfo(info *oidc.UserInfo) {
	info.Subject = id.Subject
	info.Name = id.Name
	info.PreferredUsername = id.PreferredUsername
	info.Nickname = id.Nickname
	info.Email = id.Email
	info.EmailVerified = oidc.Bool(id.EmailVerified)

	if id.PhoneNumber != "" {
		info.PhoneNumber = id.PhoneNumber
		info.PhoneNumberVerified = oidc.Bool(true)
	}

	if len(id.Groups) > 0 {
		info.AppendClaims("groups", id.Groups)
	}

	for k, v := range id.Claims {
		info.AppendClaims(k, v)
	}

	info.AppendClaims("private_claim", "test")
}
