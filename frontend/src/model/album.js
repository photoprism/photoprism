/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import RestModel from "model/rest";
import Api from "common/api";
import {DateTime} from "luxon";
import {config} from "../session";
import {$gettext} from "common/vm";

export class Album extends RestModel {
    getDefaults() {
        return {
            UID: "",
            Cover: "",
            Parent: "",
            Folder: "",
            Slug: "",
            Type: "",
            Title: "",
            Location: "",
            Caption: "",
            Category: "",
            Description: "",
            Notes: "",
            Filter: "",
            Order: "",
            Template: "",
            Country: "",
            Day: -1,
            Year: -1,
            Month: -1,
            Favorite: true,
            Private: false,
            PhotoCount: 0,
            LinkCount: 0,
            CreatedAt: "",
            UpdatedAt: "",
        };
    }

    getEntityName() {
        return this.Slug;
    }

    getTitle() {
        return this.Title;
    }

    thumbnailUrl(size) {
        return `/api/v1/albums/${this.getId()}/t/${config.previewToken()}/${size}`;
    }

    dayString() {
        if (!this.Day || this.Day <= 0) {
            return "01";
        }

        return this.Day.toString().padStart(2, "0");
    }

    monthString() {
        if (!this.Month || this.Month <= 0) {
            return "01";
        }

        return this.Month.toString().padStart(2, "0");
    }

    yearString() {
        if (!this.Year || this.Year <= 1000) {
            return new Date().getFullYear().toString().padStart(4, "0");
        }

        return this.Year.toString();
    }

    getDate() {
        let date = this.yearString() + "-" + this.monthString() + "-" + this.dayString();

        return DateTime.fromISO(`${date}T12:00:00Z`).toUTC();
    }

    localDate(time) {
        if(!this.TakenAtLocal) {
            return this.utcDate();
        }

        let zone = this.getTimeZone();

        return DateTime.fromISO(this.localDateString(time), {zone});
    }

    getDateString() {
        if (!this.Year || this.Year <= 1000) {
            return $gettext("Unknown");
        } else if (!this.Month || this.Month <= 0) {
            return this.localYearString();
        } else if (!this.Day || this.Day <= 0) {
            return this.getDate().toLocaleString({month: "long", year: "numeric"});
        }

        return this.localDate().toLocaleString(DateTime.DATE_HUGE);
    }

    getCreatedString() {
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
        return "albums";
    }

    static getModelName() {
        return $gettext("Album");
    }
}

export default Album;
