/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import RestModel from "model/rest";
import Form from "common/form";
import Api from "common/api";
import {$gettext} from "common/vm";

export class User extends RestModel {
    getDefaults() {
        return {
            UID: "",
            Address: {},
            ParentUID: "",
            GlobalUID: "",
            UserName: "",
            DisplayName: "",
            DisplayLocation: "",
            DisplayBio: "",
            NamePrefix: "",
            GivenName: "",
            FamilyName: "",
            NameSuffix: "",
            Email: "",
            BackupEmail: "",
            URL: "",
            CompanyURL: "",
            Phone: "",
            CompanyPhone: "",
            CompanyName: "",
            DepartmentName: "",
            JobTitle: "",
            BirthYear: -1,
            BirthMonth: -1,
            BirthDay: -1,
            Status: "",
            Avatar: "",
            Active: false,
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
            CreatedAt: "",
            UpdatedAt: "",
        };
    }

    getEntityName() {
        return this.GivenName + " " + this.FamilyName;
    }

    getRegisterForm() {
        return Api.options(this.getEntityResource() + "/register").then(response => Promise.resolve(new Form(response.data)));
    }

    getProfileForm() {
        return Api.options(this.getEntityResource() + "/profile").then(response => Promise.resolve(new Form(response.data)));
    }

    changePassword(oldPassword, newPassword) {
        return Api.put(this.getEntityResource() + "/password", {
            old: oldPassword,
            new: newPassword,
        }).then((response) => Promise.resolve(response.data));
    }

    saveProfile() {
        return Api.post(this.getEntityResource() + "/profile", this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    static getCollectionResource() {
        return "users";
    }

    static getModelName() {
        return $gettext("User");
    }
}

export default User;
