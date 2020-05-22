import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";

export const FolderRootOriginals = "originals"
export const FolderRootImport = "import"

export class Folder extends RestModel {
    getDefaults() {
        return {
            Root: "",
            Path: "",
            PPID: "",
            Title: "",
            Description: "",
            Type: "",
            Order: "",
            Favorite: false,
            Ignore: false,
            Hidden: false,
            Watch: false,
            Links: [],
            CreatedAt: "",
            UpdatedAt: "",
        };
    }

    getEntityName() {
        return this.Root + "/" + this.Path;
    }

    getId() {
        return this.PPID;
    }

    thumbnailUrl(type) {
        return "/api/v1/folder/" + this.getId() + "/thumbnail/" + type;
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

    static findAll(path) {
        return this.search(path, {recursive: true})
    }

    static search(path, params) {
        const options = {
            params: params,
        };

        if (!path || path[0] !== "/") {
            path = "/" + path
        }

        return Api.get(this.getCollectionResource() + path, options).then((response) => {
            let count = response.data.length;
            let limit = 0;
            let offset = 0;

            if (response.headers) {
                if (response.headers["x-count"]) {
                    count = parseInt(response.headers["x-count"]);
                }

                if (response.headers["x-limit"]) {
                    limit = parseInt(response.headers["x-limit"]);
                }

                if (response.headers["x-offset"]) {
                    offset = parseInt(response.headers["x-offset"]);
                }
            }

            response.models = [];
            response.count = count;
            response.limit = limit;
            response.offset = offset;

            for (let i = 0; i < response.data.length; i++) {
                response.models.push(new this(response.data[i]));
            }

            return Promise.resolve(response);
        });
    }

    static getCollectionResource() {
        return "folders";
    }

    static getModelName() {
        return "Folder";
    }
}

export default Folder;
