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

export class Model {
  constructor(values) {
    this.__originalValues = {};

    if (values) {
      this.setValues(values);
    } else {
      this.setValues(this.getDefaults());
    }
  }

  setValues(values, scalarOnly) {
    if (!values) return;

    if (values.maps?.style === "basic" || values.maps?.style === "offline") {
      values.maps.style = "";
    }

    for (let key in values) {
      if (values.hasOwnProperty(key) && key !== "__originalValues") {
        this[key] = values[key];

        if (typeof values[key] !== "object") {
          this.__originalValues[key] = values[key];
        } else if (!scalarOnly) {
          this.__originalValues[key] = JSON.parse(JSON.stringify(values[key]));
        }
      }
    }

    return this;
  }

  getValues(changed) {
    const result = {};
    const defaults = this.getDefaults();

    for (let key in this.__originalValues) {
      if (this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
        let val;
        if (defaults.hasOwnProperty(key)) {
          switch (typeof defaults[key]) {
            case "string":
              if (this[key] === null || this[key] === undefined) {
                val = "";
              } else {
                val = this[key];
              }
              break;
            case "bigint":
            case "number":
              val = parseFloat(this[key]);
              break;
            case "boolean":
              val = !!this[key];
              break;
            default:
              val = this[key];
          }
        } else {
          val = this[key];
        }

        if (!changed || JSON.stringify(val) !== JSON.stringify(this.__originalValues[key])) {
          result[key] = val;
        }
      }
    }

    return result;
  }

  originalValue(key) {
    if (this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
      return this.__originalValues[key];
    } else if (this.hasOwnProperty(key) && key !== "__originalValues") {
      return this[key];
    }

    return null;
  }

  wasChanged() {
    const changed = this.getValues(true);

    if (!changed) {
      return false;
    }

    return !(changed.constructor === Object && Object.keys(changed).length === 0);
  }

  getDefaults() {
    return {};
  }
}

export default Model;
