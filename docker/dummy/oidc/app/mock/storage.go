// Package mock provides an in-memory OIDC storage used by the dummy provider in development.
package mock

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"sync"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// Compile-time assertions that AuthStorage satisfies the v3 op.Storage interface.
var _ op.Storage = (*AuthStorage)(nil)

// AuthStorage is the in-memory storage backing the dummy OIDC provider.
// It implements the op.Storage interface required by zitadel/oidc/v3.
type AuthStorage struct {
	lock          sync.Mutex
	signingKey    signingKey
	authRequests  map[string]*AuthRequest
	codes         map[string]string
	tokens        map[string]*tokenEntry
	refreshTokens map[string]*refreshTokenEntry
}

// AuthRequest captures the subset of an OIDC authorization request that the
// dummy provider needs to satisfy the op.AuthRequest interface.
type AuthRequest struct {
	ID            string
	ClientID      string
	Subject       string
	Scopes        []string
	RedirectURI   string
	ResponseType  oidc.ResponseType
	ResponseMode  oidc.ResponseMode
	Nonce         string
	State         string
	CodeChallenge *oidc.CodeChallenge

	authTime time.Time
	done     bool
}

type signingKey struct {
	id        string
	algorithm jose.SignatureAlgorithm
	key       *rsa.PrivateKey
}

// SignatureAlgorithm returns the signing algorithm used by the key.
func (s *signingKey) SignatureAlgorithm() jose.SignatureAlgorithm { return s.algorithm }

// Key returns the private key used for signing.
func (s *signingKey) Key() any { return s.key }

// ID returns the key identifier.
func (s *signingKey) ID() string { return s.id }

type publicKey struct{ signingKey }

// ID returns the key identifier.
func (p *publicKey) ID() string { return p.id }

// Algorithm returns the JOSE algorithm of the key.
func (p *publicKey) Algorithm() jose.SignatureAlgorithm { return p.algorithm }

// Use returns the key usage; the dummy only signs.
func (p *publicKey) Use() string { return "sig" }

// Key returns the public part of the signing key.
func (p *publicKey) Key() any { return &p.key.PublicKey }

type tokenEntry struct {
	id            string
	applicationID string
	subject       string
	audience      []string
	scopes        []string
	expiration    time.Time
	refreshTokens string
}

type refreshTokenEntry struct {
	id            string
	applicationID string
	subject       string
	audience      []string
	scopes        []string
	authTime      time.Time
	amr           []string
	accessToken   string
	expiration    time.Time
}

// NewAuthStorage creates an AuthStorage with a fresh RSA signing key.
func NewAuthStorage() *AuthStorage {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return &AuthStorage{
		signingKey: signingKey{
			id:        uuid.NewString(),
			algorithm: jose.RS256,
			key:       key,
		},
		authRequests:  make(map[string]*AuthRequest),
		codes:         make(map[string]string),
		tokens:        make(map[string]*tokenEntry),
		refreshTokens: make(map[string]*refreshTokenEntry),
	}
}

// MarkRequestDone flags an auth request as authenticated so the OP can issue a code.
func (s *AuthStorage) MarkRequestDone(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	req, ok := s.authRequests[id]
	if !ok {
		return errors.New("auth request not found")
	}
	req.done = true
	req.authTime = time.Now().UTC()
	return nil
}

// GetID returns the auth request id.
func (a *AuthRequest) GetID() string { return a.ID }

// GetACR returns the (unset) authentication context class reference.
func (a *AuthRequest) GetACR() string { return "" }

// GetAMR returns the authentication methods that were used.
func (a *AuthRequest) GetAMR() []string {
	if a.done {
		return []string{"pwd"}
	}
	return nil
}

// GetAudience returns the audience for issued tokens.
func (a *AuthRequest) GetAudience() []string { return []string{a.ClientID} }

// GetAuthTime returns the time the user authenticated.
func (a *AuthRequest) GetAuthTime() time.Time { return a.authTime }

// GetClientID returns the client identifier.
func (a *AuthRequest) GetClientID() string { return a.ClientID }

// GetCodeChallenge returns the PKCE code challenge if any.
func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge { return a.CodeChallenge }

// GetNonce returns the request nonce.
func (a *AuthRequest) GetNonce() string { return a.Nonce }

