import RestModel from "model/rest";
import memoizeOne from "memoize-one";
import * as auth from "options/auth";
import $util from "common/util";
import $api from "common/api";
import { T, $gettext } from "common/gettext";
import { Form } from "common/form";
import { $config, $session } from "app/session";

export let BatchSize = 99999;
export let WebDavRoles = ["admin", "manager", "user", "contributor"];
export let NoBasePathRoles = ["admin", "manager", "user", "viewer"];
export let NoUploadPathRoles = ["guest", "viewer"];

// User encapsulates account metadata, roles, and helpers for access control.
export class User extends RestModel {
  getDefaults() {
    return {
      ID: 0,
      UID: "",
      UUID: "",
      AuthProvider: "",
      AuthMethod: "",
      AuthIssuer: "",
      AuthID: "",
      Name: "",
      DisplayName: "",
      Email: "",
      BackupEmail: "",
      Role: "",
      Scope: "",
      Attr: "",
      SuperAdmin: false,
      CanLogin: false,
      CanInvite: false,
      BasePath: "",
      UploadPath: "",
      WebDAV: false,
      Thumb: "",
      ThumbSrc: "",
      Settings: {
        UITheme: "",
        UILanguage: "",
        UITimeZone: "",
        MapsStyle: "",
        MapsAnimate: 0,
        IndexPath: "",
        IndexRescan: 0,
        ImportPath: "",
        ImportMove: 0,
        UploadPath: "",
        DefaultPage: "",
        CreatedAt: "",
        UpdatedAt: "",
      },
      Details: {
        SubjUID: "",
        SubjSrc: "",
        PlaceID: "",
        PlaceSrc: "",
        CellID: "",
        BirthYear: -1,
        BirthMonth: -1,
        BirthDay: -1,
        NameTitle: "",
        GivenName: "",
        MiddleName: "",
        FamilyName: "",
        NameSuffix: "",
        NickName: "",
        NameSrc: "",
        Gender: "",
        About: "",
        Bio: "",
        Location: "",
        Country: "zz",
        Phone: "",
        SiteURL: "",
        ProfileURL: "",
        FeedURL: "",
        AvatarURL: "",
        OrgTitle: "",
        OrgName: "",
        OrgEmail: "",
        OrgPhone: "",
        OrgURL: "",
        IdURL: "",
        CreatedAt: "",
        UpdatedAt: "",
      },
      LoginAt: "",
      VerifiedAt: "",
      ConsentAt: "",
      BornAt: "",
      CreatedAt: "",
      UpdatedAt: "",
      ExpiresAt: "",
    };
  }

  getHandle() {
    if (!this.Name) {
      return "";
    }

    const s = this.Name.split("@");
    return s[0].trim();
  }

  defaultBasePath() {
    const handle = this.getHandle();

    if (!handle) {
      return "";
    }

    let dir = $config.get("usersPath");

    if (dir) {
      return `${dir}/${handle}`;
    } else {
      return `users/${handle}`;
    }
  }

  getDisplayName() {
    if (this.DisplayName) {
      return this.DisplayName;
    } else if (this.Details && this.Details.NickName) {
      return this.Details.NickName;
    } else if (this.Details && this.Details.GivenName) {
      return this.Details.GivenName;
    } else if (this.Name) {
      return T($util.capitalize(this.Name));
    }

    return $gettext("Unknown");
  }

  getAccountInfo() {
    if (this.Name) {
      return this.Name;
    } else if (this.Email) {
      return this.Email;
    } else if (this.Details && this.Details.JobTitle) {
      return this.Details.JobTitle;
    } else if (this.Role) {
      return T($util.capitalize(this.Role));
    }

    return $gettext("Account");
  }

  getEntityName() {
    return this.getDisplayName();
  }

  getRegisterForm() {
    return $api.options(this.getEntityResource() + "/register").then((response) => Promise.resolve(new Form(response.data)));
  }

  getAvatarURL(size, config) {
    if (!size) {
      size = "tile_500";
    }

    if (!config) {
      config = $config;
    }

    if (this.Thumb) {
      return `${config.contentUri}/t/${this.Thumb}/${config.previewToken}/${size}`;
    } else {
      return `${config.staticUri}/img/avatar/${size}.jpg`;
    }
  }

