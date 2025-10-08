import $api from "common/api";
import Model from "./model";

// Settings stores the nested user/admin settings tree used across the UI.
export class Settings extends Model {
  changed(area, key) {
    if (typeof this.__originalValues[area] === "undefined") {
      return false;
    }

    return this[area][key] !== this.__originalValues[area][key];
  }

  setValues(values, scalarOnly) {
    if (!values) return;

    if (values.maps?.style === "basic" || values.maps?.style === "offline") {
      values.maps.style = "";
    }

    super.setValues(values, scalarOnly);

    return this;
  }

  load() {
    return $api.get("settings").then((response) => {
      return Promise.resolve(this.setValues(response.data));
    });
  }

  save() {
    return $api
      .post("settings", this.getValues(true))
      .then((response) => Promise.resolve(this.setValues(response.data)));
  }
}

export default Settings;
