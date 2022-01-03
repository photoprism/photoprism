/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

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

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import RestModel from "model/rest";
import Api from "common/api";
import { $gettext } from "common/vm";
import { config } from "../session";

export class Account extends RestModel {
  getDefaults() {
    return {
      ID: 0,
      AccName: "",
      AccOwner: "",
      AccURL: "",
      AccType: "",
      AccKey: "",
      AccUser: "",
      AccPass: "",
      AccError: "",
      AccErrors: 0,
      AccShare: true,
      AccSync: false,
      RetryLimit: 3,
      SharePath: "/",
      ShareSize: "",
      ShareExpires: 0,
      SyncPath: "/",
      SyncStatus: "",
      SyncInterval: 86400,
      SyncDate: null,
      SyncFilenames: true,
      SyncUpload: false,
      SyncDownload: !config.get("readonly"),
      SyncRaw: true,
      CreatedAt: "",
      UpdatedAt: "",
      DeletedAt: null,
    };
  }

  getEntityName() {
    return this.AccName;
  }

  getId() {
    return this.ID;
  }

  Folders() {
    return Api.get(this.getEntityResource() + "/folders").then((response) =>
      Promise.resolve(response.data)
    );
  }

  Share(photos, dest) {
    const values = { Photos: photos, Destination: dest };

    return Api.post(this.getEntityResource() + "/share", values).then((response) =>
      Promise.resolve(response.data)
    );
  }

  static getCollectionResource() {
    return "accounts";
  }

  static getModelName() {
    return $gettext("Account");
  }
}

export default Account;
