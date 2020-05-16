import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";
import Util from "common/util";

const SrcManual = "manual";
const CodecAvc1 = "avc1";
const TypeMP4 = "mp4";
const TypeJpeg = "jpg";
const YearUnknown = -1;
const MonthUnknown = -1;

class Photo extends RestModel {
    getDefaults() {
        return {
            ID: 0,
            TakenAt: "",
            TakenAtLocal: "",
            TakenSrc: "",
            TimeZone: "",
            PhotoUUID: "",
            PhotoPath: "",
            PhotoName: "",
            PhotoTitle: "",
            TitleSrc: "",
            PhotoFavorite: false,
            PhotoPrivate: false,
            PhotoVideo: false,
            PhotoResolution: 0,
            PhotoQuality: 0,
            PhotoLat: 0.0,
            PhotoLng: 0.0,
            PhotoAltitude: 0,
            PhotoIso: 0,
            PhotoFocalLength: 0,
            PhotoFNumber: 0.0,
            PhotoExposure: "",
            PhotoViews: 0,
            Camera: {},
            CameraID: 0,
            CameraSrc: "",
            Lens: {},
            LensID: 0,
            Location: null,
            LocationID: "",
            LocationSrc: "",
            Place: null,
            PlaceID: "",
            PhotoCountry: "",
            PhotoYear: YearUnknown,
            PhotoMonth: MonthUnknown,
            Description: {
                PhotoDescription: "",
                PhotoKeywords: "",
                PhotoNotes: "",
                PhotoSubject: "",
                PhotoArtist: "",
                PhotoCopyright: "",
                PhotoLicense: "",
            },
            DescriptionSrc: "",
            Files: [],
            Labels: [],
            Keywords: [],
            Albums: [],
            Links: [],
            CreatedAt: "",
            UpdatedAt: "",
            DeletedAt: null,
        };
    }

    getEntityName() {
        return this.PhotoTitle;
    }

    getId() {
        return this.PhotoUUID;
    }

    getTitle() {
        return this.PhotoTitle;
    }

    getColor() {
        switch (this.PhotoColor) {
        case "brown":
        case "black":
        case "white":
        case "grey":
            return "grey lighten-2";
        default:
            return this.PhotoColor + " lighten-4";
        }
    }

    getGoogleMapsLink() {
        return "https://www.google.com/maps/place/" + this.PhotoLat + "," + this.PhotoLng;
    }

    refreshFileAttr() {
        const file = this.mainFile();

        if (!file || !file.FileHash) {
            return;
        }

        this.FileHash = file.FileHash;
        this.FileWidth = file.FileWidth;
        this.FileHeight = file.FileHeight;
    }

    isPlayable() {
        if (!this.Files) {
            return false;
        }

        return this.Files.findIndex(f => f.FileCodec === CodecAvc1) !== -1 || this.Files.findIndex(f => f.FileType === TypeMP4) !== -1;
    }

    videoFile() {
        if (!this.Files) {
            return false;
        }

        let file = this.Files.find(f => f.FileCodec === CodecAvc1);

        if (!file) {
            file = this.Files.find(f => f.FileType === TypeMP4);
        }

        if (!file) {
            file = this.Files.find(f => !!f.FileVideo);
        }

        return file;
    }

    videoUri() {
        const file = this.videoFile();

        if (!file) {
            return "";
        }

        return "/api/v1/videos/" + file.FileHash + "/" + TypeMP4;
    }

    mainFile() {
        if (!this.Files) {
            return false;
        }

        let file = this.Files.find(f => !!f.FilePrimary);

        if (!file) {
            file = this.Files.find(f => f.FileType === TypeJpeg);
        }

        return file;
    }

    mainFileHash() {
        if (this.Files) {
            let file = this.mainFile();

            if (file && file.FileHash) {
                return file.FileHash;
            }
        } else if (this.FileHash) {
            return this.FileHash;
        }

        return "";
    }

    getThumbnailUrl(type) {
        let hash = this.mainFileHash();

        if (!hash) {
            let video = this.videoFile();

            if (video && video.FileHash) {
                return "/api/v1/thumbnails/" + video.FileHash + "/" + type;
            }

            return "/api/v1/svg/photo";
        }

        return "/api/v1/thumbnails/" + hash + "/" + type;
    }

    getDownloadUrl() {
        return "/api/v1/download/" + this.mainFileHash();
    }

    getThumbnailSrcset() {
        const result = [];

        result.push(this.getThumbnailUrl("fit_720") + " 720w");
        result.push(this.getThumbnailUrl("fit_1280") + " 1280w");
        result.push(this.getThumbnailUrl("fit_1920") + " 1920w");
        result.push(this.getThumbnailUrl("fit_2560") + " 2560w");
        result.push(this.getThumbnailUrl("fit_3840") + " 3840w");

        return result.join(", ");
    }

