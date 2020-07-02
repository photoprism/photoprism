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
            Caption: "",
            Category: "",
            Description: "",
            Notes: "",
            Filter: "",
            Order: "",
            Template: "",
            Country: "",
            Year: 0,
            Month: 0,
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

    thumbnailUrl(type) {
        return `/api/v1/albums/${this.getId()}/t/${config.previewToken()}/${type}`;
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
        return "albums";
    }

    static getModelName() {
        return $gettext("Album");
    }
}

export default Album;
