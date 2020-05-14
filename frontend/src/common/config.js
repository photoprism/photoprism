import Event from "pubsub-js";
import themes from "../resources/themes.json";
import translations from "../resources/translations.json";

class Config {
    /**
     * @param {Storage} storage
     * @param {object} values
     */
    constructor(storage, values) {
        this.storage = storage;
        this.storage_key = "config";

        this.translations = translations;
        this.values = values;
        this.debug = !!values.debug;
        this.page = {
            title: "PhotoPrism",
        };

        this.$vuetify = null;

        Event.subscribe("config.updated", (ev, data) => this.setValues(data));
        Event.subscribe("count", (ev, data) => this.onCount(ev, data));

        if (this.has("settings")) {
            this.setTheme(this.get("settings").theme);
        } else {
            this.setTheme("default");
        }
    }

    setValues(values) {
        if (!values) return;

        if (this.debug) {
            console.log("new config values", values);
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
        case "videos":
            this.values.count.videos += data.count;
            break;
        case "favorites":
            this.values.count.favorites += data.count;
            break;
        case "private":
            this.values.count.private += data.count;
            break;
        case "albums":
            this.values.count.albums += data.count;
            break;
        case "photos":
            this.values.count.photos += data.count;
            break;
        case "countries":
            this.values.count.countries += data.count;
            break;
        case "places":
            this.values.count.places += data.count;
            break;
        case "labels":
            this.values.count.labels += data.count;
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
}

export default Config;
