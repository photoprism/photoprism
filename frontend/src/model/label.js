import Collection from "model/collection";
import $api from "common/api";
import { DateTime } from "luxon";
import { $config } from "app/session";
import { $gettext } from "common/gettext";

export let BatchSize = 180;

// Label models user-defined keywords and AI-generated tags.
export class Label extends Collection {
  getDefaults() {
    return {
      ID: 0,
      UID: "",
      Slug: "",
      CustomSlug: "",
      Name: "",
      Priority: 0,
      Favorite: false,
      Description: "",
      Notes: "",
      PhotoCount: 0,
      CreatedAt: "",
      UpdatedAt: "",
      DeletedAt: "",
    };
  }

  route(view) {
    return { name: view, query: { q: "label:" + (this.CustomSlug ? this.CustomSlug : this.Slug) } };
  }

  classes(selected) {
    let classes = ["is-label", "uid-" + this.UID];

    if (this.Favorite) classes.push("is-favorite");
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
    if (this.Thumb) {
      return `${$config.contentUri}/t/${this.Thumb}/${$config.previewToken}/${size}`;
    } else if (this.UID) {
      return `${$config.contentUri}/labels/${this.UID}/t/${$config.previewToken}/${size}`;
    } else {
      return `${$config.contentUri}/svg/label`;
    }
  }

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
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
    return "labels";
  }

  static getModelName() {
    return $gettext("Label");
  }
}

export default Label;
