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

import Event from "pubsub-js";
import themes from "options/themes.json";
import translations from "locales/translations.json";
import Api from "./api";

export default class Config {
    /**
     * @param {Storage} storage
     * @param {object} values
     */
    constructor(storage, values) {
        this.disconnected = false;
        this.storage = storage;
        this.storage_key = "config";

        this.$vuetify = null;
        this.translations = translations;

        if (!values || !values.siteTitle) {
            console.warn("config: values are empty");
            this.debug = true;
            this.values = {};
            this.page = {
                title: "PhotoPrism",
            };
            return;
        }

        this.page = {
            title: values.siteTitle,
        };

        this.values = values;
        this.debug = !!values.debug;

        Event.subscribe("config.updated", (ev, data) => this.setValues(data.config));
        Event.subscribe("count", (ev, data) => this.onCount(ev, data));

        if (this.has("settings")) {
            this.setTheme(this.get("settings").theme);
        } else {
            this.setTheme("default");
        }
    }

    update() {
        Api.get("config").then(
            (response) => this.setValues(response.data),
            () => console.warn("failed pulling updated client config")
        );
    }

    setValues(values) {
        if (!values) return;

        if (this.debug) {
            console.log("config: new values", values);
        }

        if(values.jsHash && this.values.jsHash !== values.jsHash) {
            Event.publish("dialog.reload", {values});
        }

        for (let key in values) {
            if (values.hasOwnProperty(key)) {
                this.set(key, values[key]);
            }
        }

        if (values.settings) {
            this.setTheme(values.settings.theme);
        }

        return this;
    }

    onCount(ev, data) {
        const type = ev.split(".")[1];

        switch (type) {
            case "cameras":
                this.values.count.cameras += data.count;
                this.update();
                break;
            case "lenses":
                this.values.count.lenses += data.count;
                break;
            case "countries":
                this.values.count.countries += data.count;
                this.update();
                break;
            case "states":
                this.values.count.states += data.count;
                break;
            case "places":
                this.values.count.places += data.count;
                break;
            case "labels":
                this.values.count.labels += data.count;
                break;
            case "videos":
                this.values.count.videos += data.count;
                break;
            case "albums":
                this.values.count.albums += data.count;
                break;
            case "moments":
                this.values.count.moments += data.count;
                break;
            case "months":
                this.values.count.months += data.count;
                break;
            case "folders":
                this.values.count.folders += data.count;
                break;
            case "files":
                this.values.count.files += data.count;
                break;
            case "favorites":
                this.values.count.favorites += data.count;
                break;
            case "review":
                this.values.count.review += data.count;
                break;
            case "private":
                this.values.count.private += data.count;
                break;
            case "photos":
                this.values.count.photos += data.count;
                break;
            default:
                console.warn("unknown count type", ev, data);
        }

        this.values.count;
    }

    setVuetify(instance) {
        this.$vuetify = instance;
    }

    setTheme(name) {
        this.theme = themes[name] ? themes[name] : themes["default"];

        if (this.$vuetify) {
            this.$vuetify.theme = this.theme;
        }

        return this;
    }

    getValues() {
        return this.values;
    }

    storeValues() {
        this.storage.setItem(this.storage_key, JSON.stringify(this.getValues()));
        return this;
    }

    set(key, value) {
        this.values[key] = value;
        return this;
    }

    has(key) {
        return !!this.values[key];
    }

    get(key) {
        return this.values[key];
    }

    feature(name) {
        return this.values.settings.features[name];
    }

    settings() {
        return this.values.settings;
    }

    downloadToken() {
        return this.values["downloadToken"];
    }

    previewToken() {
        return this.values["previewToken"];
    }

    albumCategories() {
        if (this.values["albumCategories"]) {
            return this.values["albumCategories"];
        }

        return [];
    }
}