  uploadAvatar(files) {
    if (this.busy) {
      return Promise.reject(this);
    } else if (!files || files.length !== 1) {
      return Promise.reject(this);
    }

    let file = files[0];
    let formData = new FormData();
    let formConf = { headers: { "Content-Type": "multipart/form-data" } };

    formData.append("files", file);

    return $api.post(this.getEntityResource() + `/avatar`, formData, formConf).then((response) => Promise.resolve(this.setValues(response.data)));
  }

  getProfileForm() {
    return $api.options(this.getEntityResource() + "/profile").then((response) => Promise.resolve(new Form(response.data)));
  }

  hasScope() {
    return Boolean(this.Scope) && this.Scope !== "*";
  }

  getScope() {
    if (this.hasScope()) {
      return this.Scope;
    }

    return "*";
  }

  // isRemote returns true when the user is authenticated through a remote provider (currently LDAP).
  isRemote() {
    return this.AuthProvider === "ldap";
  }

  // isAdmin returns true when the account holds an admin-tier role (admin or
  // cluster_admin), mirroring the backend acl.IsAdminRole set. Role-based only
  // (it ignores SuperAdmin) so it can gate controls such as the Super Admin toggle.
  isAdmin() {
    return this.Role === "admin" || this.Role === "cluster_admin";
  }

  // isClusterAdmin returns true when the account holds the Portal operator role.
  isClusterAdmin() {
    return this.Role === "cluster_admin";
  }

  // isCurrentUser returns true when this account belongs to the signed-in user.
  // The admin UI uses it to lock fields that would otherwise let an operator
  // lock themselves out (own role, web login, and authentication provider); the
  // backend rejects these self-changes regardless.
  isCurrentUser() {
    const current = $session.getUser();
    return !!(current && current.UID && this.UID && current.UID === this.UID);
  }

  requiresPassword() {
    return !this.AuthProvider || this.AuthProvider === "default" || this.AuthProvider === "local";
  }

  // showsPasswordField reports whether the local password input is shown.
  // OIDC may keep a local password as a fallback; LDAP replaces it; "none" disables auth.
  showsPasswordField() {
    return ["default", "local", "oidc"].includes(this.AuthProvider);
  }

  // passwordIsRequired reports whether a local password must be set to create the account.
  // "default" needs one unless an external identity can be supplied instead (OIDC/LDAP configured).
  passwordIsRequired() {
    if (this.AuthProvider === "local") {
      return true;
    }
    if (this.AuthProvider === "default") {
      return !this.showsAuthIdField();
    }
    return false;
  }

  // showsAuthIdField reports whether the external identity input (OIDC Subject ID or LDAP DN) is shown.
  // "default" offers it only when OIDC or LDAP is configured, to pre-provision external accounts.
  showsAuthIdField() {
    const p = this.AuthProvider;
    return p === "oidc" || p === "ldap" || (p === "default" && ($config.oidcEnabled() || $config.ldapEnabled()));
  }

  // authIdIsRequired reports whether the external identity must be set; only OIDC requires it up front.
  authIdIsRequired() {
    return this.AuthProvider === "oidc";
  }

  // authIdIsDn reports whether the external identity is an LDAP DN rather than an OIDC Subject ID.
  authIdIsDn() {
    const p = this.AuthProvider;
    return p === "ldap" || (p === "default" && $config.ldapEnabled() && !$config.oidcEnabled());
  }

  // authIdFieldLabel returns the label for the external identity input.
  authIdFieldLabel() {
    return this.authIdIsDn() ? "Distinguished Name (DN)" : "Subject ID";
  }

  // hasLoginCredential reports whether enough is set for the new account to authenticate.
  // Covers the gap per-field rules miss: "default" with external auth configured needs a
  // password, an LDAP directory (username resolves it), or an OIDC Subject ID — else it is inert.
  hasLoginCredential() {
    if (this.AuthProvider !== "default" || !this.showsAuthIdField()) {
      return true;
    }
    return !!this.Password || $config.ldapEnabled() || !!this.AuthID;
  }

