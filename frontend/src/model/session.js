/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

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

import RestModel from "model/rest";
import { $gettext } from "common/vm";
import Util from "common/util";
import * as admin from "options/admin";
import memoizeOne from "memoize-one";

export class Session extends RestModel {
  getDefaults() {
    return {
      ID: "",
      ClientIP: "",
      LoginIP: "",
      LoginAt: "",
      UserUID: "",
      UserName: "",
      UserAgent: "",
      AuthProvider: "",
      AuthMethod: "",
      AuthDomain: "",
      AuthScope: "",
      AuthID: "",
      LastActive: 0,
      Expires: 0,
      Timeout: 0,
      CreatedAt: "",
      UpdatedAt: "",
    };
  }

  getEntityName() {
    return this.getDisplayName();
  }

  authInfo() {
    if (!this || !this.AuthProvider) {
      return $gettext("Default");
    }

    let providerName = memoizeOne(admin.AuthProviders)()[this.AuthProvider];

    if (providerName) {
      providerName = $gettext(providerName);
    } else {
      providerName = Util.capitalize(this.AuthProvider);
    }

    if (!this.AuthMethod || this.AuthMethod === "" || this.AuthMethod === "default") {
      return providerName;
    }

    let methodName = memoizeOne(admin.AuthMethods)()[this.AuthMethod];

    if (!methodName) {
      methodName = this.AuthMethod;
    }

    return `${providerName} (${methodName})`;
  }

  scopeInfo() {
    if (!this || !this.AuthScope) {
      return "*";
    }

    return this.AuthScope;
  }

  static getCollectionResource() {
    return "session";
  }

  static getModelName() {
    return $gettext("Session");
  }
}

export default Session;