// GetRedirectURI returns the client redirect URI.
func (a *AuthRequest) GetRedirectURI() string { return a.RedirectURI }

// GetResponseType returns the requested response_type.
func (a *AuthRequest) GetResponseType() oidc.ResponseType { return a.ResponseType }

// GetResponseMode returns the response_mode (defaulting to query).
func (a *AuthRequest) GetResponseMode() oidc.ResponseMode {
	if a.ResponseMode != "" {
		return a.ResponseMode
	}
	return oidc.ResponseModeQuery
}

// GetScopes returns the scopes granted to the request.
func (a *AuthRequest) GetScopes() []string {
	if len(a.Scopes) > 0 {
		return a.Scopes
	}
	return []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail}
}

// GetState returns the state parameter that the client provided.
func (a *AuthRequest) GetState() string { return a.State }

// GetSubject returns the authenticated user's subject identifier.
func (a *AuthRequest) GetSubject() string { return a.Subject }

// Done reports whether the request has completed authentication.
func (a *AuthRequest) Done() bool { return a.done }

// Health implements op.Storage.
func (s *AuthStorage) Health(_ context.Context) error { return nil }

// CreateAuthRequest stores a new request and returns it as op.AuthRequest.
func (s *AuthStorage) CreateAuthRequest(_ context.Context, authReq *oidc.AuthRequest, _ string) (op.AuthRequest, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	id := uuid.NewString()
	req := &AuthRequest{
		ID:           id,
		ClientID:     authReq.ClientID,
		Subject:      "sub00000001",
		Scopes:       authReq.Scopes,
		RedirectURI:  authReq.RedirectURI,
		ResponseType: authReq.ResponseType,
		ResponseMode: authReq.ResponseMode,
		Nonce:        authReq.Nonce,
		State:        authReq.State,
	}
	if authReq.CodeChallenge != "" {
		req.CodeChallenge = &oidc.CodeChallenge{
			Challenge: authReq.CodeChallenge,
			Method:    authReq.CodeChallengeMethod,
		}
	}
	s.authRequests[id] = req
	return req, nil
}

// AuthRequestByID looks up an auth request by id.
func (s *AuthStorage) AuthRequestByID(_ context.Context, id string) (op.AuthRequest, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	req, ok := s.authRequests[id]
	if !ok {
		return nil, errors.New("auth request not found")
	}
	return req, nil
}

// AuthRequestByCode looks up an auth request by an issued authorization code.
func (s *AuthStorage) AuthRequestByCode(_ context.Context, code string) (op.AuthRequest, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	id, ok := s.codes[code]
	if !ok {
		return nil, errors.New("code not found")
	}
	req, ok := s.authRequests[id]
	if !ok {
		return nil, errors.New("auth request not found")
	}
	return req, nil
}

// SaveAuthCode stores the code-to-request mapping.
func (s *AuthStorage) SaveAuthCode(_ context.Context, id, code string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.codes[code] = id
	return nil
}

// DeleteAuthRequest removes an auth request and any associated code.
func (s *AuthStorage) DeleteAuthRequest(_ context.Context, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.authRequests, id)
	for code, reqID := range s.codes {
		if reqID == id {
			delete(s.codes, code)
		}
	}
	return nil
}

// CreateAccessToken issues an opaque access token id.
func (s *AuthStorage) CreateAccessToken(_ context.Context, request op.TokenRequest) (string, time.Time, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	tok := s.newAccessTokenLocked(request, "")
	return tok.id, tok.expiration, nil
}

// CreateAccessAndRefreshTokens issues an access token alongside a refresh token.
func (s *AuthStorage) CreateAccessAndRefreshTokens(_ context.Context, request op.TokenRequest, currentRefreshToken string) (string, string, time.Time, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if currentRefreshToken != "" {
		if _, ok := s.refreshTokens[currentRefreshToken]; !ok {
			return "", "", time.Time{}, errors.New("invalid refresh token")
		}
		delete(s.refreshTokens, currentRefreshToken)
	}

	refreshID := uuid.NewString()
	tok := s.newAccessTokenLocked(request, refreshID)

	authTime := time.Now().UTC()
	var amr []string
	if authReq, ok := request.(*AuthRequest); ok {
		authTime = authReq.authTime
		amr = authReq.GetAMR()
	}

	s.refreshTokens[refreshID] = &refreshTokenEntry{
		id:            refreshID,
		applicationID: tok.applicationID,
		subject:       tok.subject,
		audience:      tok.audience,
		scopes:        tok.scopes,
		authTime:      authTime,
		amr:           amr,
		accessToken:   tok.id,
		expiration:    time.Now().UTC().Add(5 * time.Hour),
	}
	return tok.id, refreshID, tok.expiration, nil
}