  // canEnableLogin reports whether web/API login can be enabled for this account.
  // System users, role-less accounts, visitors, and deactivated accounts (provider
  // "none") cannot log in regardless of the toggle. LDAP and OIDC accounts can.
  canEnableLogin() {
    if (this.ID < 1 || !this.Name || this.AuthProvider === "none") {
      return false;
    }

    return !!this.Role && this.Role !== "visitor";
  }

  // canHavePassword reports whether a local password can be set and used for this account.
  // Requires login eligibility plus a local provider — remote accounts (LDAP) have their
  // credentials managed externally, so a local password would be inert.
  canHavePassword() {
    return this.canEnableLogin() && !this.isRemote();
  }

  // hasWebDAV returns true when WebDAV access is enabled for this user and the role permits it.
  hasWebDAV() {
    return !!this.WebDAV && this.canEnableWebDAV();
  }

  // Checks if the user role permits WebDAV access.
  canEnableWebDAV() {
    if (this.AuthProvider === "none" || !this.Name) {
      return false;
    }

    return WebDavRoles.includes(this.Role);
  }

  // Checks if the user role supports a custom base path.
  canHaveBasePath() {
    return !NoBasePathRoles.includes(this.Role);
  }

  // Checks if the user role supports a custom upload path.
  canHaveUploadPath() {
    return !NoUploadPathRoles.includes(this.Role);
  }

  authInfo() {
    if (!this || !this.AuthProvider) {
      return $gettext("Default");
    }

    let providerName = memoizeOne(auth.Providers)()[this.AuthProvider];

    if (providerName) {
      providerName = T(providerName);
    } else {
      providerName = $util.capitalize(this.AuthProvider);
    }

    if (!this.AuthMethod || this.AuthMethod === "" || this.AuthMethod === "default") {
      return providerName;
    }

    let methodName = memoizeOne(auth.Methods)()[this.AuthMethod];

    if (!methodName) {
      methodName = this.AuthMethod;
    }

    return `${providerName} (${methodName})`;
  }

  changePassword(oldPassword, newPassword) {
    return $api
      .put(this.getEntityResource() + "/password", {
        old: oldPassword,
        new: newPassword,
      })
      .then((response) => Promise.resolve(response.data));
  }

  createPasscode(password) {
    return $api
      .post(this.getEntityResource() + "/passcode", {
        type: "totp",
        password: password,
      })
      .then((response) => Promise.resolve(response.data));
  }

  confirmPasscode(code) {
    return $api
      .post(this.getEntityResource() + "/passcode/confirm", {
        type: "totp",
        code: code,
      })
      .then((response) => Promise.resolve(response.data));
  }

  activatePasscode() {
    return $api
      .post(this.getEntityResource() + "/passcode/activate", {
        type: "totp",
      })
      .then((response) => Promise.resolve(response.data));
  }

  deactivatePasscode(password) {
    return $api
      .post(this.getEntityResource() + "/passcode/deactivate", {
        type: "totp",
        password: password,
      })
      .then((response) => Promise.resolve(response.data));
  }

  disablePasscodeSetup(hasPassword) {
    if (!this.Name || !this.CanLogin || this.ID < 1) {
      return true;
    }

    switch (this.AuthProvider) {
      case "":
      case "default":
      case "oidc":
        return !hasPassword;
      case "local":
      case "ldap":
        return false;
      default:
        return true;
    }
  }

  findApps() {
    if (!this.Name || !this.CanLogin || this.ID < 1) {
      return Promise.reject();
    }

    const params = {
      provider: "application",
      method: "default",
      count: 10000,
      offset: 0,
      order: "client_name",
    };

    return $api
      .get(this.getEntityResource() + "/sessions", {
        params,
      })
      .then((response) => Promise.resolve(response.data));
  }

  static batchSize() {
    return BatchSize;
  }

  static getCollectionResource() {
    return "users";
  }

  static getModelName() {
    return $gettext("User");
  }
}

export default User;
