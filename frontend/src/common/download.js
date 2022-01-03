/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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
