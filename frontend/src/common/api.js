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

import Axios from "axios";
import $notify from "common/notify";
import { $gettext } from "common/gettext";
import $event from "common/event";

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

const $api = Axios.create({
  baseURL: c.apiUri,
  headers: {
    common: {
      "X-Auth-Token": window.localStorage.getItem("authToken"),
      "X-Client-Uri": c.jsUri,
      "X-Client-Version": c.version,
    },
  },
});

$api.interceptors.request.use(
  function (req) {
    // Do something before request is sent
    $notify.ajaxStart();
    return req;
  },
  function (error) {
    // Do something with request error
    return Promise.reject(error);
  }
);

$api.interceptors.response.use(
  function (resp) {
    $notify.ajaxEnd();

    if (typeof resp.data == "string") {
      $notify.error($gettext("Request failed - invalid response"));
      console.warn("WARNING: Server returned HTML instead of JSON - API not implemented?");
    }

    // Update tokens if provided.
    if (resp.headers && resp.headers["x-preview-token"] && resp.headers["x-download-token"]) {
      $event.publish("config.tokens", {
        previewToken: resp.headers["x-preview-token"],
        downloadToken: resp.headers["x-download-token"],
      });
    }

    return resp;
  },
  function (error) {
    $notify.ajaxEnd();

    // Skip error handling if request was canceled.
    if (Axios.isCancel(error)) {
      return Promise.reject(error);
    }

    // Log error for debugging.
    if (console && console.log && error) {
      console.log(error);
    }

    // Default error message.
    let errorMessage = $gettext("Request failed - are you offline?");
    let code = error.response && error.response.status ? error.response.status : 0;

    // Extract error details from response.
    if (error.response && typeof error.response.data === "object") {
      let data = error.response.data;

      if (data.code) {
        code = data.code;
      }

      if (data.message) {
        errorMessage = data.message;
      } else if (data.error) {
        errorMessage = data.error;
      }
    }

    // Show error notification.
    if (code === 32) {
      $notify.info($gettext("Enter verification code"));
    } else if (code === 429) {
      $notify.error($gettext("Too many requests"));
    } else if (errorMessage) {
      if (code === 401) {
        $notify.logout(errorMessage);
      } else {
        $notify.error(errorMessage);
      }
    }

    return Promise.reject(error);
  }
);

export default $api;
