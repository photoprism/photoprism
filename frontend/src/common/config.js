import Api from "common/api";

class Config {
    /**
     * @param {Storage} storage
     * @param {object} values
     */
    constructor(storage, values) {
        this.storage = storage;
        this.storage_key = "config";

        this.values = values;

        // this.setValues(JSON.parse(this.storage.getItem(this.storage_key)));
        // this.setValues(values);
    }

    setValues(values) {
        if(!values) return;

        for(let key in values) {
            if(values.hasOwnProperty(key)) {
                this.setValue[key] = values[key];
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

    pullFromServer() {
        return Api.get("config").then(
            (result) => {
                this.setValues(result.data);
            }
        );
    }
}

export default Config;
