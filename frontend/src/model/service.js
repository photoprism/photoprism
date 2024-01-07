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
import Api from "common/api";
import { $gettext } from "common/vm";
import { config } from "app/session";

export class Service extends RestModel {
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
      AccTimeout: "",
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
    return this.ID ? this.ID : false;
  }

  Folders() {
    return Api.get(this.getEntityResource() + "/folders").then((response) =>
      Promise.resolve(response.data)
    );
  }

  Upload(selection, folder) {
    if (!selection) {
      return;
    }

    if (Array.isArray(selection)) {
      selection = { Photos: selection };
    }

    return Api.post(this.getEntityResource() + "/upload", { selection, folder }).then((response) =>
      Promise.resolve(response.data)
    );
  }

  static getCollectionResource() {
    return "services";
  }

  static getModelName() {
    return $gettext("Account");
  }
}

export default Service;
