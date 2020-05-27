import Model from "./model";
import Api from "../common/api";
import {config} from "../session";

const thumbs = window.__CONFIG__.thumbnails;

export class Thumb extends Model {
    getDefaults() {
        return {
            uid: "",
            title: "",
            taken: "",
            description: "",
            favorite: false,
            playable: false,
            original_w: 0,
            original_h: 0,
            download_url: "",
        };
    }

    toggleLike() {
        this.favorite = !this.favorite;

        if (this.favorite) {
            return Api.post("photos/" + this.uid + "/like");
        } else {
            return Api.delete("photos/" + this.uid + "/like");
        }
    }

    static thumbNotFound() {
        const result = {
            uid: "",
            title: "Not Found",
            taken: "",
            description: "",
            favorite: false,
            playable: false,
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
            return this.fromFile(photo, photo.Files.find(f => !!f.Primary));
        }

        if (!photo || !photo.Hash) {
            return this.thumbNotFound();
        }

        const result = {
            uid: photo.UID,
            title: photo.Title,
            taken: photo.getDateString(),
            description: photo.Description,
            favorite: photo.Favorite,
            playable: photo.isPlayable(),
            download_url: this.downloadUrl(photo),
            original_w: photo.Width,
            original_h: photo.Height,
        };

        for (let i = 0; i < thumbs.length; i++) {
            let size = photo.calculateSize(thumbs[i].Width, thumbs[i].Height);

            result[thumbs[i].Name] = {
                src: photo.thumbnailUrl(thumbs[i].Name),
                w: size.width,
                h: size.height,
            };
        }

        return new this(result);
    }

    static fromFile(photo, file) {
        if (!photo || !file || !file.Hash) {
            return this.thumbNotFound();
        }

        const result = {
            uid: photo.UID,
            title: photo.Title,
            taken: photo.getDateString(),
            description: photo.Description,
            favorite: photo.Favorite,
            playable: photo.isPlayable(),
            download_url: this.downloadUrl(file),
            original_w: file.Width,
            original_h: file.Height,
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
                if (f && f.Type === "jpg") {
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
        if (width >= file.Width && height >= file.Height) { // Smaller
            return {width: file.Width, height: file.Height};
        }

        const srcAspectRatio = file.Width / file.Height;
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
        if (!file.Hash) {
            return "/api/v1/svg/photo";

        }

        return `/api/v1/t/${file.Hash}/${config.previewToken()}/${type}`;
    }

    static downloadUrl(file) {
        if (!file || !file.Hash) {
            return "";
        }

        return `/api/v1/dl/${file.Hash}?t=${config.downloadToken()}`;
    }
}

export default Thumb;
