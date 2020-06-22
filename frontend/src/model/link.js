import Model from "./model";
import Api from "../common/api";
import {DateTime} from "luxon";

export default class Link extends Model {
    getDefaults() {
        return {
            UID: "",
            ShareUID: "",
            ShareToken: "",
            ShareExpires: 0,
            Password: "",
            HasPassword: false,
            CanComment: false,
            CanEdit: false,
            CreatedAt: "",
            UpdatedAt: "",
        };
    }

    url() {
        let token = this.ShareToken.toLowerCase();

        if(!token) {
            token = "...";
        }

        return `${window.location.origin}/s/${token}/${this.ShareUID}`;
    }

    caption() {
        return `/s/${this.ShareToken.toLowerCase()}`;
    }

    getId() {
        return this.UID;
    }

    hasId() {
        return !!this.getId();
    }

    clone() {
        return new this.constructor(this.getValues());
    }

    find(id, params) {
        return Api.get(this.getEntityResource(id), params).then((response) => Promise.resolve(new this.constructor(response.data)));
    }

    save() {
        if (this.hasId()) {
            return this.update();
        }

        return Api.post(this.constructor.getCollectionResource(), this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    update() {
        return Api.put(this.getEntityResource(), this.getValues(true)).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    remove() {
        return Api.delete(this.getEntityResource()).then(() => Promise.resolve(this));
    }

    expires() {
        return DateTime.fromISO(this.UpdatedAt).plus({ seconds: this.ShareExpires }).toLocaleString(DateTime.DATE_SHORT);
    }

    static getCollectionResource() {
        return "links";
    }

    static getModelName() {
        return "Link";
    }
}
