package cluster

import "time"

// BootstrapAutoJoinEnabled indicates whether cluster bootstrap logic is enabled
// for nodes by default. Portal nodes ignore this value; gating is decided by
// runtime checks (e.g., conf.Portal() and conf.NodeRole()).
var BootstrapAutoJoinEnabled = true

// BootstrapAutoThemeEnabled indicates whether bootstrap should attempt to
// download and install a Portal-provided theme when appropriate.
var BootstrapAutoThemeEnabled = true

// BootstrapRegisterMaxAttempts defines how many attempts the bootstrap logic
// makes when contacting the Portal for registration before giving up.
var BootstrapRegisterMaxAttempts = 6

// BootstrapRegisterRetryDelay defines the delay between registration attempts
// when the Portal is temporarily unavailable.
var BootstrapRegisterRetryDelay = 15 * time.Second

// BootstrapRegisterTimeout defines the HTTP client timeout per registration
// request to the Portal.
var BootstrapRegisterTimeout = 15 * time.Second

// BootstrapThemeInstallOnlyIfMissingJS ensures theme installation only happens
// when the local theme directory is missing or does not contain an app.js file.
var BootstrapThemeInstallOnlyIfMissingJS = true

// BootstrapAllowThemeOverwrite indicates whether bootstrap may overwrite an
// existing local theme. The default is false to protect local modifications.
var BootstrapAllowThemeOverwrite = false

// TODO: Consider allowing these policy defaults to be overridden via environment
// variables (e.g., for CI) without exposing user-facing config flags. Keep the
// public surface area small until we see demand.
