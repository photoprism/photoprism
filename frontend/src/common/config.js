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

        Event.subscribe("config.updated", (ev, data) => this.setValues(data));
        Event.subscribe("count", (ev, data) => this.onCount(ev, data));

        if(this.hasValue("settings")) {
            this.setTheme(this.getValue("settings").theme);
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
                this.setValue(key, values[key]);
            }
        }

        return this;
    }

    onCount(ev, data) {
        const type = ev.split(".")[1];

        switch (type) {
        case "favorites":
            this.values.count.favorites += data.count;
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

    updateSettings(settings, $vuetify) {
        this.setValue("settings", settings);
        this.setTheme(settings.theme);
        $vuetify.theme = this.theme;
    }

    setTheme(name) {
        this.theme = themes[name] ? themes[name] : themes["default"];
        return this;
    }

    getValues() {
        return this.values;
    }

    storeValues() {
        this.storage.setItem(this.storage_key, JSON.stringify(this.getValues()));

        return this;
    }

    setValue(key, value) {
        this.values[key] = value;

        return this;
    }

    hasValue(key) {
        return !!this.values[key];
    }

    getValue(key) {
        return this.values[key];
    }

    deleteValue(key) {
        delete this.values[key];

        return this;
    }

    feature(name) {
        return this.values.settings.features[name];
    }

    settings() {
        return this.values.settings
    }
}

export default Config;
