/*

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

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

import Axios from "axios";
import Notify from "common/notify";
import { $gettext } from "vm.js";
import Event from "pubsub-js";

const testConfig = {
  baseUri: "",
  staticUri: "/static",
  apiUri: "/api/v1",
  contentUri: "/api/v1",
  debug: false,
  previewToken: "public",
  downloadToken: "public",
  cssUri: "/static/build/app.2259c0edcc020e7af593.css",
  jsUri: "/static/build/app.9bd7132eaee8e4c7c7e3.js",
  manifestUri: "/manifest.json",
};

const c = window.__CONFIG__ ? window.__CONFIG__ : testConfig;

const Api = Axios.create({
  baseURL: c.apiUri,
  headers: {
    common: {
      "X-Session-ID": window.localStorage.getItem("session_id"),
      "X-Client-Uri": c.jsUri,
      "X-Client-Version": c.version,
    },
  },
});

Api.interceptors.request.use(
  function (req) {
    // Do something before request is sent
    Notify.ajaxStart();
    return req;
  },
  function (error) {
    // Do something with request error
    return Promise.reject(error);
  }
);

Api.interceptors.response.use(
  function (resp) {
    Notify.ajaxEnd();

    if (typeof resp.data == "string") {
      Notify.error($gettext("Request failed - invalid response"));
      console.warn("WARNING: Server returned HTML instead of JSON - API not implemented?");
    }

    // Update tokens if provided.
    if (resp.headers && resp.headers["x-preview-token"] && resp.headers["x-download-token"]) {
      Event.publish("config.tokens", {
        previewToken: resp.headers["x-preview-token"],
        downloadToken: resp.headers["x-download-token"],
      });
    }

    return resp;
  },
  function (error) {
    Notify.ajaxEnd();

    // Skip error handling if request was canceled.
    if (Axios.isCancel(error)) {
      return Promise.reject(error);
    }

    // Log error for debugging.
    if (console && console.log && error) {
      console.log(error);
    }

    // Default error message.
    let errorMessage = $gettext("Something went wrong, try again");
    let code = error.code;

    // Extract error details from response.
    if (error.response && error.response.data) {
      let data = error.response.data;
      code = data.code;
      errorMessage = data.message ? data.message : data.error;
    }

    // Show error notification.
    if (errorMessage) {
      if (code === 401) {
        Notify.logout(errorMessage);
      } else {
        Notify.error(errorMessage);
      }
    }

    return Promise.reject(error);
  }
);

export default Api;
