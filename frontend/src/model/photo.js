import Abstract from "model/abstract";
import Api from "common/api";
import truncate from "truncate";
import { DateTime } from "luxon";

class Photo extends Abstract {
    getEntityName() {
        return this.PhotoTitle;
    }

    getId() {
        return this.ID;
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

    getColors() {
        return this.PhotoColors;
    }

    getGoogleMapsLink() {
        return "https://www.google.com/maps/place/" + this.PhotoLat + "," + this.PhotoLong;
    }

    getThumbnailUrl(type) {
        return "/api/v1/thumbnails/" + this.FileHash + "/" + type;
    }

    getDownloadUrl() {
        return "/api/v1/download/" + this.FileHash;
    }

    getThumbnailSrcset() {
        const result = [];

        result.push(this.getThumbnailUrl("fit_720")  + " 720w");
        result.push(this.getThumbnailUrl("fit_1280") + " 1280w");
        result.push(this.getThumbnailUrl("fit_1920") + " 1920w");
        result.push(this.getThumbnailUrl("fit_2560") + " 2560w");
        result.push(this.getThumbnailUrl("fit_3840") + " 3840w");

        return result.join(", ");
    }

    calculateSize(width, height) {
        if(width >= this.FileWidth && height >= this.FileHeight) { // Smaller
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
        if(this.TimeZone) {
            return DateTime.fromISO(this.TakenAt).setZone(this.TimeZone).toLocaleString(DateTime.DATETIME_FULL);
        } else if(this.TakenAt) {
            return DateTime.fromISO(this.TakenAt).toLocaleString(DateTime.DATE_HUGE);
        } else {
            return "Unknown";
        }
    }

    hasLocation() {
        return this.PhotoLat !== 0 || this.PhotoLong !== 0;
    }

    getLocation() {
        const location = [];

        if (this.LocationID) {
            if (this.LocName && !this.LocCity && !this.LocCounty) {
                location.push(truncate(this.LocName, 20));
            } else if (this.LocCity && this.LocCity.length < 20) {
                location.push(this.LocCity);
            } else if (this.LocCounty && this.LocCity.length < 20) {
                location.push(this.LocCounty);
            }

            if (this.LocState && this.LocState !== this.LocCity) {
                location.push(this.LocState);
            }

            if (this.LocCountry) {
                location.push(this.LocCountry);
            }
        } else if (this.CountryName) {
            location.push(this.CountryName);
        } else {
            location.push("Unknown");
        }

        return location.join(", ");
    }

    getFullLocation() {
        const location = [];

        if (this.LocationID) {
            if (this.LocName) {
                location.push(this.LocName);
            }

            if (this.LocCity) {
                location.push(this.LocCity);
            }

            if (this.LocPostcode) {
                location.push(this.LocPostcode);
            }

            if (this.LocCounty) {
                location.push(this.LocCounty);
            }

            if (this.LocState) {
                location.push(this.LocState);
            }

            if (this.LocCountry) {
                location.push(this.LocCountry);
            }
        } else if (this.CountryName) {
            location.push(this.CountryName);
        } else {
            location.push("Unknown");
        }

        return location.join(", ");
    }

    getCamera() {
        if (this.CameraModel) {
            return this.CameraMake + " " + this.CameraModel;
        }

        return "Unknown";
    }

    toggleLike() {
        this.PhotoFavorite = !this.PhotoFavorite;

        if(this.PhotoFavorite) {
            return Api.post(this.getEntityResource() + "/like");
        } else {
            return Api.delete(this.getEntityResource() + "/like");
        }
    }

    like() {
        this.PhotoFavorite = true;
        return Api.post(this.getEntityResource() + "/like");
    }

    unlike() {
        this.PhotoFavorite = false;
        return Api.delete(this.getEntityResource() + "/like");
    }

    static getCollectionResource() {
        return "photos";
    }

    static getModelName() {
        return "Photo";
    }
}

export default Photo;
