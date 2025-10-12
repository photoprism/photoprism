/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

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

import { $gettext } from "common/gettext";

// FormPropertyType enumerates supported types for form properties.
export const FormPropertyType = Object.freeze({
  String: "string",
  Number: "number",
  Object: "object",
});

// Form encapsulates a simple key/value form definition with helpers for type-checked assignments.
export class Form {
  // constructor optionally accepts an initial definition.
  constructor(definition) {
    this.definition = definition;
  }

  // setValues assigns values in bulk while respecting the schema.
  setValues(values) {
    const def = this.getDefinition();

    for (let prop in def) {
      if (values.hasOwnProperty(prop)) {
        this.setValue(prop, values[prop]);
      }
    }

    return this;
  }

  // getValues returns a map of current values.
  getValues() {
    const result = {};
    const def = this.getDefinition();

    for (let prop in def) {
      result[prop] = this.getValue(prop);
    }

    return result;
  }

  // setValue updates a single value ensuring the type matches the definition.
  setValue(name, value) {
    const def = this.getDefinition();

    if (!def.hasOwnProperty(name)) {
      throw `Property ${name} not found`;
    } else if (typeof value !== def[name].type) {
      throw `Property ${name} must be ${def[name].type}`;
    } else {
      def[name].value = value;
    }

    return this;
  }

  // getValue fetches a single property value.
  getValue(name) {
    const def = this.getDefinition();

    if (def.hasOwnProperty(name)) {
      return def[name].value;
    } else {
      throw `Property ${name} not found`;
    }
  }

  // setDefinition replaces the current form schema.
  setDefinition(definition) {
    this.definition = definition;
  }

  // getDefinition returns the current schema or an empty object.
  getDefinition() {
    return this.definition ? this.definition : {};
  }

  // getOptions resolves the options array for select-style fields.
  getOptions(fieldName) {
    if (
      this.definition &&
      this.definition.hasOwnProperty(fieldName) &&
      this.definition[fieldName].hasOwnProperty("options")
    ) {
      return this.definition[fieldName].options;
    }

    return [{ option: "", label: "" }];
  }
}

// rules centralizes reusable validation helpers and Vuetify rule factories used across the UI.
export class rules {
  // maxLen ensures that a string does not exceed the provided maximum length.
  static maxLen(v, max) {
    if (!v || typeof v !== "string" || max <= 0) {
      return true;
    }

    return v.length <= max;
  }

  // minLen ensures that a string meets the minimum length.
  static minLen(v, min) {
    if (!v || typeof v !== "string" || min <= 0) {
      return true;
    }

    return v.length >= min;
  }

  // isLat validates latitude values in decimal degrees.
  static isLat(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    }

    const lat = Number(v);

    if (isNaN(lat)) {
      return false;
    }

