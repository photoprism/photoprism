import Api from "common/api";
import Model from "./model";

export class Settings extends Model {
    changed(key) {
        return (this[key] !== this.__originalValues[key]);
    }

    load() {
        return Api.get("settings").then((response) => {
            return Promise.resolve(this.setValues(response.data));
        });
    }

    save() {
        return Api.post("settings", this.getValues(true)).then((response) => Promise.resolve(this.setValues(response.data)));
    }
}

export default Settings;
