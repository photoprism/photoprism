package authn

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Generic error messages for authentication and authorization:
var (
	ErrUnauthorized           = errors.New("unauthorized")
	ErrAccountAlreadyExists   = errors.New("account already exists")
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountDisabled        = errors.New("account disabled")
	ErrAccountCreateFailed    = errors.New("failed to create account")
	ErrAccountUpdateFailed    = errors.New("failed to update account")
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidShareToken      = errors.New("invalid share token")
	ErrTokenRequired          = errors.New("token required")
	ErrInvalidToken           = errors.New("invalid token")
	ErrInvalidTokenType       = errors.New("invalid token type")
	ErrInsufficientScope      = errors.New("insufficient scope")
	ErrNameRequired           = errors.New("name required")
	ErrScopeRequired          = errors.New("scope required")
	ErrDisabledInPublicMode   = errors.New("disabled in public mode")
	ErrAuthenticationDisabled = errors.New("authentication disabled")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
)

// OIDC and OAuth2-related error messages:
var (
	ErrInvalidProviderConfiguration = errors.New("invalid provider configuration")
	ErrInvalidGrantType             = errors.New("invalid grant type")
	ErrInvalidClientID              = errors.New("invalid client id")
	ErrInvalidAuthID                = errors.New("invalid auth id")
	ErrAuthProviderIsNotOIDC        = errors.New("auth provider is not oidc")
	ErrAuthCodeRequired             = errors.New("auth code required")
	ErrClientIDRequired             = errors.New("client id required")
	ErrInvalidClientSecret          = errors.New("invalid client secret")
	ErrClientSecretRequired         = errors.New("client secret required")
	ErrVerifiedEmailRequired        = errors.New("verified email required")
	ErrRegistrationDisabled         = errors.New("registration disabled")
)

// User-related error messages:
var (
	ErrUsernameRequired           = errors.New("username required")
	ErrUsernameRequiredToRegister = errors.New("username required to register")
	ErrInvalidUsername            = errors.New("invalid username")
	ErrUsernameDoesNotMatch       = errors.New("specified username does not match")
)

// Passcode-related error messages:
var (
	ErrPasscodeRequired           = errors.New("passcode required")
	ErrPasscodeNotSetUp           = errors.New("passcode required, but not configured")
	ErrPasscodeNotVerified        = errors.New("passcode not verified")
	ErrPasscodeAlreadyActivated   = errors.New("passcode already activated")
	ErrPasscodeGenerateFailed     = errors.New("failed to generate passcode")
	ErrPasscodeCreateFailed       = errors.New("failed to create passcode")
	ErrPasscodeSaveFailed         = errors.New("failed to save passcode")
	ErrPasscodeVerificationFailed = errors.New("failed to verify passcode")
	ErrPasscodeActivationFailed   = errors.New("failed to activate passcode")
	ErrPasscodeDeactivationFailed = errors.New("failed to deactivate passcode")
	ErrPasscodeNotSupported       = errors.New("passcode not supported")
	ErrInvalidPasscode            = errors.New("invalid passcode")
	ErrInvalidPasscodeFormat      = errors.New("invalid passcode format")
	ErrInvalidPasscodeKey         = errors.New("invalid passcode key")
	ErrInvalidPasscodeType        = errors.New("invalid passcode type")
)

// Password-related error messages:
var (
	ErrInvalidPassword     = errors.New("invalid password")
	ErrPasswordRequired    = errors.New("password required")
	ErrPasswordTooShort    = errors.New("password is too short")
	ErrPasswordTooLong     = errors.New(fmt.Sprintf("password must have less than %d characters", txt.ClipPassword))
	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
)

// WebDAV-related error messages:
var (
	ErrWebDAVAccessDisabled     = errors.New("webdav access is disabled")
	ErrFailedToCreateUploadPath = errors.New("failed to create upload path")
)
