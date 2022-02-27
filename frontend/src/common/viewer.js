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

import PhotoSwipe from "photoswipe";
import PhotoSwipeUI_Default from "photoswipe/dist/photoswipe-ui-default.js";
import Event from "pubsub-js";
import Util from "util.js";

const thumbs = window.__CONFIG__.thumbs;

class Viewer {
  constructor() {
    this.el = null;
    this.gallery = null;
  }

  getEl() {
    if (!this.el) {
      this.el = document.getElementById("photo-viewer");

      if (this.el === null) {
        let err = "no photo viewer element found";
        console.warn(err);
        throw err;
      }
    }

    return this.el;
  }

  play(params) {
    Event.publish("player.open", params);
  }

  show(items, index = 0) {
    if (!Array.isArray(items) || items.length === 0 || index >= items.length) {
      console.log("photo list passed to gallery was empty:", items);
      return;
    }

    const shareButtons = [
      {
        id: "fit_720",
        template: "Tiny (size)",
        label: "Tiny",
        url: "{{raw_image_url}}",
        download: true,
      },
      {
        id: "fit_1280",
        template: "Small (size)",
        label: "Small",
        url: "{{raw_image_url}}",
        download: true,
      },
      {
        id: "fit_2048",
        template: "Medium (size)",
        label: "Medium",
        url: "{{raw_image_url}}",
        download: true,
      },
      {
        id: "fit_2560",
        template: "Large (size)",
        label: "Large",
        url: "{{raw_image_url}}",
        download: true,
      },
      {
        id: "original",
        template: "Original (size)",
        label: "Original",
        url: "{{raw_image_url}}",
        download: true,
      },
    ];

    const options = {
      index: index,
      history: true,
      preload: [1, 1],
      focus: true,
      modal: true,
      closeEl: true,
      captionEl: true,
      fullscreenEl: true,
      zoomEl: true,
      shareEl: true,
      shareButtons: shareButtons,
      counterEl: false,
      arrowEl: true,
      preloaderEl: true,
      addCaptionHTMLFn: function (item, captionEl /*, isFake */) {
        // item      - slide object
        // captionEl - caption DOM element
        // isFake    - true when content is added to fake caption container
        //             (used to get size of next or previous caption)

        if (!item.title) {
          captionEl.children[0].innerHTML = "";
          return false;
        }

        captionEl.children[0].innerHTML = Util.encodeHTML(item.title);

        if (item.playable) {
          captionEl.children[0].innerHTML +=
            ' <i aria-hidden="true" class="v-icon material-icons theme--dark" title="Play">play_circle_fill</i>';
        }

        if (item.description) {
          captionEl.children[0].innerHTML +=
            '<br><span class="description">' + Util.encodeHTML(item.description) + "</span>";
        }

        if (item.playable) {
          captionEl.children[0].innerHTML =
            "<button>" + captionEl.children[0].innerHTML + "</button>";
        }

        return true;
      },
    };

    let gallery = new PhotoSwipe(this.getEl(), PhotoSwipeUI_Default, items, options);
    let realViewportWidth;
    let realViewportHeight;
    let previousSize;
    let nextSize;
    let firstResize = true;
    let photoSrcWillChange;

    this.gallery = gallery;

    Event.publish("viewer.show");

    gallery.listen("close", () => {
      Event.publish("viewer.pause");
      Event.publish("viewer.hide");
    });
    gallery.listen("shareLinkClick", () => Event.publish("viewer.pause"));
    gallery.listen("initialZoomIn", () => Event.publish("viewer.pause"));
    gallery.listen("initialZoomOut", () => Event.publish("viewer.pause"));

    gallery.listen("beforeChange", () =>
      Event.publish("viewer.change", { gallery: gallery, item: gallery.currItem })
    );

    gallery.listen("beforeResize", () => {
      realViewportWidth = gallery.viewportSize.x * window.devicePixelRatio;
      realViewportHeight = gallery.viewportSize.y * window.devicePixelRatio;

      if (!previousSize) {
        previousSize = "tile_720";
      }

      nextSize = this.constructor.mapViewportToImageSize(realViewportWidth, realViewportHeight);

      if (nextSize !== previousSize) {
        photoSrcWillChange = true;
      }

      if (photoSrcWillChange && !firstResize) {
        gallery.invalidateCurrItems();
      }

      if (firstResize) {
        firstResize = false;
      }

      photoSrcWillChange = false;
    });

    gallery.listen("gettingData", function (index, item) {
      item.src = item[nextSize].src;
      item.w = item[nextSize].w;
      item.h = item[nextSize].h;
      previousSize = nextSize;
    });

    gallery.init();
  }

  static mapViewportToImageSize(viewportWidth, viewportHeight) {
    for (let i = 0; i < thumbs.length; i++) {
      let t = thumbs[i];

      if (t.w >= viewportWidth || t.h >= viewportHeight) {
        return t.size;
      }
    }

    return "fit_7680";
  }
}

export default Viewer;
