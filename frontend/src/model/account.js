import Abstract from "model/abstract";
import Api from "../common/api";

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
            SyncFilenames: false,
            SyncUpload: false,
            SyncDownload: true,
            SyncRaw: true,
            SyncVideo: true,
            SyncSidecar: true,
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
        const values = {Photos: UUIDs, Destination: dest};

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
