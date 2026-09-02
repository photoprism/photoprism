import $api from "common/api";
import { DateTime } from "luxon";
import { $config } from "app/session";
import { $gettext } from "common/gettext";
import Collection from "model/collection";

const SubjPerson = "person";

export let BatchSize = 60;

// MaxLength mirrors the backend setter clip in internal/entity/subject.go (SetName → clean.Name).
export const MaxLength = Object.freeze({
  Name: 160,
});

// BirthYearMin mirrors the constant of the same name in internal/entity/subject.go, which rejects a
// date of birth before it. Bounds the picker so an implausible year cannot be offered at all.
export const BirthYearMin = 1800;

// Subject tracks people and other recognizable subjects derived from face/marker data.
export class Subject extends Collection {
  getDefaults() {
    return {
      UID: "",
      Type: "",
      Src: "",
      Slug: "",
      Name: "",
      Alias: "",
      // Null rather than "", so clearing the field sends a JSON null the API reads as "no date".
      Birthday: null,
      About: "",
      Bio: "",
      Notes: "",
      Favorite: false,
      Verified: false,
      Hidden: false,
      Private: false,
      Excluded: false,
      FileCount: 0,
      PhotoCount: 0,
      Thumb: "",
      ThumbSrc: "",
      CreatedAt: "",
      UpdatedAt: "",
      DeletedAt: "",
    };
  }

  route(view) {
    if (this.Slug && (!this.Type || this.Type === SubjPerson)) {
      return { name: view, query: { q: `person:${this.Slug}` } };
    }

    return { name: view, query: { q: `subject:${this.UID}` } };
  }

  classes(selected) {
    let classes = ["is-subject", "uid-" + this.UID];

    if (this.Favorite) {
      classes.push("is-favorite");
    }
    if (this.Hidden) {
      classes.push("is-hidden");
    }
    if (this.Private) {
      classes.push("is-private");
    }
    if (this.Excluded) {
      classes.push("is-excluded");
    }
    if (selected) {
      classes.push("is-selected");
    }

    return classes;
  }

  getEntityName() {
    return this.Slug;
  }

  // trimInputs strips whitespace from MaxLength string fields before save.
  trimInputs() {
    for (const key of Object.keys(MaxLength)) {
      if (typeof this[key] === "string") {
        this[key] = this[key].trim();
      }
    }
  }

  getTitle() {
    return this.Name;
  }

  // getBirthday returns the date of birth as a local Date for the picker, or null when it is unset.
  // The day is read in UTC and rebuilt locally, so neither the browser's zone nor the offset the API
  // happens to render the instant in can move it - the value is stored as UTC midnight, and reading
  // local parts west of Greenwich or slicing an offset rendering both land on the day before.
  getBirthday() {
    const born = typeof this.Birthday === "string" ? new Date(this.Birthday) : null;

    if (!born || isNaN(born.getTime())) {
      return null;
    }

    return new Date(born.getUTCFullYear(), born.getUTCMonth(), born.getUTCDate());
  }

  // setBirthday stores a local Date as UTC midnight on the same calendar day, so the day the user
  // picked is the day the API records. Anything that is not a date clears the field.
  setBirthday(date) {
    if (!(date instanceof Date) || isNaN(date.getTime())) {
      this.Birthday = null;
      return;
    }

    const pad = (n, len) => String(n).padStart(len, "0");

    this.Birthday = `${pad(date.getFullYear(), 4)}-${pad(date.getMonth() + 1, 2)}-${pad(date.getDate(), 2)}T00:00:00Z`;
  }

  thumbnailUrl(size) {
    if (!this.Thumb) {
      return `${$config.contentUri}/svg/portrait`;
    }

    if (!size) {
      size = "tile_160";
    }

    return `${$config.contentUri}/t/${this.Thumb}/${$config.previewToken}/${size}`;
  }

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
  }

  show() {
    this.Hidden = false;
    return this.update();
  }

  hide() {
    this.Hidden = true;
    return this.update();
  }

  toggleHidden() {
    this.Hidden = !this.Hidden;

    return $api.put(this.getEntityResource(), { Hidden: this.Hidden });
  }

  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return $api.post(this.getEntityResource() + "/like");
    } else {
      return $api.delete(this.getEntityResource() + "/like");
    }
  }

  like() {
    this.Favorite = true;
    return $api.post(this.getEntityResource() + "/like");
  }

  unlike() {
    this.Favorite = false;
    return $api.delete(this.getEntityResource() + "/like");
  }

  static batchSize() {
    return BatchSize;
  }

  static setBatchSize(count) {
    const s = parseInt(count);
    if (!isNaN(s) && s >= 24) {
      BatchSize = s;
    }
  }

  static getCollectionResource() {
    return "subjects";
  }

  static getModelName() {
    return $gettext("Person");
  }
}

export default Subject;
