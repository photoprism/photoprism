/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

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

// withResolvers defines Promise.withResolvers where the browser lacks it (Safari 16.4 to 17.3),
// because pdf.js calls it on the main thread and in its worker without a polyfill. Remove this
// once the supported browsers all provide it natively (Safari 17.4).
export function withResolvers() {
  if (typeof Promise.withResolvers === "function") {
    return false;
  }

  Object.defineProperty(Promise, "withResolvers", {
    configurable: true,
    writable: true,
    value: function () {
      let resolve;
      let reject;
      const promise = new this((res, rej) => {
        resolve = res;
        reject = rej;
      });
      return { promise, resolve, reject };
    },
  });

  return true;
}

withResolvers();
