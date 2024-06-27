package mock

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"gopkg.in/square/go-jose.v2"

	"github.com/caos/oidc/pkg/oidc"
	"github.com/caos/oidc/pkg/op"
)

type AuthStorage struct {
	key *rsa.PrivateKey
	kid string
}

func NewAuthStorage() op.Storage {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	b := make([]byte, 16)
	rand.Read(b)

	return &AuthStorage{
		key: key,
		kid: string(b),
	}
}

type AuthRequest struct {
	ID            string
	ResponseType  oidc.ResponseType
	RedirectURI   string
	Nonce         string
	ClientID      string
	CodeChallenge *oidc.CodeChallenge
}

func (a *AuthRequest) GetACR() string {
	return ""
}

func (a *AuthRequest) GetAMR() []string {
	return []string{}
}

func (a *AuthRequest) GetAudience() []string {
	return []string{
		a.ClientID,
	}
}

func (a *AuthRequest) GetAuthTime() time.Time {
	return time.Now().UTC()
}

func (a *AuthRequest) GetClientID() string {
	return a.ClientID
}

func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge {
	fmt.Println("GetCodeChallenge: ", a.CodeChallenge.Challenge, a.CodeChallenge.Method)
	return a.CodeChallenge
}

func (a *AuthRequest) GetID() string {
	return a.ID
}

func (a *AuthRequest) GetNonce() string {
	return a.Nonce
}

func (a *AuthRequest) GetRedirectURI() string {
	return a.RedirectURI
	// return "http://localhost:5556/auth/callback"
}

func (a *AuthRequest) GetResponseType() oidc.ResponseType {
	return a.ResponseType
}

func (a *AuthRequest) GetScopes() []string {
	return []string{
		"openid",
		"profile",
		"email",
	}
}

func (a *AuthRequest) SetCurrentScopes(scopes []string) {}

func (a *AuthRequest) GetState() string {
	return state
}

func (a *AuthRequest) GetSubject() string {
	return "sub00000001"
}

func (a *AuthRequest) Done() bool {
	return true
}

var (
	a     = &AuthRequest{}
	t     bool
	c     string
	state string
)

func (s *AuthStorage) Health(ctx context.Context) error {
	return nil
}

func (s *AuthStorage) CreateAuthRequest(_ context.Context, authReq *oidc.AuthRequest, userId string) (op.AuthRequest, error) {
	fmt.Println("Userid: ", userId)
	fmt.Println("CreateAuthRequest ID: ", authReq.ID)
	fmt.Println("CreateAuthRequest CodeChallenge: ", authReq.CodeChallenge)
	fmt.Println("CreateAuthRequest CodeChallengeMethod: ", authReq.CodeChallengeMethod)
	fmt.Println("CreateAuthRequest State: ", authReq.State)
	fmt.Println("CreateAuthRequest ClientID: ", authReq.ClientID)
	fmt.Println("CreateAuthRequest ResponseType: ", authReq.ResponseType)
	fmt.Println("CreateAuthRequest Nonce: ", authReq.Nonce)
	fmt.Println("CreateAuthRequest Scopes: ", authReq.Scopes)
	fmt.Println("CreateAuthRequest Display: ", authReq.Display)
	fmt.Println("CreateAuthRequest LoginHint: ", authReq.LoginHint)
	fmt.Println("CreateAuthRequest IDTokenHint: ", authReq.IDTokenHint)
	a = &AuthRequest{ID: "authReqUserAgentId", ClientID: authReq.ClientID, ResponseType: authReq.ResponseType, Nonce: authReq.Nonce, RedirectURI: authReq.RedirectURI}
	if authReq.CodeChallenge != "" {
		a.CodeChallenge = &oidc.CodeChallenge{
			Challenge: authReq.CodeChallenge,
			Method:    authReq.CodeChallengeMethod,
		}
	}
	state = authReq.State
	t = false
	return a, nil
}
func (s *AuthStorage) AuthRequestByCode(_ context.Context, code string) (op.AuthRequest, error) {
	if code != c {
		return nil, errors.New("invalid code")
	}
	return a, nil
}
func (s *AuthStorage) SaveAuthCode(_ context.Context, id, code string) error {
	if a.ID != id {
		return errors.New("SaveAuthCode: not found")
	}
	c = code
	return nil
}
func (s *AuthStorage) DeleteAuthRequest(context.Context, string) error {
	t = true
	return nil
}
func (s *AuthStorage) AuthRequestByID(_ context.Context, id string) (op.AuthRequest, error) {
	fmt.Println("AuthRequestByID: ", id)
	if id != "authReqUserAgentId:usertoken" || t {
		return nil, errors.New("AuthRequestByID: not found")
	}
	return a, nil
}
func (s *AuthStorage) CreateAccessToken(ctx context.Context, request op.TokenRequest) (string, time.Time, error) {
	return "loginId", time.Now().UTC().Add(5 * time.Minute), nil
}
func (s *AuthStorage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshToken string, expiration time.Time, err error) {
	return "loginId", "refreshToken", time.Now().UTC().Add(5 * time.Minute), nil
}
func (s *AuthStorage) TokenRequestByRefreshToken(ctx context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
	if refreshToken != c {
		return nil, errors.New("invalid token")
	}
	return a, nil
}

