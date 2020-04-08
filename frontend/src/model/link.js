import Abstract from "model/abstract";

class Link extends Abstract {
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