func (s *AuthStorage) newAccessTokenLocked(request op.TokenRequest, refreshID string) *tokenEntry {
	id := uuid.NewString()
	tok := &tokenEntry{
		id:            id,
		applicationID: clientIDFromRequest(request),
		subject:       request.GetSubject(),
		audience:      request.GetAudience(),
		scopes:        request.GetScopes(),
		expiration:    time.Now().UTC().Add(5 * time.Minute),
		refreshTokens: refreshID,
	}
	s.tokens[id] = tok
	return tok
}

func clientIDFromRequest(request op.TokenRequest) string {
	switch req := request.(type) {
	case *AuthRequest:
		return req.ClientID
	case *refreshTokenRequest:
		return req.applicationID
	}
	return ""
}

// TokenRequestByRefreshToken returns a refresh-token-backed token request.
func (s *AuthStorage) TokenRequestByRefreshToken(_ context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	rt, ok := s.refreshTokens[refreshToken]
	if !ok {
		return nil, errors.New("invalid refresh token")
	}
	return &refreshTokenRequest{
		id:            rt.id,
		applicationID: rt.applicationID,
		subject:       rt.subject,
		audience:      rt.audience,
		scopes:        rt.scopes,
		authTime:      rt.authTime,
		amr:           rt.amr,
	}, nil
}

// TerminateSession removes any tokens for a given user/client.
func (s *AuthStorage) TerminateSession(_ context.Context, userID, clientID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for id, tok := range s.tokens {
		if tok.subject == userID && tok.applicationID == clientID {
			delete(s.tokens, id)
		}
	}
	for id, rt := range s.refreshTokens {
		if rt.subject == userID && rt.applicationID == clientID {
			delete(s.refreshTokens, id)
		}
	}
	return nil
}

// GetRefreshTokenInfo returns the user and token id for a refresh token.
func (s *AuthStorage) GetRefreshTokenInfo(_ context.Context, _, token string) (string, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	rt, ok := s.refreshTokens[token]
	if !ok {
		return "", "", op.ErrInvalidRefreshToken
	}
	return rt.subject, rt.id, nil
}

// RevokeToken removes an access or refresh token.
func (s *AuthStorage) RevokeToken(_ context.Context, tokenOrID, _, clientID string) *oidc.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if tok, ok := s.tokens[tokenOrID]; ok {
		if clientID != "" && tok.applicationID != clientID {
			return oidc.ErrInvalidClient().WithDescription("token was not issued for this client")
		}
		delete(s.tokens, tokenOrID)
		return nil
	}
	if rt, ok := s.refreshTokens[tokenOrID]; ok {
		if clientID != "" && rt.applicationID != clientID {
			return oidc.ErrInvalidClient().WithDescription("token was not issued for this client")
		}
		delete(s.refreshTokens, tokenOrID)
		delete(s.tokens, rt.accessToken)
	}
	return nil
}

// SigningKey returns the static dummy signing key.
func (s *AuthStorage) SigningKey(_ context.Context) (op.SigningKey, error) {
	return &s.signingKey, nil
}

// SignatureAlgorithms reports the algorithms this dummy supports.
func (s *AuthStorage) SignatureAlgorithms(_ context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{s.signingKey.algorithm}, nil
}

// KeySet returns the dummy's public JWKS.
func (s *AuthStorage) KeySet(_ context.Context) ([]op.Key, error) {
	return []op.Key{&publicKey{s.signingKey}}, nil
}

// GetClientByClientID returns a permissive client for any id.
func (s *AuthStorage) GetClientByClientID(_ context.Context, id string) (op.Client, error) {
	if id == "none" {
		return nil, errors.New("client not found")
	}
	return NewClient(id), nil
}

// AuthorizeClientIDSecret accepts any client/secret pair.
func (s *AuthStorage) AuthorizeClientIDSecret(_ context.Context, _, _ string) error {
	return nil
}

