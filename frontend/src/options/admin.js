import { $gettext } from "common/vm";

// All user role with their display name.
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

// Selectable user role options.
export const RoleOptions = () => [
  {
    text: $gettext("Admin"),
    value: "admin",
  },
  {
    text: $gettext("User"),
    value: "user",
  },
  {
    text: $gettext("Viewer"),
    value: "viewer",
  },
  {
    text: $gettext("Contributor"),
    value: "contributor",
  },
  {
    text: $gettext("Guest"),
    value: "guest",
  },
];

// AuthProviders maps known auth providers to their display name.
export const AuthProviders = () => {
  return {
    "": $gettext("Default"),
    default: $gettext("Default"),
    local: $gettext("Local"),
    client: $gettext("Client"),
    password: $gettext("Local"),
    ldap: $gettext("LDAP/AD"),
    link: $gettext("Link"),
    token: $gettext("Link"),
    none: $gettext("None"),
  };
};

// AuthMethods maps known auth methods to their display name.
export const AuthMethods = () => {
  return {
    "": $gettext("Default"),
    default: $gettext("Default"),
    access_token: $gettext("Access Token"),
    session: $gettext("Session"),
    "2fa": "2FA",
    oauth2: "OAuth2",
    oidc: "OIDC",
  };
};

// Selectable auth provider options.
export const AuthProviderOptions = (includeLdap) => {
  if (includeLdap) {
    return [
      {
        text: $gettext("Default"),
        value: "default",
      },
      {
        text: $gettext("Local"),
        value: "local",
      },
      {
        text: $gettext("LDAP/AD"),
        value: "ldap",
      },
      {
        text: $gettext("None"),
        value: "none",
      },
    ];
  } else {
    return [
      {
        text: $gettext("Default"),
        value: "default",
      },
      {
        text: $gettext("Local"),
        value: "local",
      },
      {
        text: $gettext("None"),
        value: "none",
      },
    ];
  }
};
