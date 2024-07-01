/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import RestModel from "model/rest";
import Form from "common/form";
import Util from "common/util";
import Api from "common/api";
import { T, $gettext } from "common/vm";
import { config } from "app/session";
import memoizeOne from "memoize-one";
import * as auth from "../options/auth";

export class User extends RestModel {
  getDefaults() {
    return {
      ID: 0,
      UID: "",
      UUID: "",
      AuthProvider: "",
      AuthMethod: "",
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
        Country: "",
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

    let dir = config.get("usersPath");

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
      return T(Util.capitalize(this.Name));
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
      return T(Util.capitalize(this.Role));
    }

    return $gettext("Account");
  }

  getEntityName() {
    return this.getDisplayName();
  }

  getRegisterForm() {
    return Api.options(this.getEntityResource() + "/register").then((response) => Promise.resolve(new Form(response.data)));
  }

  getAvatarURL(size) {
    if (!size) {
      size = "tile_500";
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

    return Api.post(this.getEntityResource() + `/avatar`, formData, formConf).then((response) => Promise.resolve(this.setValues(response.data)));
  }

  getProfileForm() {
    return Api.options(this.getEntityResource() + "/profile").then((response) => Promise.resolve(new Form(response.data)));
  }

  isRemote() {
    return this.AuthProvider && (this.AuthProvider === "ldap" || this.AuthProvider === "oidc");
  }

  authInfo() {
    if (!this || !this.AuthProvider) {
      return $gettext("Default");
    }

    let providerName = memoizeOne(auth.Providers)()[this.AuthProvider];

    if (providerName) {
      providerName = T(providerName);
    } else {
      providerName = Util.capitalize(this.AuthProvider);
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
    return Api.put(this.getEntityResource() + "/password", {
      old: oldPassword,
      new: newPassword,
    }).then((response) => Promise.resolve(response.data));
  }

  createPasscode(password) {
    return Api.post(this.getEntityResource() + "/passcode", {
      type: "totp",
      password: password,
    }).then((response) => Promise.resolve(response.data));
  }

  confirmPasscode(code) {
    return Api.post(this.getEntityResource() + "/passcode/confirm", {
      type: "totp",
      code: code,
    }).then((response) => Promise.resolve(response.data));
  }

  activatePasscode() {
    return Api.post(this.getEntityResource() + "/passcode/activate", {
      type: "totp",
    }).then((response) => Promise.resolve(response.data));
  }

  deactivatePasscode(password) {
    return Api.post(this.getEntityResource() + "/passcode/deactivate", {
      type: "totp",
      password: password,
    }).then((response) => Promise.resolve(response.data));
  }

  disablePasscodeSetup() {
    if (!this.Name || !this.CanLogin || this.ID < 1) {
      return true;
    }

    switch (this.AuthProvider) {
      case "":
      case "default":
      case "local":
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

    return Api.get(this.getEntityResource() + "/sessions", {
      params,
    }).then((response) => Promise.resolve(response.data));
  }

  static getCollectionResource() {
    return "users";
  }

  static getModelName() {
    return $gettext("User");
  }
}

export default User;
