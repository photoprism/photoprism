/*

Copyright (c) 2018 - 2022 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import Model from "model.js";
import { DateTime } from "luxon";
import { $gettext } from "common/vm";

const c = window.__CONFIG__;

export default class Link extends Model {
  getDefaults() {
    return {
      UID: "",
      Share: "",
      Slug: "",
      Token: "",
      Expires: 0,
      Views: 0,
      MaxViews: 0,
      Password: "",
      HasPassword: false,
      CanComment: false,
      CanEdit: false,
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

    return `${siteUrl}s/${token}/${this.Share}`;
  }

  caption() {
    return `/s/${this.getToken()}`;
  }

  getId() {
    return this.UID;
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
    return DateTime.fromISO(this.ModifiedAt)
      .plus({ seconds: this.Expires })
      .toLocaleString(DateTime.DATE_MED);
  }

  static getCollectionResource() {
    return "links";
  }

  static getModelName() {
    return $gettext("Link");
  }
}