func (s *AuthStorage) TerminateSession(_ context.Context, userID, clientID string) error {
	return nil
}
func (s *AuthStorage) GetSigningKey(_ context.Context, keyCh chan<- jose.SigningKey) {
	//keyCh <- jose.SigningKey{Algorithm: jose.RS256, Key: s.key}
	keyCh <- jose.SigningKey{Algorithm: jose.RS256, Key: &jose.JSONWebKey{Key: s.key, KeyID: s.kid}}
}
func (s *AuthStorage) GetKey(_ context.Context) (*rsa.PrivateKey, error) {
	return s.key, nil
}
func (s *AuthStorage) GetKeySet(_ context.Context) (*jose.JSONWebKeySet, error) {
	pubkey := s.key.Public()

	wrongkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       wrongkey.Public(),
				Use:       "sig",
				Algorithm: "RS256",
				KeyID:     "wrongkey0002",
			}, {
				Key:       pubkey,
				Use:       "sig",
				Algorithm: "RS256",
				KeyID:     s.kid,
			},
		},
	}, nil
}
func (s *AuthStorage) GetKeyByIDAndUserID(_ context.Context, _, _ string) (*jose.JSONWebKey, error) {
	pubkey := s.key.Public()

	return &jose.JSONWebKey{
		Key:       pubkey,
		Use:       "sig",
		Algorithm: "RS256",
		KeyID:     s.kid,
	}, nil
}

func (s *AuthStorage) GetClientByClientID(_ context.Context, id string) (op.Client, error) {
	if id == "none" {
		return nil, errors.New("GetClientByClientID: not found")
	}
	var appType op.ApplicationType
	var authMethod oidc.AuthMethod
	var accessTokenType op.AccessTokenType
	var responseTypes []oidc.ResponseType
	var grantTypes = []oidc.GrantType{
		oidc.GrantTypeCode,
	}
	if id == "web" {
		appType = op.ApplicationTypeWeb
		authMethod = oidc.AuthMethodBasic
		accessTokenType = op.AccessTokenTypeBearer
		responseTypes = []oidc.ResponseType{oidc.ResponseTypeCode}
	} else if id == "native" {
		appType = op.ApplicationTypeNative
		authMethod = oidc.AuthMethodBasic
		accessTokenType = op.AccessTokenTypeBearer
		responseTypes = []oidc.ResponseType{oidc.ResponseTypeCode, oidc.ResponseTypeIDToken, oidc.ResponseTypeIDTokenOnly}
	} else {
		appType = op.ApplicationTypeUserAgent
		authMethod = oidc.AuthMethodNone
		accessTokenType = op.AccessTokenTypeJWT
		responseTypes = []oidc.ResponseType{oidc.ResponseTypeIDToken, oidc.ResponseTypeIDTokenOnly}
	}
	return &ConfClient{ID: id, applicationType: appType, authMethod: authMethod, accessTokenType: accessTokenType, responseTypes: responseTypes, grantTypes: grantTypes, devMode: true}, nil
}

