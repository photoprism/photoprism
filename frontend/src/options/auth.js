import { $gettext } from "common/gettext";

// Roles maps account roles to their display name.
export const Roles = () => {
  return {
    "admin": $gettext("Admin"),
    "cluster_admin": $gettext("Cluster Admin"),
    "manager": $gettext("Manager"),
    "user": $gettext("User"),
    "viewer": $gettext("Viewer"),
    "contributor": $gettext("Contributor"),
    "guest": $gettext("Guest"),
    "client": $gettext("Client"),
    "visitor": $gettext("Visitor"),
    "": $gettext("Unauthorized"),
  };
};

// RoleLabel returns the display name for an account role, or the key if unmapped.
export const RoleLabel = (role) => {
  const labels = Roles();
  return Object.prototype.hasOwnProperty.call(labels, role) ? labels[role] : role;
};

// RoleOptions builds {value, [labelKey]} options for the given role keys, with
// labels from the shared Roles() map; editions pass their own keys (labelKey is
// "text" or "title"). A non-null `current` outside the list (a downgrade leftover
// or the empty "" Unauthorized role) is appended labeled + disabled; null = none.
export const RoleOptions = (roles, labelKey = "text", current = null) => {
  const options = roles.map((role) => ({ value: role, [labelKey]: RoleLabel(role) }));
  if (current !== null && !roles.includes(current)) {
    options.push({ value: current, [labelKey]: RoleLabel(current), disabled: true });
  }
  return options;
};

// Providers maps authentication providers to their display name.
export const Providers = () => {
  return {
    "": $gettext("Default"),
    "default": $gettext("Default"),
    "local": $gettext("Local"),
    "client": $gettext("Client"),
    "client_credentials": "Client Credentials",
    "application": $gettext("Application"),
    "access_token": $gettext("Access Token"),
    "password": $gettext("Local"),
    "oidc": "OIDC",
    "ldap": $gettext("LDAP/AD"),
    "link": $gettext("Link"),
    "token": $gettext("Link"),
    "none": $gettext("None"),
  };
};

// ProviderLabel returns the display name for an auth provider, or the key if unmapped.
export const ProviderLabel = (provider) => {
  const labels = Providers();
  return Object.prototype.hasOwnProperty.call(labels, provider) ? labels[provider] : provider;
};

// ProviderOptions builds {value, [labelKey]} options for the given provider
// keys, with labels from the shared Providers() map. Editions pass their own keys.
export const ProviderOptions = (providers, labelKey = "text") => providers.map((provider) => ({ value: provider, [labelKey]: ProviderLabel(provider) }));

// Methods maps authentication methods to their display name.
export const Methods = () => {
  return {
    "": $gettext("Default"),
    "default": $gettext("Default"),
    "session": $gettext("Session"),
    "personal": $gettext("Personal"),
    "client": $gettext("Client"),
    "access_token": $gettext("Access Token"),
    "oauth2": "OAuth2",
    "2fa": $gettext("2FA"),
    "oidc": "OIDC",
  };
};

// Scopes maps application scope types to their display name.
export const Scopes = () => {
  return {
    "*": $gettext("Full Access"),
    "webdav": $gettext("WebDAV"),
    "metrics": $gettext("Metrics"),
  };
};

// ScopeOptions returns selectable application scope types.
export const ScopeOptions = () => {
  return [
    {
      text: $gettext("Full Access"),
      value: "*",
    },
    {
      text: $gettext("WebDAV"),
      value: "webdav",
    },
    {
      text: $gettext("Metrics"),
      value: "metrics",
    },
    /* TODO: Show additional input field so advanced users can specify a custom scope when this option is selected.
    {
      text: $gettext("Custom"),
      value: "~",
    },
    */
  ];
};

// GrantTypes maps grant types to their display name.
export const GrantTypes = () => {
  return {
    "": "Default",
    "cli": "CLI",
    "implicit": "Implicit",
    "session": $gettext("Session"),
    "password": $gettext("Password"),
    "client_credentials": "Client Credentials",
    "share_token": "Share Token",
    "refresh_token": "Refresh Token",
    "authorization_code": "Authorization Code",
    "urn:ietf:params:oauth:grant-type:jwt-bearer": "JWT Bearer Assertion",
    "urn:ietf:params:oauth:grant-type:saml2-bearer": "SAML2 Bearer Assertion",
    "urn:ietf:params:oauth:grant-type:token-exchange": "Token Exchange",
  };
};
