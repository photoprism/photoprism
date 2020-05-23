import RestModel from "model/rest";

export class Link extends RestModel {
    getDefaults() {
        return {
            Token: "",
            Password: "",
            Expires: "",
            ShareUID: "",
            CanComment: false,
            CanEdit: false,
            CreatedAt: "",
            UpdatedAt: "",
            Links: [],
        };
    }

    getId() {
        return this.Token;
    }

    static getCollectionResource() {
        return "links";
    }

    static getModelName() {
        return "Link";
    }
}

export default Link;
