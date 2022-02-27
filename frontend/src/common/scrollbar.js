/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

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

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

let hidePending = 0;
let hideDefault = document.body.classList.contains("hide-scrollbar");

const Scrollbar = {
  html: function () {
    return document.getElementsByTagName("html")[0];
  },
  body: function () {
    return document.body;
  },
  update: function () {
    const htmlEl = this.html();
    const bodyEl = this.body();

    if (!htmlEl || !bodyEl) {
      return;
    }

    if (this.hidden()) {
      htmlEl.setAttribute("class", "overflow-y-hidden");
      bodyEl.classList.add("hide-scrollbar");
    } else {
      htmlEl.removeAttribute("class");
      bodyEl.classList.remove("hide-scrollbar");
    }
  },
  show: function () {
    if (hidePending > 0) {
      hidePending--;
    }

    this.update();
  },
  hide: function () {
    hidePending++;

    this.update();
  },
  hidden: function () {
    return hidePending > 0 || hideDefault;
  },
};

Scrollbar.update();

export default Scrollbar;
