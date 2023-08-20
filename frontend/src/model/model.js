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

  /**
   * Returns the values of this model, all of them or only changed values, with changed
   * subobjects traversed for changed fields or simply included as complete objects
   *
   * @param {boolean} changed - get only changed values
   * @param {boolean} traverseSubObjects - get only changed values for subobjects too (parameter {@link changed} must also be `true`)
   * @return {object}
   */
  getValues(changed, traverseSubObjects) {
    return getObjectValues(this, this.__originalValues, this.getDefaults());

    function getObjectValues(obj, originalValues, defaults) {
      const result = {};
      for (let key in originalValues) {
        if (originalValues.hasOwnProperty(key) && key !== "__originalValues") {
          let val;
          if (defaults.hasOwnProperty(key)) {
            val = getTypedValue(key);
          } else {
            val = obj[key];
          }

          if (!changed || JSON.stringify(val) !== JSON.stringify(originalValues[key])) {
            if (changed && traverseSubObjects && typeof val === "object") {
              result[key] = getObjectValues(val, originalValues[key], defaults[key]);
            } else {
              result[key] = val;
            }
          }
        }
      }

      return result;

      function getTypedValue(key) {
        let typedVal;
        switch (typeof defaults[key]) {
          case "string":
            if (obj[key] === null || obj[key] === undefined) {
              typedVal = "";
            } else {
              typedVal = obj[key];
            }
            break;
          case "bigint":
          case "number":
            typedVal = parseFloat(obj[key]);
            break;
          case "boolean":
            typedVal = !!obj[key];
            break;
          default:
            typedVal = obj[key];
        }
        return typedVal;
      }
    }
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
