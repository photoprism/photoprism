import Event from "pubsub-js";

class Config {
    /**
     * @param {Storage} storage
     * @param {object} values
     */
    constructor(storage, values) {
        this.storage = storage;
        this.storage_key = "config";

        this.values = values;

        this.subscriptionId = Event.subscribe('config.updated', (ev, data) => this.setValues(data));
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

    getValue(key) {
        return this.values[key];
    }

    deleteValue(key) {
        delete this.values[key];

        return this;
    }
}

export default Config;
