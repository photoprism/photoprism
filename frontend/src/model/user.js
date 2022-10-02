/*

Copyright (c) 2018 - 2022 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

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

export class User extends RestModel {
  getDefaults() {
    return {
      UID: "",
      UUID: "",
      AuthProvider: "",
      AuthID: "",
      Name: "",
      DisplayName: "",
      Email: "",
      BackupEmail: "",
      Role: "",
      Attr: "",
      SuperAdmin: false,
      CanLogin: false,
      BasePath: "",
      UploadPath: "",
      CanSync: false,
      CanInvite: false,
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
        IdURL: "",
        AvatarURL: "",
        SiteURL: "",
        FeedURL: "",
        UserGender: "",
        NamePrefix: "",
        GivenName: "",
        MiddleName: "",
        FamilyName: "",
        NameSuffix: "",
        NickName: "",
        UserPhone: "",
        UserAddress: "",
        UserCountry: "",
        UserBio: "",
        JobTitle: "",
        Department: "",
        Company: "",
        CompanyURL: "",
        BirthYear: -1,
        BirthMonth: -1,
        BirthDay: -1,
        CreatedAt: "",
        UpdatedAt: "",
      },
      VerifiedAt: "",
      ConsentAt: "",
      BornAt: "",
      CreatedAt: "",
      UpdatedAt: "",
      ExpiresAt: "",
    };
  }

  getDisplayName() {
    if (this.DisplayName) {
      return this.DisplayName;
    } else if (this.Details && this.Details.NickName) {
      return this.Details.NickName;
    } else if (this.Details && this.Details.GivenName) {
      return this.Details.GivenName;
    } else if (this.Name) {
      return Util.capitalize(this.Name);
    } else if (this.Details && this.Details.JobTitle) {
      return this.Details.JobTitle;
    } else if (this.Email) {
      return this.Email;
    } else if (this.Role) {
      return T(Util.capitalize(this.Role));
    }

    return $gettext("Unregistered");
  }

  getAccountInfo() {
    if (this.Email) {
      return this.Email;
    } else if (this.Details && this.Details.JobTitle) {
      return this.Details.JobTitle;
    } else if (this.Role) {
      return T(Util.capitalize(this.Role));
    } else if (this.Name) {
      return this.Name;
    }

    return $gettext("Account");
  }

  getEntityName() {
    return this.getDisplayName();
  }

  getRegisterForm() {
    return Api.options(this.getEntityResource() + "/register").then((response) =>
      Promise.resolve(new Form(response.data))
    );
  }

  getProfileForm() {
    return Api.options(this.getEntityResource() + "/profile").then((response) =>
      Promise.resolve(new Form(response.data))
    );
  }

  changePassword(oldPassword, newPassword) {
    return Api.put(this.getEntityResource() + "/password", {
      old: oldPassword,
      new: newPassword,
    }).then((response) => Promise.resolve(response.data));
  }

  saveProfile() {
    return Api.post(this.getEntityResource() + "/profile", this.getValues()).then((response) =>
      Promise.resolve(this.setValues(response.data))
    );
  }

  static getCollectionResource() {
    return "users";
  }

  static getModelName() {
    return $gettext("User");
  }
}

export default User;
