import Event from "pubsub-js";
import themes from "../resources/themes.json";
import translations from "../resources/translations.json";
import Vue from "vue";

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

        this.subscriptionId = Event.subscribe('config.updated', (ev, data) => this.setValues(data));

        if(this.hasValue("settings")) {
            this.setTheme(this.getValue("settings").theme);
        } else {
            this.setTheme("default");
        }
    }

    setValues(values) {
        if (!values) return;

        for (let key in values) {
            if (values.hasOwnProperty(key)) {
                this.setValue(key, values[key]);
            }
        }

        return this;
    }

    updateSettings(values, $vuetify) {
        this.setValue("settings", values);
        this.setTheme(values.theme);
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
}

export default Config;
