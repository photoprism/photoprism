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

export const FormPropertyType = Object.freeze({
  String: "string",
  Number: "number",
  Object: "object",
});

export class Form {
  constructor(definition) {
    this.definition = definition;
  }

  setValues(values) {
    const def = this.getDefinition();

    for (let prop in def) {
      if (values.hasOwnProperty(prop)) {
        this.setValue(prop, values[prop]);
      }
    }

    return this;
  }

  getValues() {
    const result = {};
    const def = this.getDefinition();

    for (let prop in def) {
      result[prop] = this.getValue(prop);
    }

    return result;
  }

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

  getValue(name) {
    const def = this.getDefinition();

    if (def.hasOwnProperty(name)) {
      return def[name].value;
    } else {
      throw `Property ${name} not found`;
    }
  }

  setDefinition(definition) {
    this.definition = definition;
  }

  getDefinition() {
    return this.definition ? this.definition : {};
  }

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

export class rules {
  static maxLen(v, max) {
    if (!v || typeof v !== "string" || max <= 0) {
      return true;
    }

    return v.length <= max;
  }

  static minLen(v, min) {
    if (!v || typeof v !== "string" || min <= 0) {
      return true;
    }

    return v.length >= min;
  }

  static isLat(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    }

    const lat = Number(v);

    if (isNaN(lat)) {
      return false;
    }

    return -91 < lat < 91;
  }

  static isLng(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    }

    const lng = Number(v);

    if (isNaN(lng)) {
      return false;
    }

    return -181 < lng < 181;
  }

  static isNumber(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    }

    return !isNaN(Number(v));
  }

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

  static isTime(v) {
    return /^(2[0-3]|[0-1][0-9])\D[0-5][0-9]\D[0-5][0-9]$/.test(v); // 23:59:59
  }

  static isEmail(v) {
    if (typeof v !== "string" || v === "") {
      return true;
    } else if (!this.maxLen(v, 250)) {
      return false;
    }

    return /^\w+([.-]?\w+)*@\w+([.-]?\w+)*(\.\w{2,32})+$/.test(v);
  }

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

  static lat(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => this.isLat(v) || $gettext("Invalid")];
    } else {
      return [(v) => this.isLat(v) || $gettext("Invalid")];
    }
  }

  static lng(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => this.isLng(v) || $gettext("Invalid")];
    } else {
      return [(v) => this.isLng(v) || $gettext("Invalid")];
    }
  }

  static time(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => this.isTime(v) || $gettext("Invalid time")];
    } else {
      return [(v) => this.isTime(v) || $gettext("Invalid time")];
    }
  }

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

  static url(required) {
    if (required) {
      return [(v) => !!v || $gettext("This field is required"), (v) => !v || this.isUrl(v) || $gettext("Invalid URL")];
    } else {
      return [(v) => !v || this.isUrl(v) || $gettext("Invalid URL")];
    }
  }

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

  static number(required, min, max) {
    if (!min) {
      min = 0;
    }

    if (!max) {
      max = 2147483647;
    }

    if (required) {
      return [
        (v) => !!v || $gettext("This field is required"),
        (v) => (this.isNumber(v) && v >= min) || $gettext("Invalid"),
        (v) => (this.isNumber(v) && v <= max) || $gettext("Invalid"),
      ];
    } else {
      return [
        (v) => (this.isNumber(v) && v >= min) || $gettext("Invalid"),
        (v) => (this.isNumber(v) && v <= max) || $gettext("Invalid"),
      ];
    }
  }

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
