package mock

import (
	"context"
	"sort"
	"testing"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestIdentityByHint(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		id, err := IdentityByHint("")
		if err != nil || id.Subject != DefaultSubject {
			t.Fatalf("got %v, %v", id.Subject, err)
		}
	})
	t.Run("CaseAndSpace", func(t *testing.T) {
		id, err := IdentityByHint(" Olivia ")
		if err != nil || id.PreferredUsername != "olivia" {
			t.Fatalf("got %v, %v", id.PreferredUsername, err)
		}
	})
	t.Run("Unknown", func(t *testing.T) {
		if _, err := IdentityByHint("zz-unknown"); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestIdentityBySubject(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		if id := IdentityBySubject("sub00000003"); id.PreferredUsername != "oscar" {
			t.Fatalf("got %q", id.PreferredUsername)
		}
	})
	t.Run("FallsBackToDefault", func(t *testing.T) {
		if id := IdentityBySubject("unknown"); id.Subject != DefaultSubject {
			t.Fatalf("got %q", id.Subject)
		}
	})
}

func TestIdentities(t *testing.T) {
	t.Run("UniqueSubjects", func(t *testing.T) {
		seen := make(map[string]string)

		for hint, id := range identities {
			if other, ok := seen[id.Subject]; ok {
				t.Errorf("%q and %q share subject %s", hint, other, id.Subject)
			}

			seen[id.Subject] = hint
		}
	})
}

func TestLoginHints(t *testing.T) {
	hints := LoginHints()

	if len(hints) != len(identities)-1 {
		t.Fatalf("got %d hints for %d identities", len(hints), len(identities))
	}

	if !sort.StringsAreSorted(hints) {
		t.Errorf("hints not sorted: %v", hints)
	}

	for _, h := range hints {
		if h == "" {
			t.Error("default identity must not be listed")
		}
	}
}

func TestIdentity_SetUserInfo(t *testing.T) {
	t.Run("GroupsAndClaims", func(t *testing.T) {
		info := &oidc.UserInfo{}
		identities["olivia"].SetUserInfo(info)

		if info.Subject != "sub00000002" || !bool(info.EmailVerified) {
			t.Fatalf("unexpected userinfo: %+v", info)
		}

		groups, _ := info.Claims["groups"].([]string)
		if len(groups) != 2 {
			t.Errorf("groups = %v", info.Claims["groups"])
		}
	})
	t.Run("UnverifiedWithoutGroups", func(t *testing.T) {
		info := &oidc.UserInfo{}
		identities["otto"].SetUserInfo(info)

		if bool(info.EmailVerified) {
			t.Error("email must not be verified")
		}

		if _, ok := info.Claims["groups"]; ok {
			t.Error("unexpected groups claim")
		}
	})
	t.Run("PortalRole", func(t *testing.T) {
		info := &oidc.UserInfo{}
		identities["petra"].SetUserInfo(info)

		if info.Claims["pp_issuer_kind"] != "portal" || info.Claims["pp_role"] != "manager" {
			t.Errorf("unexpected claims: %v", info.Claims)
		}
	})
}

func TestCreateAuthRequest_LoginHint(t *testing.T) {
	s := NewAuthStorage()

	t.Run("SelectsIdentity", func(t *testing.T) {
		req, err := s.CreateAuthRequest(context.Background(), &oidc.AuthRequest{ClientID: "c", LoginHint: "oscar"}, "")
		if err != nil {
			t.Fatal(err)
		}

		if req.GetSubject() != "sub00000003" {
			t.Errorf("subject = %q", req.GetSubject())
		}
	})
	t.Run("UnknownHint", func(t *testing.T) {
		if _, err := s.CreateAuthRequest(context.Background(), &oidc.AuthRequest{ClientID: "c", LoginHint: "zz"}, ""); err == nil {
			t.Error("expected error")
		}
	})
}
