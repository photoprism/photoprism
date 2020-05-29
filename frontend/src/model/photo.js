import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";
import Util from "common/util";
import {config} from "../session";
import countries from "resources/countries.json";

export const SrcManual = "manual";
export const CodecAvc1 = "avc1";
export const TypeMP4 = "mp4";
export const TypeJpeg = "jpg";
export const TypeImage = "image";
export const YearUnknown = -1;
export const MonthUnknown = -1;

export class Photo extends RestModel {
    getDefaults() {
        return {
            DocumentID: "",
            UID: "",
            Type: TypeImage,
            Favorite: false,
            Private: false,
            TakenAt: "",
            TakenAtLocal: "",
            TakenSrc: "",
            TimeZone: "",
            Path: "",
            Color: "",
            Name: "",
            Title: "",
            TitleSrc: "",
            Description: "",
            DescriptionSrc: "",
            Resolution: 0,
            Quality: 0,
            Lat: 0.0,
            Lng: 0.0,
            Altitude: 0,
            Iso: 0,
            FocalLength: 0,
            FNumber: 0.0,
            Exposure: "",
            Views: 0,
            Camera: {},
            CameraID: 0,
            CameraSrc: "",
            Lens: {},
            LensID: 0,
            Country: "",
            Year: YearUnknown,
            Month: MonthUnknown,
            Details: {
                Keywords: "",
                Notes: "",
                Subject: "",
                Artist: "",
                Copyright: "",
                License: "",
            },
            Files: [],
            Labels: [],
            Keywords: [],
            Albums: [],
            Links: [],
            Location: {},
            Place: {},
            PlaceID: "",
            LocationID: "",
            LocSrc: "",
            // Additional data in result lists.
            LocLabel: "",
            LocCity: "",
            LocState: "",
            LocCountry: "",
            FileUID: "",
            FileRoot: "",
            FileName: "",
            Hash: "",
            Width: "",
            Height: "",
            // Date fields.
            CreatedAt: "",
            UpdatedAt: "",
            DeletedAt: null,
        };
    }

    baseName(truncate) {
        let result = this.fileBase(this.FileName ? this.FileName : this.mainFile().Name);

        if (truncate) {
            result = Util.truncate(result, truncate, "...");
        }

        return result;
    }

    fileBase(name) {
        let result = name;
        const slash = result.lastIndexOf("/");

        if (slash >= 0) {
            result = name.substring(slash + 1);
        }

        return result;
    }

    getEntityName() {
        return this.Title;
    }

    getId() {
        return this.UID;
    }

    getTitle() {
        return this.Title;
    }

    getGoogleMapsLink() {
        return "https://www.google.com/maps/place/" + this.Lat + "," + this.Lng;
    }

    refreshFileAttr() {
        const file = this.mainFile();

        if (!file || !file.Hash) {
            return;
        }

        this.Hash = file.Hash;
        this.Width = file.Width;
        this.Height = file.Height;
    }

    isPlayable() {
        if (!this.Files) {
            return false;
        }

        return this.Files.findIndex(f => f.Codec === CodecAvc1) !== -1 || this.Files.findIndex(f => f.Type === TypeMP4) !== -1;
    }

    videoFile() {
        if (!this.Files) {
            return false;
        }

        let file = this.Files.find(f => f.Codec === CodecAvc1);

        if (!file) {
            file = this.Files.find(f => f.Type === TypeMP4);
        }

        if (!file) {
            file = this.Files.find(f => !!f.Video);
        }

        return file;
    }

    videoUrl() {
        const file = this.videoFile();

        if (!file) {
            return "";
        }

        return `/api/v1/videos/${file.Hash}/${config.previewToken()}/${TypeMP4}`;
    }

    mainFile() {
        if (!this.Files) {
            return false;
        }

        let file = this.Files.find(f => !!f.Primary);

        if (!file) {
            file = this.Files.find(f => f.Type === TypeJpeg);
        }

        return file;
    }

    mainFileHash() {
        if (this.Files) {
            let file = this.mainFile();

            if (file && file.Hash) {
                return file.Hash;
            }
        } else if (this.Hash) {
            return this.Hash;
        }

        return "";
    }

    thumbnailUrl(type) {
        let hash = this.mainFileHash();

        if (!hash) {
            let video = this.videoFile();

            if (video && video.Hash) {
                return `/api/v1/t/${video.Hash}/${config.previewToken()}/${type}`;
            }

            return "/api/v1/svg/photo";
        }

        return `/api/v1/t/${hash}/${config.previewToken()}/${type}`;
    }

    getDownloadUrl() {
        return `/api/v1/dl/${this.mainFileHash()}?t=${config.downloadToken()}`;
    }

    downloadAll() {
        if (!this.Files) {
            let link = document.createElement("a");
            link.href = `/api/v1/dl/${this.mainFileHash()}?t=${config.downloadToken()}`;
            link.download = this.baseName(false);
            link.click();
            return;
        }

        this.Files.forEach((file) => {
            if (!file || !file.Hash) {
                console.warn("no file hash found for download", file);
                return;
            }

            let link = document.createElement("a");
            link.href = `/api/v1/dl/${file.Hash}?t=${config.downloadToken()}`;
            link.download = this.fileBase(file.Name);
            link.click();
        });
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
        if (!this.TakenAt || this.Year === YearUnknown) {
            return "Unknown";
        }

        if (this.TimeZone) {
            return DateTime.fromISO(this.TakenAt).setZone(this.TimeZone).toLocaleString(DateTime.DATETIME_FULL);
        }

        return DateTime.fromISO(this.TakenAt).setZone("UTC").toLocaleString(DateTime.DATE_HUGE);
    }

