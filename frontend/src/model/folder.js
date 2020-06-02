import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";
import File from "model/file";
import Util from "../common/util";

export const RootImport = "import";
export const RootOriginals = "originals";

export class Folder extends RestModel {
    getDefaults() {
        return {
            Folder: true,
            Path: "",
            Root: "",
            UID: "",
            Type: "",
            Title: "",
            Category: "",
            Description: "",
            Order: "",
            Country: "",
            Year: "",
            Month: "",
            Favorite: false,
            Private: false,
            Ignore: false,
            Watch: false,
            FileCount: 0,
            Links: [],
            CreatedAt: "",
            UpdatedAt: "",
        };
    }

    baseName(truncate) {
        let result = this.Name;
        const slash = result.lastIndexOf("/");

        if (slash >= 0) {
            result = this.Name.substring(slash + 1);
        }

        if(truncate) {
            result = Util.truncate(result, truncate, "...");
        }

        return result;
    }

    isFile() {
        return false;
    }

    getEntityName() {
        return this.Root + "/" + this.Path;
    }

    getId() {
        return this.UID;
    }

    thumbnailUrl() {
        return "/api/v1/svg/folder";
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
        return this.search(path, {recursive: true});
    }

    static findAllUncached(path) {
        return this.search(path, {recursive: true, uncached: true});
    }

    static originals(path, params) {
        if(!path) {
            path = "/";
        }

        return this.search(RootOriginals + path, params);
    }

    static search(path, params) {
        const options = {
            params: params,
        };

        if (!path || path[0] !== "/") {
            path = "/" + path;
        }

        return Api.get(this.getCollectionResource() + path, options).then((response) => {
            let folders = response.data.folders;
            let files = response.data.files ? response.data.files : [];

            let count = folders.length + files.length;

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

            for (let i = 0; i < folders.length; i++) {
                response.models.push(new this(folders[i]));
            }

            for (let i = 0; i < files.length; i++) {
                response.models.push(new File(files[i]));
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