    calculateSize(width, height) {
        if (width >= this.FileWidth && height >= this.FileHeight) { // Smaller
            return {width: this.FileWidth, height: this.FileHeight};
        }

        const srcAspectRatio = this.FileWidth / this.FileHeight;
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

    getThumbnailSizes() {
        const result = [];

        result.push("(min-width: 2560px) 3840px");
        result.push("(min-width: 1920px) 2560px");
        result.push("(min-width: 1280px) 1920px");
        result.push("(min-width: 720px) 1280px");
        result.push("720px");

        return result.join(", ");
    }

    getDateString() {
        if (!this.TakenAt || this.PhotoYear === YearUnknown) {
            return "Unknown";
        }

        if (this.TimeZone) {
            return DateTime.fromISO(this.TakenAt).setZone(this.TimeZone).toLocaleString(DateTime.DATETIME_FULL);
        }

        return DateTime.fromISO(this.TakenAt).setZone("UTC").toLocaleString(DateTime.DATE_HUGE);
    }

    shortDateString() {
        if (!this.TakenAt || this.PhotoYear === YearUnknown) {
            return "Unknown";
        }


        if (this.TimeZone) {
            return DateTime.fromISO(this.TakenAt).setZone(this.TimeZone).toLocaleString(DateTime.DATE_MED);
        }

        return DateTime.fromISO(this.TakenAt).setZone("UTC").toLocaleString(DateTime.DATE_MED);
    }

    hasLocation() {
        return this.PhotoLat !== 0 || this.PhotoLng !== 0;
    }

    getLocation() {
        if (this.LocLabel) {
            return this.LocLabel;
        }

        return "Unknown";
    }

    addSizeInfo(file, info) {
        if (!file) {
            return;
        }

        if (file.FileWidth && file.FileHeight) {
            info.push(file.FileWidth + " × " + file.FileHeight);
        } else if (!file.FilePrimary) {
            let main = this.mainFile();
            if (main && main.FileWidth && main.FileHeight) {
                info.push(main.FileWidth + " × " + main.FileHeight);
            }
        }

        if (file.FileSize > 102400) {
            const size = Number.parseFloat(file.FileSize) / 1048576;

            info.push(size.toFixed(1) + " MB");
        } else if (file.FileSize) {
            const size = Number.parseFloat(file.FileSize) / 1024;

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

        if (file.FileDuration > 0) {
            info.push(Util.duration(file.FileDuration));
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
            info.push(this.Camera.CameraMake + " " + this.Camera.CameraModel);
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
            return this.Camera.CameraMake + " " + this.Camera.CameraModel;
        } else if (this.CameraModel) {
            return this.CameraMake + " " + this.CameraModel;
        }

        return "Unknown";
    }

    toggleLike() {
        this.PhotoFavorite = !this.PhotoFavorite;

        if (this.PhotoFavorite) {
            return Api.post(this.getEntityResource() + "/like");
        } else {
            return Api.delete(this.getEntityResource() + "/like");
        }
    }

    togglePrivate() {
        this.PhotoPrivate = !this.PhotoPrivate;

        return Api.put(this.getEntityResource(), {PhotoPrivate: this.PhotoPrivate});
    }

    setPrimary(fileUUID) {
        return Api.post(this.getEntityResource() + "/primary/" + fileUUID).then((r) => Promise.resolve(this.setValues(r.data)));
    }

    like() {
        this.PhotoFavorite = true;
        return Api.post(this.getEntityResource() + "/like");
    }

    unlike() {
        this.PhotoFavorite = false;
        return Api.delete(this.getEntityResource() + "/like");
    }

    addLabel(name) {
        return Api.post(this.getEntityResource() + "/label", {LabelName: name, LabelPriority: 10})
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    activateLabel(id) {
        return Api.put(this.getEntityResource() + "/label/" + id, {Uncertainty: 0})
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    renameLabel(id, name) {
        return Api.put(this.getEntityResource() + "/label/" + id, {Label: {LabelName: name}})
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    removeLabel(id) {
        return Api.delete(this.getEntityResource() + "/label/" + id)
            .then((r) => Promise.resolve(this.setValues(r.data)));
    }

    update() {
        const values = this.getValues(true);

        if (values.PhotoTitle) {
            values.TitleSrc = SrcManual;
        }

        if (values.Description) {
            values.DescriptionSrc = SrcManual;
        }

        if (values.PhotoLat || values.PhotoLng) {
            values.LocationSrc = SrcManual;
        }

        if (values.TakenAt || values.TimeZone) {
            values.TakenSrc = SrcManual;
        }

        if (values.CameraID || values.LensID || values.PhotoFocalLength || values.PhotoFNumber || values.PhotoIso || values.PhotoExposure) {
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

            if (results[i].PhotoUUID === response.models[0].PhotoUUID) {
                const first = response.models.shift();
                results[i].Files = results[i].Files.concat(first.Files);
            }
        }

        return results.concat(response.models);
    }
}

export default Photo;
