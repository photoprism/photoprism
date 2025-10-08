import Model from "model.js";
import { DateTime } from "luxon";
import { $gettext } from "common/gettext";

const c = window.__CONFIG__;

// Link encapsulates share links for albums/photos exposed via the public API.
export default class Link extends Model {
  getDefaults() {
    return {
      UID: "",
      ShareUID: "",
      Slug: "",
      Token: "",
      Expires: 0,
      Views: 0,
      MaxViews: 0,
      Password: "",
      HasPassword: false,
      Comment: "",
      Perm: 0,
      CreatedAt: "",
      ModifiedAt: "",
    };
  }

  getToken() {
    return this.Token.toLowerCase().trim();
  }

  siteUrl() {
    let siteUrl = c.siteUrl ? c.siteUrl : window.location.origin;

    if (siteUrl.slice(-1) !== "/") {
      siteUrl = siteUrl + "/";
    }

    return siteUrl;
  }

  url() {
    const siteUrl = this.siteUrl();
    let token = this.getToken();

    if (!token) {
      token = "…";
    }

    if (this.hasSlug()) {
      return `${siteUrl}s/${token}/${this.Slug}`;
    }

    return `${siteUrl}s/${token}/${this.ShareUID}`;
  }

  caption() {
    return `/s/${this.getToken()}`;
  }

  getId() {
    if (this.UID) {
      return this.UID;
    }

    return this.ID ? this.ID : false;
  }

  hasId() {
    return !!this.getId();
  }

  getSlug() {
    return this.Slug ? this.Slug : "";
  }

  hasSlug() {
    return !!this.getSlug();
  }

  clone() {
    return new this.constructor(this.getValues());
  }

  expires() {
    return DateTime.fromISO(this.ModifiedAt).plus({ seconds: this.Expires }).toLocaleString(DateTime.DATE_MED);
  }

  static getCollectionResource() {
    return "links";
  }

  static getModelName() {
    return $gettext("Link");
  }
}