    shortDateString() {
        if (!this.TakenAt || this.Year === YearUnknown) {
            return "Unknown";
        }


        if (this.TimeZone) {
            return DateTime.fromISO(this.TakenAt).setZone(this.TimeZone).toLocaleString(DateTime.DATE_MED);
        }

        return DateTime.fromISO(this.TakenAt).setZone("UTC").toLocaleString(DateTime.DATE_MED);
    }

    hasLocation() {
        return this.Lat !== 0 || this.Lng !== 0;
    }

    locationInfo() {
        if (this.PlaceID === "zz" && this.Country !== "zz") {
            const country = countries.find(c => c.Code === this.Country);

            if(country) {
                return country.Name;
            }
        }

        return this.LocLabel ? this.LocLabel : "Unknown";
    }

    addSizeInfo(file, info) {
        if (!file) {
            return;
        }

        if (file.Width && file.Height) {
            info.push(file.Width + " × " + file.Height);
        } else if (!file.Primary) {
            let main = this.mainFile();
            if (main && main.Width && main.Height) {
                info.push(main.Width + " × " + main.Height);
            }
        }

        if (file.Size > 102400) {
            const size = Number.parseFloat(file.Size) / 1048576;

            info.push(size.toFixed(1) + " MB");
        } else if (file.Size) {
            const size = Number.parseFloat(file.Size) / 1024;

            info.push(size.toFixed(1) + " KB");
        }
    }

    getVideoInfo() {
        let info = [];
        let file = this.videoFile();

        if (!file) {
            file = this.mainFile();
        }

        if (!file) {
            return "Video";
        }

        if (file.Duration > 0) {
            info.push(Util.duration(file.Duration));
        }

        this.addSizeInfo(file, info);

        if (!info) {
            return "Video";
        }

        return info.join(", ");
    }

    getPhotoInfo() {
        let info = [];

        if (this.Camera) {
            info.push(this.Camera.Make + " " + this.Camera.Model);
        } else if (this.CameraModel && this.CameraMake) {
            info.push(this.CameraMake + " " + this.CameraModel);
        }

        let file = this.mainFile();

        this.addSizeInfo(file, info);

        if (!info) {
            return "Unknown";
        }

        return info.join(", ");
    }

    getCamera() {
        if (this.Camera) {
            return this.Camera.Make + " " + this.Camera.Model;
        } else if (this.CameraModel) {
            return this.CameraMake + " " + this.CameraModel;
        }

        return "Unknown";
    }

    toggleLike() {
        this.Favorite = !this.Favorite;

        if (this.Favorite) {
            return Api.post(this.getEntityResource() + "/like");
        } else {
            return Api.delete(this.getEntityResource() + "/like");
        }
    }

    togglePrivate() {
        this.Private = !this.Private;

        return Api.put(this.getEntityResource(), {Private: this.Private});
    }

    setPrimary(uid) {
        return Api.post(this.getEntityResource() + "/primary/" + uid).then((r) => Promise.resolve(this.setValues(r.data)));
    }

    like() {
        this.Favorite = true;
        return Api.post(this.getEntityResource() + "/like");
    }

    unlike() {
        this.Favorite = false;
        return Api.delete(this.getEntityResource() + "/like");
    }

    addLabel(name) {
        return Api.post(this.getEntityResource() + "/label", {Name: name, Priority: 10})
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    activateLabel(id) {
        return Api.put(this.getEntityResource() + "/label/" + id, {Uncertainty: 0})
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    renameLabel(id, name) {
        return Api.put(this.getEntityResource() + "/label/" + id, {Label: {Name: name}})
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    removeLabel(id) {
        return Api.delete(this.getEntityResource() + "/label/" + id)
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    update() {
        const values = this.getValues(true);

        if (values.Title) {
            values.TitleSrc = SrcManual;
        }

        if (values.Description) {
            values.DescriptionSrc = SrcManual;
        }

        if (values.Lat || values.Lng || values.Country) {
            values.LocSrc = SrcManual;
        }

        if (values.TakenAt || values.TimeZone) {
            values.TakenSrc = SrcManual;
        }

        if (values.CameraID || values.LensID || values.FocalLength || values.FNumber || values.Iso || values.Exposure) {
            values.CameraSrc = SrcManual;
        }

        return Api.put(this.getEntityResource(), values).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    static getCollectionResource() {
        return "photos";
    }

    static getModelName() {
        return "Photo";
    }

    static mergeResponse(results, response) {
        if (response.offset === 0 || results.length === 0) {
            return response.models;
        }

        if (response.models.length > 0) {
            let i = results.length - 1;

            if (results[i].UID === response.models[0].UID) {
                const first = response.models.shift();
                results[i].Files = results[i].Files.concat(first.Files);
            }
        }

        return results.concat(response.models);
    }
}

export default Photo;
