import RestModel from "model/rest";
import * as auth from "options/auth";
import memoizeOne from "memoize-one";
import $util from "common/util";
import { $gettext, T } from "common/gettext";

export let BatchSize = 99999;

// Session represents authenticated sessions (user, public, or client tokens).
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
      ClientUID: "",
      ClientName: "",
      AuthProvider: "",
      AuthMethod: "",
      AuthIssuer: "",
      AuthID: "",
      AuthScope: "",
      GrantType: "",
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

    let providerName = memoizeOne(auth.Providers)()[this.AuthProvider];

    if (providerName) {
      providerName = T(providerName);
    } else {
      providerName = $util.capitalize(this.AuthProvider);
    }

    if (!this.AuthMethod || this.AuthMethod === "" || this.AuthMethod === "default") {
      return providerName;
    }

    let methodName = memoizeOne(auth.Methods)()[this.AuthMethod];

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

  static batchSize() {
    return BatchSize;
  }

  static getCollectionResource() {
    return "sessions";
  }

  static getModelName() {
    return $gettext("Session");
  }
}

export default Session;
