import Abstract from "model/abstract";
import Api from "../common/api";
import {DateTime} from "luxon";

class Account extends Abstract {
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
            AccShare: true,
            AccSync: false,
            RetryLimit: 3,
            SharePath: "/",
            ShareSize: "fit_2048",
            ShareExpires: 0,
            ShareExif: true,
            ShareSidecar: false,
            SyncPath: "/",
            SyncInterval: 86400,
            SyncUpload: false,
            SyncDownload: true,
            SyncDelete: false,
            SyncRaw: true,
            SyncVideo: true,
            SyncSidecar: true,
            SyncStart: null,
            SyncedAt: null,
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

    toggleShare() {
        const values = { AccShare: !this.AccShare };

        return Api.put(this.getEntityResource(), values).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    toggleSync() {
        const values = { AccSync: !this.AccSync };

        return Api.put(this.getEntityResource(), values).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    Ls() {
        return Api.get(this.getEntityResource() + "/ls").then((response) => Promise.resolve(response.data));
    }

    static getCollectionResource() {
        return "accounts";
    }

    static getModelName() {
        return "Account";
    }
}

export default Account;
