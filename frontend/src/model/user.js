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
import Api from "common/api";
import { $gettext } from "common/vm";

export class User extends RestModel {
  getDefaults() {
    return {
      UID: "",
      Address: {},
      MotherUID: "",
      FatherUID: "",
      GlobalUID: "",
      FullName: "",
      NickName: "",
      MaidenName: "",
      ArtistName: "",
      UserName: "",
      UserStatus: "",
      UserDisabled: false,
      UserSettings: "",
      PrimaryEmail: "",
      EmailConfirmed: false,
      BackupEmail: "",
      PersonURL: "",
      PersonPhone: "",
      PersonStatus: "",
      PersonAvatar: "",
      PersonLocation: "",
      PersonBio: "",
      BusinessURL: "",
      BusinessPhone: "",
      BusinessEmail: "",
      CompanyName: "",
      DepartmentName: "",
      JobTitle: "",
      BirthYear: -1,
      BirthMonth: -1,
      BirthDay: -1,
      TermsAccepted: false,
      IsArtist: false,
      IsSubject: false,
      RoleAdmin: false,
      RoleGuest: false,
      RoleChild: false,
      RoleFamily: false,
      RoleFriend: false,
      WebDAV: false,
      StoragePath: "",
      CanInvite: false,
      InviteToken: "",
      InvitedBy: "",
      CreatedAt: "",
      UpdatedAt: "",
    };
  }

  getEntityName() {
    return this.FullName ? this.FullName : this.UserName;
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
