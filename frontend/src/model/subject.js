import $api from "common/api";
import { DateTime } from "luxon";
import { $config } from "app/session";
import { $gettext } from "common/gettext";
import Collection from "model/collection";

const SubjPerson = "person";

export let BatchSize = 60;

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
      About: "",
      Bio: "",
      Notes: "",
      Favorite: false,
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

    if (this.Favorite) classes.push("is-favorite");
    if (this.Hidden) classes.push("is-hidden");
    if (this.Private) classes.push("is-private");
    if (this.Excluded) classes.push("is-excluded");
    if (selected) classes.push("is-selected");

    return classes;
  }

  getEntityName() {
    return this.Slug;
  }

  getTitle() {
    return this.Name;
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
