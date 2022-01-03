/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

export const FormPropertyType = Object.freeze({
  String: "string",
  Number: "number",
  Object: "object",
});

export default class Form {
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
    } else if (typeof value != def[name].type) {
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
