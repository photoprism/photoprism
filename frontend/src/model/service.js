import RestModel from "model/rest";
import $api from "common/api";
import { $gettext } from "common/gettext";
import { $config } from "app/session";

// Service captures external service connections (WebDAV, S3, etc.).
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
      SyncDownload: !$config.get("readonly"),
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
    return $api.get(this.getEntityResource() + "/folders").then((response) => Promise.resolve(response.data));
  }

  Upload(selection, folder) {
    if (!selection) {
      return;
    }

    if (Array.isArray(selection)) {
      selection = { Photos: selection };
    }

    return $api
      .post(this.getEntityResource() + "/upload", { selection, folder })
      .then((response) => Promise.resolve(response.data));
  }

  static getCollectionResource() {
    return "services";
  }

  static getModelName() {
    return $gettext("Account");
  }
}

export default Service;
