import RestModel from "model/rest";

class Link extends RestModel {
    getDefaults() {
        return {
            LinkToken: "",
            LinkPassword: "",
            LinkExpires: "",
            ShareUUID: "",
            CanComment: false,
            CanEdit: false,
            CreatedAt: "",
            UpdatedAt: "",
            Links: [],
        };
    }

    getId() {
        return this.LinkToken;
    }

    static getCollectionResource() {
        return "links";
    }

    static getModelName() {
        return "Link";
    }
}

export default Link;
