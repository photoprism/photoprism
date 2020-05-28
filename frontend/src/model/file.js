import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";
import Util from "common/util";
import {config} from "../session";

export class File extends RestModel {
    getDefaults() {
        return {
            InstanceID: "",
            UID: "",
            PhotoUID: "",
            Root: "",
            Name: "",
            OriginalName: "",
            Hash: "",
            Modified: "",
            Size: 0,
            Codec: "",
            Type: "",
            Mime: "",
            Primary: false,
            Sidecar: false,
            Missing: false,
            Duplicate: false,
            Portrait: false,
            Video: false,
            Duration: 0,
            Width: 0,
            Height: 0,
            Orientation: 0,
            AspectRatio: 1.0,
            MainColor: "",
            Colors: "",
            Luminance: "",
            Diff: 0,
            Chroma: 0,
            Notes: "",
            Error: "",
            Links: [],
            CreatedAt: "",
            CreatedIn: 0,
            UpdatedAt: "",
            UpdatedIn: 0,
            DeletedAt: "",
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
        return true;
    }

    getEntityName() {
        return this.Root + "/" + this.Name;
    }

    getId() {
        return this.UID;
    }

    thumbnailUrl(type) {
        if(this.Error) {
            return "/api/v1/svg/broken";
        } else if (this.Type === "raw") {
            return "/api/v1/svg/raw";
        }

        return `/api/v1/t/${this.Hash}/${config.previewToken()}/${type}`;
    }

    getDownloadUrl() {
        return "/api/v1/dl/" + this.Hash + "?t=" + config.downloadToken();
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

    calculateSize(width, height) {
        if (width >= this.Width && height >= this.Height) { // Smaller
            return {width: this.Width, height: this.Height};
        }

        const srcAspectRatio = this.Width / this.Height;
        const maxAspectRatio = width / height;

        let newW, newH;

        if (srcAspectRatio > maxAspectRatio) {
            newW = width;
            newH = Math.round(newW / srcAspectRatio);

        } else {
            newH = height;
            newW = Math.round(newH * srcAspectRatio);
        }

        return {width: newW, height: newH};
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

    getInfo() {
        let info = [];

        if (this.Type) {
            info.push(this.Type.toUpperCase());
        }

        if (this.Duration > 0) {
            info.push(Util.duration(this.Duration));
        }

        this.addSizeInfo(info);

        return info.join(", ");
    }

    addSizeInfo(info) {
        if (this.Width && this.Height) {
            info.push(this.Width + " × " + this.Height);
        }

        if (this.Size > 102400) {
            const size = Number.parseFloat(this.Size) / 1048576;

            info.push(size.toFixed(1) + " MB");
        } else if (this.Size) {
            const size = Number.parseFloat(this.Size) / 1024;

            info.push(size.toFixed(1) + " KB");
        }
    }

    toggleLike() {
        this.Favorite = !this.Favorite;

        if (this.Favorite) {
            return Api.post(this.getPhotoResource() + "/like");
        } else {
            return Api.delete(this.getPhotoResource() + "/like");
        }
    }

    getPhotoResource() {
        return "photos/" + this.PhotoUID;
    }

    like() {
        this.Favorite = true;
        return Api.post(this.getPhotoResource() + "/like");
    }

    unlike() {
        this.Favorite = false;
        return Api.delete(this.getPhotoResource() + "/like");
    }

    static getCollectionResource() {
        return "files";
    }

    static getModelName() {
        return "File";
    }
}

export default File;
