package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"caos-test-op/mock"
)

const (
	testProxyHost   = "dummy-oidc.localssl.dev"
	testRedirectURI = "https://app.localssl.dev/api/v1/oidc/redirect"
)

// testFlow drives the provider through requests that appear to come from a TLS-terminating proxy.
type testFlow struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
}

// newTestFlow starts the dummy provider on a local test server.
func newTestFlow(t *testing.T) *testFlow {
	t.Helper()

	handler, err := newHandler(defaultIssuer, mock.NewAuthStorage())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &testFlow{
		t:      t,
		server: server,
		client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

// do sends a request for the path and query of target, with the headers a proxy adds for HTTPS.
func (f *testFlow) do(method, target string, body url.Values, header http.Header) *http.Response {
	f.t.Helper()

	u, err := url.Parse(target)
	if err != nil {
		f.t.Fatalf("parse %s: %v", target, err)
	}

	var req *http.Request

	if body != nil {
		req, err = http.NewRequest(method, f.server.URL+u.RequestURI(), strings.NewReader(body.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, f.server.URL+u.RequestURI(), nil)
	}

	if err != nil {
		f.t.Fatalf("new request: %v", err)
	}

	for k, v := range header {
		req.Header[k] = v
	}

	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", testProxyHost)

	resp, err := f.client.Do(req)
	if err != nil {
		f.t.Fatalf("%s %s: %v", method, target, err)
	}

	f.t.Cleanup(func() { _ = resp.Body.Close() })

	return resp
}

// location returns the redirect target of a response, failing the test if there is none.
func (f *testFlow) location(resp *http.Response) *url.URL {
	f.t.Helper()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		f.t.Fatalf("expected redirect, got %d", resp.StatusCode)
	}

	u, err := resp.Location()
	if err != nil {
		f.t.Fatalf("location: %v", err)
	}

	return u
}

// authorize starts an authorization request with the specified login_hint and returns the
// redirect back to the client.
func (f *testFlow) authorize(loginHint string) *url.URL {
	f.t.Helper()

	q := url.Values{
		"client_id":     {"photoprism-develop"},
		"redirect_uri":  {testRedirectURI},
		"response_type": {"code"},
		"scope":         {"openid profile email"},
		"state":         {"state1"},
		"nonce":         {"nonce1"},
	}

	if loginHint != "" {
		q.Set("login_hint", loginHint)
	}

	next := f.location(f.do(http.MethodGet, "/authorize?"+q.Encode(), nil, nil))

	// Follow the login and callback redirects until the provider returns to the client.
	for i := 0; i < 3 && !strings.HasPrefix(next.String(), testRedirectURI); i++ {
		next = f.location(f.do(http.MethodGet, next.String(), nil, nil))
	}

	if !strings.HasPrefix(next.String(), testRedirectURI) {
		f.t.Fatalf("expected redirect to client, got %s", next)
	}

	return next
}

// tokens exchanges an authorization code and returns the decoded token response.
func (f *testFlow) tokens(code string) map[string]any {
	f.t.Helper()

	header := http.Header{}
	header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("photoprism-develop:secret")))

	resp := f.do(http.MethodPost, "/oauth/token", url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {testRedirectURI},
	}, header)

	if resp.StatusCode != http.StatusOK {
		f.t.Fatalf("token exchange returned %d", resp.StatusCode)
	}

	var result map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		f.t.Fatalf("decode tokens: %v", err)
	}

	return result
}

// jwtClaims returns the payload of a JWT without verifying its signature.
func jwtClaims(t *testing.T, token string) map[string]any {
	t.Helper()

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid jwt")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode jwt: %v", err)
	}

	var claims map[string]any

	if err = json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("unmarshal jwt: %v", err)
	}

	return claims
}

func TestDiscoveryIssuer(t *testing.T) {
	f := newTestFlow(t)

	t.Run("BehindHttpsProxy", func(t *testing.T) {
		var doc map[string]any

		if err := json.NewDecoder(f.do(http.MethodGet, "/.well-known/openid-configuration", nil, nil).Body).Decode(&doc); err != nil {
			t.Fatalf("decode discovery: %v", err)
		}

		if doc["issuer"] != "https://"+testProxyHost {
			t.Errorf("issuer = %v", doc["issuer"])
		}

		if doc["authorization_endpoint"] != "https://"+testProxyHost+"/authorize" {
			t.Errorf("authorization_endpoint = %v", doc["authorization_endpoint"])
		}
	})
	t.Run("Direct", func(t *testing.T) {
		resp, err := http.Get(f.server.URL + "/.well-known/openid-configuration")
		if err != nil {
			t.Fatalf("get discovery: %v", err)
		}
		defer resp.Body.Close()

		var doc map[string]any

		if err = json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			t.Fatalf("decode discovery: %v", err)
		}

		if doc["issuer"] != defaultIssuer {
			t.Errorf("issuer = %v", doc["issuer"])
		}
	})
}