func (s *AuthStorage) AuthorizeClientIDSecret(_ context.Context, id string, _ string) error {
	return nil
}

func (s *AuthStorage) SetUserinfoFromToken(ctx context.Context, userinfo oidc.UserInfoSetter, _, _, _ string) error {
	return s.SetUserinfoFromScopes(ctx, userinfo, "", "", []string{})
}
func (s *AuthStorage) SetUserinfoFromScopes(ctx context.Context, userinfo oidc.UserInfoSetter, _, _ string, _ []string) error {
	userinfo.SetSubject(a.GetSubject())
	//userinfo.SetAddress(oidc.NewUserInfoAddress("Test 789\nPostfach 2", "", "", "", "", ""))
	userinfo.SetEmail("test@example.com", true)
	userinfo.SetPhone("0791234567", true)
	userinfo.SetName("Test")
	userinfo.AppendClaims("private_claim", "test")
	userinfo.SetNickname("testnick")
	userinfo.SetPreferredUsername("prefname")
	return nil
}
func (s *AuthStorage) GetPrivateClaimsFromScopes(_ context.Context, _, _ string, _ []string) (map[string]interface{}, error) {
	return map[string]interface{}{"private_claim": "test"}, nil
}

func (s *AuthStorage) SetIntrospectionFromToken(ctx context.Context, introspect oidc.IntrospectionResponse, tokenID, subject, clientID string) error {
	if err := s.SetUserinfoFromScopes(ctx, introspect, "", "", []string{}); err != nil {
		return err
	}
	introspect.SetClientID(a.ClientID)
	return nil
}

func (s *AuthStorage) ValidateJWTProfileScopes(ctx context.Context, userID string, scope []string) ([]string, error) {
	return scope, nil
}

type ConfClient struct {
	applicationType op.ApplicationType
	authMethod      oidc.AuthMethod
	responseTypes   []oidc.ResponseType
	grantTypes      []oidc.GrantType
	ID              string
	accessTokenType op.AccessTokenType
	devMode         bool
}

func (c *ConfClient) GetID() string {
	return c.ID
}
func (c *ConfClient) RedirectURIs() []string {
	return []string{
		"https://registered.com/callback",
		"http://localhost:9999/callback",
		"http://localhost:5556/auth/callback",
		"custom://callback",
		"https://localhost:8443/test/a/instructions-example/callback",
		"https://op.certification.openid.net:62064/authz_cb",
		"https://op.certification.openid.net:62064/authz_post",
		"http://localhost:2342/api/v1/oidc/redirect",
		"https://app.localssl.dev/api/v1/oidc/redirect",
	}
}
func (c *ConfClient) PostLogoutRedirectURIs() []string {
	return []string{}
}

func (c *ConfClient) LoginURL(id string) string {
	//return "authorize/callback?id=" + id
	return "login?id=" + id
}

func (c *ConfClient) ApplicationType() op.ApplicationType {
	return c.applicationType
}

func (c *ConfClient) AuthMethod() oidc.AuthMethod {
	return c.authMethod
}

func (c *ConfClient) IDTokenLifetime() time.Duration {
	return 60 * time.Minute
}
func (c *ConfClient) AccessTokenType() op.AccessTokenType {
	return c.accessTokenType
}
func (c *ConfClient) ResponseTypes() []oidc.ResponseType {
	return c.responseTypes
}
func (c *ConfClient) GrantTypes() []oidc.GrantType {
	return c.grantTypes
}

func (c *ConfClient) DevMode() bool {
	return c.devMode
}

func (c *ConfClient) AllowedScopes() []string {
	return nil
}

func (c *ConfClient) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *ConfClient) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *ConfClient) IsScopeAllowed(scope string) bool {
	return false
}

func (c *ConfClient) IDTokenUserinfoClaimsAssertion() bool {
	return false
}

func (c *ConfClient) ClockSkew() time.Duration {
	return 0
}