// SetUserinfoFromScopes is required by op.Storage but the dummy populates claims via SetUserinfoFromToken.
func (s *AuthStorage) SetUserinfoFromScopes(_ context.Context, info *oidc.UserInfo, _, _ string, _ []string) error {
	fillUserInfo(info)
	return nil
}

// SetUserinfoFromToken populates the userinfo response for a given token.
func (s *AuthStorage) SetUserinfoFromToken(_ context.Context, info *oidc.UserInfo, tokenID, _, _ string) error {
	s.lock.Lock()
	tok, ok := s.tokens[tokenID]
	s.lock.Unlock()
	if !ok {
		return errors.New("token not found")
	}
	if tok.expiration.Before(time.Now().UTC()) {
		return errors.New("token expired")
	}
	fillUserInfo(info)
	return nil
}

// SetIntrospectionFromToken populates the introspection response for a given token.
func (s *AuthStorage) SetIntrospectionFromToken(_ context.Context, response *oidc.IntrospectionResponse, tokenID, _, clientID string) error {
	s.lock.Lock()
	tok, ok := s.tokens[tokenID]
	s.lock.Unlock()
	if !ok {
		return errors.New("token not found")
	}
	if tok.expiration.Before(time.Now().UTC()) {
		return errors.New("token expired")
	}
	for _, aud := range tok.audience {
		if aud == clientID {
			response.Expiration = oidc.FromTime(tok.expiration)
			response.Scope = tok.scopes
			response.ClientID = tok.applicationID
			info := new(oidc.UserInfo)
			fillUserInfo(info)
			response.SetUserInfo(info)
			return nil
		}
	}
	return errors.New("token is not valid for this client")
}

// GetPrivateClaimsFromScopes returns custom claims for a token. The dummy always returns one.
func (s *AuthStorage) GetPrivateClaimsFromScopes(_ context.Context, _, _ string, _ []string) (map[string]any, error) {
	return map[string]any{"private_claim": "test"}, nil
}

// GetKeyByIDAndClientID returns the dummy's public key regardless of clientID.
func (s *AuthStorage) GetKeyByIDAndClientID(_ context.Context, _, _ string) (*jose.JSONWebKey, error) {
	return &jose.JSONWebKey{
		KeyID:     s.signingKey.id,
		Algorithm: string(s.signingKey.algorithm),
		Use:       "sig",
		Key:       s.signingKey.key.Public(),
	}, nil
}

// ValidateJWTProfileScopes accepts the requested scopes verbatim.
func (s *AuthStorage) ValidateJWTProfileScopes(_ context.Context, _ string, scopes []string) ([]string, error) {
	return scopes, nil
}

// fillUserInfo populates the userinfo response with the dummy user identity.
func fillUserInfo(info *oidc.UserInfo) {
	info.Subject = "sub00000001"
	info.Email = "test@example.com"
	info.EmailVerified = oidc.Bool(true)
	info.Name = "Test"
	info.Nickname = "testnick"
	info.PreferredUsername = "prefname"
	info.PhoneNumber = "0791234567"
	info.PhoneNumberVerified = oidc.Bool(true)
	info.AppendClaims("private_claim", "test")
}

// refreshTokenRequest implements op.RefreshTokenRequest for the dummy.
type refreshTokenRequest struct {
	id            string
	applicationID string
	subject       string
	audience      []string
	scopes        []string
	authTime      time.Time
	amr           []string
}

// GetAMR returns the authentication methods recorded with the refresh token.
func (r *refreshTokenRequest) GetAMR() []string { return r.amr }

// GetAudience returns the token audience.
func (r *refreshTokenRequest) GetAudience() []string { return r.audience }

// GetAuthTime returns the original auth time.
func (r *refreshTokenRequest) GetAuthTime() time.Time { return r.authTime }

// GetClientID returns the client_id the token was issued to.
func (r *refreshTokenRequest) GetClientID() string { return r.applicationID }

// GetScopes returns the scopes that the refresh token grants.
func (r *refreshTokenRequest) GetScopes() []string { return r.scopes }

// GetSubject returns the subject identifier.
func (r *refreshTokenRequest) GetSubject() string { return r.subject }

// SetCurrentScopes mutates the active scopes when the OP narrows them.
func (r *refreshTokenRequest) SetCurrentScopes(scopes []string) { r.scopes = scopes }
