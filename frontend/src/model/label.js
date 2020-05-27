import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";
import {config} from "../session";

export class Label extends RestModel {
    getDefaults() {
        return {
            ID: 0,
            UID: "",
            Slug: "",
            CustomSlug: "",
            Name: "",
            Priority: 0,
            Favorite: false,
            Description: "",
            Notes: "",
            PhotoCount: 0,
            Links: [],
            CreatedAt: "",
            UpdatedAt: "",
            DeletedAt: "",
        };
    }

    getEntityName() {
        return this.Slug;
    }

    getId() {
        return this.UID;
    }

    getTitle() {
        return this.Name;
    }

    thumbnailUrl(type) {
        return `/api/v1/labels/${this.getId()}/t/${config.previewToken()}/${type}`;
    }

    thumbnailSrcset() {
        const result = [];

        result.push(this.thumbnailUrl("fit_720") + " 720w");
        result.push(this.thumbnailUrl("fit_1280") + " 1280w");
        result.push(this.thumbnailUrl("fit_1920") + " 1920w");
        result.push(this.thumbnailUrl("fit_2560") + " 2560w");
        result.push(this.thumbnailUrl("fit_3840") + " 3840w");

        return result.join(", ");
    }

    thumbnailSizes() {
        const result = [];

        result.push("(min-width: 2560px) 3840px");
        result.push("(min-width: 1920px) 2560px");
        result.push("(min-width: 1280px) 1920px");
        result.push("(min-width: 720px) 1280px");
        result.push("720px");

        return result.join(", ");
    }

    getDateString() {
        return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
    }

    toggleLike() {
        this.Favorite = !this.Favorite;

        if (this.Favorite) {
            return Api.post(this.getEntityResource() + "/like");
        } else {
            return Api.delete(this.getEntityResource() + "/like");
        }
    }

    like() {
        this.Favorite = true;
        return Api.post(this.getEntityResource() + "/like");
    }

    unlike() {
        this.Favorite = false;
        return Api.delete(this.getEntityResource() + "/like");
    }

    static getCollectionResource() {
        return "labels";
    }

    static getModelName() {
        return "Label";
    }
}

export default Label;