    return lat >= -90 && lat <= 90;
  }

  // isLng validates longitude values in decimal degrees.
  static isLng(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    }

    const lng = Number(v);

    if (isNaN(lng)) {
      return false;
    }

    return lng >= -180 && lng <= 180;
  }

  // isNumber validates that a value is a parsable number or empty.
  static isNumber(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    }

    return !isNaN(Number(v));
  }

  // isNumberRange validates numeric strings within optional inclusive bounds.
  static isNumberRange(v, min, max) {
    if (typeof v !== "string" || !v || v === "-1") {
      return true;
    }

    v = Number(v);

    if (isNaN(v)) {
      return false;
    }

    if (typeof min === "number" && v < min) {
      return false;
    }

    if (typeof max === "number" && v > max) {
      return false;
    }

    return true;
  }

  // isTime validates HH:MM:SS style times with any non-digit separator.
  static isTime(v) {
    return /^(2[0-3]|[0-1][0-9])\D[0-5][0-9]\D[0-5][0-9]$/.test(v); // 23:59:59
  }

  // isEmail verifies that strings match the backend email sanitizer rules while staying lenient for empty inputs.
  static isEmail(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    } else if (!this.maxLen(v, 250)) {
      return false;
    }

    return /^[A-Za-z0-9.!#$%&'*+/=?^_`{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$/.test(
      v
    );
  }

  // isUrl validates strings by length and URL parsing.
  static isUrl(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    } else if (!this.maxLen(v, 500)) {
      return false;
    }

    try {
      new URL(v);
    } catch (e) {
      return false;
    }
    return true;
  }

  // lat returns Vuetify rule callbacks for latitude validation.
  static lat(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => this.isLat(v) || $gettext("Invalid")];
    } else {
      return [(v) => this.isLat(v) || $gettext("Invalid")];
    }
  }

  // lng returns Vuetify rule callbacks for longitude validation.
  static lng(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => this.isLng(v) || $gettext("Invalid")];
    } else {
      return [(v) => this.isLng(v) || $gettext("Invalid")];
    }
  }

  // time returns Vuetify rule callbacks enforcing HH:MM:SS format.
  static time(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => this.isTime(v) || $gettext("Invalid time")];
    } else {
      return [(v) => !v || this.isTime(v) || $gettext("Invalid time")];
    }
  }

  // email returns Vuetify rule callbacks for email validation.
  static email(required) {
    if (required) {
      return [
        (v) => !!v || $gettext("This field is required"),
        (v) => !v || this.isEmail(v) || $gettext("Invalid address"),
      ];
    } else {
      return [(v) => !v || this.isEmail(v) || $gettext("Invalid address")];
    }
  }

  // url returns Vuetify rule callbacks for URL validation.
  static url(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => !v || this.isUrl(v) || $gettext("Invalid URL")];
    } else {
      return [(v) => !v || this.isUrl(v) || $gettext("Invalid URL")];
    }
  }

  // text returns string length validators with optional localization label.
  static text(required, min, max, s) {
    if (!s) {
      s = $gettext("Text");
    }

    if (required) {
      return [
        (v) => !!v || $gettext("This field is required"),
        (v) => this.minLen(v, min ? min : 0) || $gettext(`%{s} is too short`, { s }),
        (v) => this.maxLen(v, max ? max : 200) || $gettext("%{s} is too long", { s }),
      ];
    } else {
      return [
        (v) => this.minLen(v, min ? min : 0) || $gettext("%{s} is too short", { s }),
        (v) => this.maxLen(v, max ? max : 200) || $gettext("%{s} is too long", { s }),
      ];
    }
  }

  // number returns numeric validators with inclusive min/max checks.
  static number(required, min, max) {
    if (!min) {
      min = 0;
    }

    if (!max) {
      max = 2147483647;
    }

    const minValidator = (v) => {
      if (v === "" || v === undefined || v === null) {
        return true;
      }

      const value = Number(v);

      if (isNaN(value)) {
        return $gettext("Invalid");
      }

      return value >= min || $gettext("Invalid");
    };

    const maxValidator = (v) => {
      if (v === "" || v === undefined || v === null) {
        return true;
      }

      const value = Number(v);

      if (isNaN(value)) {
        return $gettext("Invalid");
      }

      return value <= max || $gettext("Invalid");
    };

    if (required) {
      return [(v) => !!v || $gettext("This field is required"), minValidator, maxValidator];
    } else {
      return [minValidator, maxValidator];
    }
  }

  // country validates ISO-style country codes via length checks.
  static country(required) {
    if (required) {
      return [
        (v) => !!v || $gettext("This field is required"),
        (v) => this.minLen(v, 2) || $gettext("Invalid country"),
        (v) => this.maxLen(v, 2) || $gettext("Invalid country"),
      ];
    } else {
      return [
        (v) => this.minLen(v, 2) || $gettext("Invalid country"),
        (v) => this.maxLen(v, 2) || $gettext("Invalid country"),
      ];
    }
  }

  // day validates day-of-month values between 1 and 31.
  static day(required) {
    if (required) {
      return [
        (v) => !!v || Number(v) < -1 || $gettext("This field is required"),
        (v) => this.isNumberRange(v, 1, 31) || $gettext("Invalid"),
      ];
    } else {
      return [(v) => this.isNumberRange(v, 1, 31) || $gettext("Invalid")];
    }
  }

  // month validates month values between 1 and 12.
  static month(required) {
    if (required) {
      return [
        (v) => !!v || Number(v) < -1 || $gettext("This field is required"),
        (v) => this.isNumberRange(v, 1, 12) || $gettext("Invalid"),
      ];
    } else {
      return [(v) => this.isNumberRange(v, 1, 12) || $gettext("Invalid")];
    }
  }

  // year validates year values using optional bounds (defaults 1800..current year).
  static year(required, min, max) {
    if (!min) {
      min = 1800;
    }

    if (!max) {
      max = new Date().getFullYear();
    }

    if (required) {
      return [
        (v) => !!v || Number(v) < -1 || $gettext("This field is required"),
        (v) => this.isNumberRange(v, min, max) || $gettext("Invalid"),
      ];
    } else {
      return [(v) => this.isNumberRange(v, min, max) || $gettext("Invalid")];
    }
  }
}
