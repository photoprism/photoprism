import Api from "common/api";

class Settings {
    constructor(values) {
        this.__originalValues = {};

        if (!values) {
            throw "can't create settings with empty values"
        }

        this.setValues(values);
    }

    setValues(values) {
        if(!values) return;

        for(let key in values) {
            if(values.hasOwnProperty(key) && key !== "__originalValues") {
                this[key] = values[key];
                this.__originalValues[key] = values[key];
            }
        }

        return this;
    }

    getValues() {
        const result = {};

        for(let key in this.__originalValues) {
            if(this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
                result[key] = this[key];
            }
        }

        return result;
    }

    load() {
        return Api.get("settings").then((response) => {
            return Promise.resolve(this.setValues(response.data));
        });
    }

    save() {
        return Api.post("settings", this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
    }
}

export default Settings;
