/*

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import Api from "common/api";
import Model from "./model";

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
    return Api.get("settings").then((response) => {
      return Promise.resolve(this.setValues(response.data));
    });
  }

  save() {
    return Api.post("settings", this.getValues(true)).then((response) =>
      Promise.resolve(this.setValues(response.data))
    );
  }
}

export default Settings;
