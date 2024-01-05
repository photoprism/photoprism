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

import saveAs from "file-saver";

// Detect Safari browser.
const isSafari =
  navigator.appVersion.indexOf("Safari/") !== -1 && navigator.appVersion.indexOf("Chrome") === -1;

// Downloads a file from the server.
export default function download(url, name) {
  // Abort if download url is empty.
  if (!url) {
    console.warn("can't download: empty url");
    return;
  }

  if (isSafari) {
    const xhr = new XMLHttpRequest();
    xhr.open("GET", url);
    xhr.responseType = "blob";

    xhr.onload = function () {
      saveAs(xhr.response, name);
    };

    xhr.onerror = function () {
      console.error("download failed", url);
    };

    xhr.send();
    return;
  }

  // Create download link.
  const link = document.createElement("a");

  if (name) {
    link.download = name;
  }

  link.href = url;
  link.style.display = "none";

  document.body.appendChild(link);

  // Start download.
  link.click();

  // Remove download link.
  document.body.removeChild(link);
}
