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

    Dirs() {
        return Api.get(this.getEntityResource() + "/dirs").then((response) => Promise.resolve(response.data));
    }

    Share(UUIDs, dest) {
        const values = { Photos: UUIDs, Destination: dest };

        return Api.post(this.getEntityResource() + "/share", values).then((response) => Promise.resolve(response.data));
    }

    static getCollectionResource() {
        return "accounts";
    }

    static getModelName() {
        return "Account";
    }
}

export default Account;
