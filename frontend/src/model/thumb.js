import Model from "./model";
import Api from "../common/api";

const thumbs = window.clientConfig.thumbnails;

class Thumb extends Model {
    getDefaults() {
        return {
            uuid: "",
            title: "",
            favorite: false,
            original_w: 0,
            original_h: 0,
            download_url: "",
        };
    }

    toggleLike() {
        this.favorite = !this.favorite;

        if (this.favorite) {
            return Api.post("photos/" + this.uuid + "/like");
        } else {
            return Api.delete("photos/" + this.uuid + "/like");
        }
    }

    static thumbNotFound() {
        const result = {
            uuid: "",
            title: "Not Found",
            favorite: false,
            original_w: 0,
            original_h: 0,
            download_url: "",
        };

        for (let i = 0; i < thumbs.length; i++) {
            result[thumbs[i].Name] = {
                src: "/api/v1/svg/photo",
                w: thumbs[i].Width,
                h: thumbs[i].Height,
            };
        }

        return result;
    }

    static fromPhotos(photos) {
        let result = [];

        photos.forEach((p) => {
            let thumb = this.fromPhoto(p);
            result.push(thumb);
        });

        return result;
    }

    static fromPhoto(photo) {
        if (photo.Files) {
            return this.fromFile(photo, photo.Files.find(f => !!f.FilePrimary));
        }

        if (!photo || !photo.FileHash) {
            return this.thumbNotFound();
        }

        const result = {
            uuid: photo.PhotoUUID,
            title: photo.PhotoTitle,
            favorite: photo.PhotoFavorite,
            download_url: this.downloadUrl(photo),
            original_w: photo.FileWidth,
            original_h: photo.FileHeight,
        };

        for (let i = 0; i < thumbs.length; i++) {
            let size = photo.calculateSize(thumbs[i].Width, thumbs[i].Height);

            result[thumbs[i].Name] = {
                src: photo.getThumbnailUrl(thumbs[i].Name),
                w: size.width,
                h: size.height,
            };
        }

        return new this(result);
    }

    static fromFile(photo, file) {
        if (!photo || !file || !file.FileHash) {
            return this.thumbNotFound();
        }

        const result = {
            uuid: photo.PhotoUUID,
            title: photo.PhotoTitle,
            favorite: photo.PhotoFavorite,
            download_url: this.downloadUrl(file),
            original_w: file.FileWidth,
            original_h: file.FileHeight,
        };

        thumbs.forEach((t) => {
            let size = this.calculateSize(file, t.Width, t.Height);

            result[t.Name] = {
                src: this.thumbnailUrl(file, t.Name),
                w: size.width,
                h: size.height,
            };
        });

        return new this(result);
    }

    static fromFiles(photos) {
        let result = [];

        photos.forEach((p) => {
            if (!p.Files) return;

            p.Files.forEach((f) => {
                if (f && f.FileType === "jpg") {
                    let thumb = this.fromFile(p, f);

                    if (thumb) {
                        result.push(thumb);
                    }
                }
            }
            );
        });

        return result;
    }

    static calculateSize(file, width, height) {
        if (width >= file.FileWidth && height >= file.FileHeight) { // Smaller
            return {width: file.FileWidth, height: file.FileHeight};
        }

        const srcAspectRatio = file.FileWidth / file.FileHeight;
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

    static thumbnailUrl(file, type) {
        if (!file.FileHash) {
            return "/api/v1/svg/photo";

        }

        return "/api/v1/thumbnails/" + file.FileHash + "/" + type;
    }

    static downloadUrl(file) {
        if (!file || !file.FileHash) {
            return "";
        }

        return "/api/v1/download/" + file.FileHash;
    }
}

export default Thumb;
