import RestModel from "model/rest";
import memoizeOne from "memoize-one";
import * as auth from "options/auth";
import $util from "common/util";
import $api from "common/api";
import { T, $gettext } from "common/gettext";
import { Form } from "common/form";
import { $config } from "app/session";

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
    return $api
      .options(this.getEntityResource() + "/register")
      .then((response) => Promise.resolve(new Form(response.data)));
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

    return $api
      .post(this.getEntityResource() + `/avatar`, formData, formConf)
      .then((response) => Promise.resolve(this.setValues(response.data)));
  }

  getProfileForm() {
    return $api
      .options(this.getEntityResource() + "/profile")
      .then((response) => Promise.resolve(new Form(response.data)));
  }

  isRemote() {
    return this.AuthProvider && this.AuthProvider === "ldap";
  }

  requiresPassword() {
    return !this.AuthProvider || this.AuthProvider === "default" || this.AuthProvider === "local";
  }

  // Checks if WebDAV access is allowed for this user.
  hasWebDAV() {
    return this.WebDAV && this.canEnableWebDAV();
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