func TestCodeFlow(t *testing.T) {
	t.Run("LoginHint", func(t *testing.T) {
		f := newTestFlow(t)
		redirect := f.authorize("olivia")

		if got := redirect.Query().Get("state"); got != "state1" {
			t.Fatalf("state = %q", got)
		}

		tokens := f.tokens(redirect.Query().Get("code"))
		claims := jwtClaims(t, tokens["id_token"].(string))

		if claims["iss"] != "https://"+testProxyHost {
			t.Errorf("iss = %v", claims["iss"])
		}

		if claims["sub"] != "sub00000002" || claims["preferred_username"] != "olivia" || claims["nonce"] != "nonce1" {
			t.Errorf("unexpected id token claims: %v", claims)
		}

		groups, _ := claims["groups"].([]any)
		if len(groups) != 2 || groups[0] != "photoprism-admin" {
			t.Errorf("groups = %v", claims["groups"])
		}

		header := http.Header{}
		header.Set("Authorization", "Bearer "+tokens["access_token"].(string))

		var info map[string]any

		if err := json.NewDecoder(f.do(http.MethodGet, "/userinfo", nil, header).Body).Decode(&info); err != nil {
			t.Fatalf("decode userinfo: %v", err)
		}

		if info["sub"] != "sub00000002" || info["email"] != "olivia@example.com" {
			t.Errorf("unexpected userinfo: %v", info)
		}
	})
	t.Run("DefaultIdentity", func(t *testing.T) {
		f := newTestFlow(t)
		claims := jwtClaims(t, f.tokens(f.authorize("").Query().Get("code"))["id_token"].(string))

		if claims["sub"] != mock.DefaultSubject || claims["preferred_username"] != "prefname" {
			t.Errorf("unexpected id token claims: %v", claims)
		}

		if _, ok := claims["groups"]; ok {
			t.Errorf("default identity must not have groups: %v", claims["groups"])
		}
	})
	t.Run("UnknownLoginHint", func(t *testing.T) {
		f := newTestFlow(t)
		q := url.Values{
			"client_id":     {"photoprism-develop"},
			"redirect_uri":  {testRedirectURI},
			"response_type": {"code"},
			"scope":         {"openid"},
			"state":         {"state1"},
			"login_hint":    {"zz-unknown"},
		}

		redirect := f.location(f.do(http.MethodGet, "/authorize?"+q.Encode(), nil, nil))

		if !strings.HasPrefix(redirect.String(), testRedirectURI) {
			t.Fatalf("expected error redirect to client, got %s", redirect)
		}

		if got := redirect.Query().Get("error"); got != "invalid_request" {
			t.Errorf("error = %q", got)
		}

		if redirect.Query().Get("code") != "" {
			t.Errorf("unexpected code for unknown login_hint")
		}
	})
}

func TestRequestIssuer(t *testing.T) {
	issuerFunc, err := requestIssuer(defaultIssuer)(true)
	if err != nil {
		t.Fatalf("request issuer: %v", err)
	}

	tests := []struct {
		name  string
		host  string
		proto string
		fwd   string
		want  string
	}{
		{"Direct", "dummy-oidc:9998", "", "", defaultIssuer},
		{"ForwardedHttp", "dummy-oidc:9998", "http", testProxyHost, defaultIssuer},
		{"ForwardedHttps", "dummy-oidc:9998", "https", testProxyHost, "https://" + testProxyHost},
		{"ForwardedHttpsUppercase", "dummy-oidc:9998", "HTTPS", testProxyHost, "https://" + testProxyHost},
		{"HostFallback", testProxyHost, "https", "", "https://" + testProxyHost},
		{"InvalidHost", "dummy-oidc:9998", "https", "evil.example/path", defaultIssuer},
		{"UserInfoHost", "dummy-oidc:9998", "https", "user@evil.example", defaultIssuer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Host = tt.host

			if tt.proto != "" {
				r.Header.Set("X-Forwarded-Proto", tt.proto)
			}

			if tt.fwd != "" {
				r.Header.Set("X-Forwarded-Host", tt.fwd)
			}

			if got := issuerFunc(r); got != tt.want {
				t.Errorf("issuer = %q, want %q", got, tt.want)
			}
		})
	}
}
