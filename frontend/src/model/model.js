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

// Model is the base class every domain model in `frontend/src/model/`
// extends. It provides setValues/getValues/wasChanged/rollback semantics
// over a single source of truth — `__originalValues`, the snapshot of the
// last load or save. Subclasses typically override getDefaults() to declare
// the field shape and type hints used by getValues() coercion.
export class Model {
  // Initializes __originalValues to an empty object and seeds the instance
  // by routing through setValues() — either with the caller-supplied values
  // or, when none are given, with the subclass's getDefaults() so every
  // tracked field starts in a known state.
  constructor(values) {
    this.__originalValues = {};

    if (values) {
      this.setValues(values);
    } else {
      this.setValues(this.getDefaults());
    }
  }

  // Copies every own enumerable key of `values` onto `this` and records a
  // snapshot in __originalValues so wasChanged()/rollback()/getValues(true)
  // can later compare the current state against it. Scalars are copied as-is;
  // objects are deep-cloned with JSON so future mutations on this[key] don't
  // bleed into the snapshot. Pass scalarOnly=true to skip object snapshots
  // (used when a partial update should not reset object diffs). The reserved
  // key "__originalValues" is always ignored. No-op for falsy `values`.
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

  // Returns a plain object containing the model's tracked fields. When
  // `changed` is true, only the fields whose current value differs from
  // __originalValues are included — that's the diff Rest.update() ships to
  // the API. Values are coerced to match the type declared by getDefaults():
  // strings replace null/undefined with "", numeric and boolean fields are
  // forced through parseFloat / Boolean. Fields without a default pass
  // through as-is, so subclasses that hold transient runtime data (functions,
  // helpers) are safe.
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

  // Returns the snapshot stored for `key` in __originalValues — i.e., the
  // value as of the last load or save. If the field was never tracked but
  // exists on the instance, returns the live value as a fallback so callers
  // don't have to special-case ad-hoc fields. Returns null when the key is
  // unknown or when "__originalValues" is requested directly.
  originalValue(key) {
    if (this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
      return this.__originalValues[key];
    } else if (this.hasOwnProperty(key) && key !== "__originalValues") {
      return this[key];
    }

    return null;
  }

  // Returns true when any tracked field has been modified since the last
  // load or save, i.e. when getValues(true) yields a non-empty diff.
  // rollback() is the natural inverse: after rollback() runs, wasChanged()
  // is false until the next mutation.
  wasChanged() {
    const changed = this.getValues(true);

    if (!changed) {
      return false;
    }

    return !(changed.constructor === Object && Object.keys(changed).length === 0);
  }

  // Restores every tracked field to the snapshot stored in __originalValues
  // (the values captured the last time setValues() ran — typically when the
  // model was loaded or saved). Object values are deep-cloned on the way
  // back so future mutations on this[key] don't bleed into the snapshot.
  // Use this to undo an optimistic local mutation when the corresponding
  // API call rejects; wasChanged() is the natural inverse check.
  rollback() {
    for (let key in this.__originalValues) {
      if (this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
        const original = this.__originalValues[key];
        if (typeof original === "object" && original !== null) {
          this[key] = JSON.parse(JSON.stringify(original));
        } else {
          this[key] = original;
        }
      }
    }
    return this;
  }

  // Returns the default field shape for this model, used by the constructor
  // (when no values are passed) and as a type hint by getValues() for
  // string/number/boolean coercion. Subclasses MUST override this and list
  // every persisted field with a representative default of the right type;
  // fields not listed here are still copied by setValues() but skip
  // coercion, so getValues() returns them verbatim.
  getDefaults() {
    return {};
  }
}

export default Model;
