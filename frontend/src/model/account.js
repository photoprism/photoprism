import Abstract from "model/abstract";

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
            AccShare: false,
            AccSync: false,
            RetryLimit: 3,
            SharePath: "",
            ShareSize: "",
            ShareExpires: 0,
            ShareExif: true,
            ShareSidecar: false,
            SyncPath: "",
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

    static getCollectionResource() {
        return "accounts";
    }

    static getModelName() {
        return "Account";
    }
}

export default Account;
