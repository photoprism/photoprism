import { $gettext } from "common/vm";

// Providers maps account roles to their display name.
export const Roles = () => {
  return {
    admin: $gettext("Admin"),
    user: $gettext("User"),
    viewer: $gettext("Viewer"),
    contributor: $gettext("Contributor"),
    guest: $gettext("Guest"),
    client: $gettext("Client"),
    visitor: $gettext("Visitor"),
    "": $gettext("Unauthorized"),
  };
};

// Providers maps authentication providers to their display name.
export const Providers = () => {
  return {
    "": $gettext("Default"),
    default: $gettext("Default"),
    local: $gettext("Local"),
    client: $gettext("Client"),
    client_credentials: $gettext("Client Credentials"),
    application: $gettext("Application"),
    access_token: $gettext("Access Token"),
    password: $gettext("Local"),
    oidc: $gettext("OIDC"),
    ldap: $gettext("LDAP/AD"),
    link: $gettext("Link"),
    token: $gettext("Link"),
    none: $gettext("None"),
  };
};

// Methods maps authentication methods to their display name.
export const Methods = () => {
  return {
    "": $gettext("Default"),
    default: $gettext("Default"),
    session: $gettext("Session"),
    personal: $gettext("Personal"),
    client: $gettext("Client"),
    access_token: $gettext("Access Token"),
    oauth2: "OAuth2",
    "2fa": $gettext("2FA"),
    oidc: "OIDC",
  };
};

// Scopes maps application scope types to their display name.
export const Scopes = () => {
  return {
    "*": $gettext("Full Access"),
    webdav: $gettext("WebDAV"),
    metrics: $gettext("Metrics"),
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
    cli: "CLI",
    implicit: "Implicit",
    session: $gettext("Session"),
    password: $gettext("Password"),
    client_credentials: "Client Credentials",
    share_token: "Share Token",
    refresh_token: "Refresh Token",
    authorization_code: "Authorization Code",
    "urn:ietf:params:oauth:grant-type:jwt-bearer": "JWT Bearer Assertion",
    "urn:ietf:params:oauth:grant-type:saml2-bearer": "SAML2 Bearer Assertion",
    "urn:ietf:params:oauth:grant-type:token-exchange": "Token Exchange",
  };
};
