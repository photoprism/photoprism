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
    <https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import $event from "common/event";
import { $gettext } from "common/gettext";

let ajaxPending = 0;
let ajaxCallbacks = [];

const $notify = {
  // info publishes an info notification.
  info: function (message, messageId, messageParams) {
    $event.publish("notify.info", { message, messageId, messageParams });
  },
  // warn publishes a warning notification.
  warn: function (message, messageId, messageParams) {
    $event.publish("notify.warning", { message, messageId, messageParams });
  },
  // error publishes an error notification; the optional messageId/messageParams
  // let notify.vue render it in the current UI locale via Tp.
  error: function (message, messageId, messageParams) {
    $event.publish("notify.error", { message, messageId, messageParams });
  },
  // success publishes a success notification.
  success: function (message, messageId, messageParams) {
    $event.publish("notify.success", { message, messageId, messageParams });
  },
  // logout publishes an error notification and triggers a session.logout event.
  logout: function (message, messageId, messageParams) {
    $event.publish("notify.error", { message, messageId, messageParams });
    $event.publish("session.logout", { message, messageId, messageParams });
  },
  // ajaxStart marks an AJAX request as started and emits ajax.start.
  ajaxStart: function () {
    ajaxPending++;
    $event.publish("ajax.start");
  },
  // ajaxEnd marks an AJAX request as finished, emits ajax.end, and runs queued idle callbacks.
  ajaxEnd: function () {
    ajaxPending--;
    $event.publish("ajax.end");

    if (!this.ajaxBusy() && ajaxCallbacks.length) {
      const callbacks = ajaxCallbacks;
      ajaxCallbacks = [];
      callbacks.forEach((cb) => cb());
    }
  },
  // ajaxBusy reports whether any AJAX requests are still pending.
  ajaxBusy: function () {
    if (ajaxPending < 0) {
      ajaxPending = 0;
    }

    return ajaxPending > 0;
  },
  // ajaxWait resolves once no AJAX requests are pending or the timeout elapses.
  ajaxWait: function (idleDelay = 64, timeout = 8000) {
    return new Promise((resolve) => {
      const start = Date.now();

      const settle = () => {
        if (timeout && Date.now() - start > timeout) {
          resolve();
          return;
        }

        if (this.ajaxBusy()) {
          ajaxCallbacks.push(settle);
          return;
        }

        window.setTimeout(() => {
          if (this.ajaxBusy()) {
            settle();
          } else {
            resolve();
          }
        }, idleDelay);
      };

      settle();
    });
  },
  // blockUI shows the busy overlay, optionally setting its CSS class.
  blockUI: function (className) {
    const el = document.getElementById("busy-overlay");

    if (el) {
      el.style.display = "block";
      if (className) {
        el.className = className;
      }
    }
  },
  // unblockUI hides the busy overlay.
  unblockUI: function () {
    const el = document.getElementById("busy-overlay");

    if (el) {
      el.style.display = "none";
      el.className = "";
    }
  },
  // wait shows a "Please wait" info notification.
  wait: function () {
    this.info($gettext("Please wait…"));
  },
  // busy shows a "Busy, please wait" warning notification.
  busy: function () {
    this.warn($gettext("Busy, please wait…"));
  },
};

export default $notify;
