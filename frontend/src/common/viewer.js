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

import PhotoSwipe from "photoswipe";
import PhotoSwipeUI_Default from "photoswipe/dist/photoswipe-ui-default.js";
import Event from "pubsub-js";
import Util from "util.js";
import Api from "./api";
import Thumb from "model/thumb";

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

        item.title = item.Title;

        if (!item.Title) {
          captionEl.children[0].innerHTML = "";
          return false;
        }

        captionEl.children[0].innerHTML = Util.encodeHTML(item.Title);

        if (item.Playable) {
          captionEl.children[0].innerHTML += ' <i aria-hidden="true" class="v-icon material-icons theme--dark" title="Play">play_circle_fill</i>';
        }

        if (item.Description) {
          captionEl.children[0].innerHTML += '<br><span class="description">' + Util.encodeHTML(item.Description) + "</span>";
        }

        if (item.Playable) {
          captionEl.children[0].innerHTML = "<button>" + captionEl.children[0].innerHTML + "</button>";
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

    gallery.listen("beforeChange", () => Event.publish("viewer.change", { gallery: gallery, item: gallery.currItem }));

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
      item.src = item.Thumbs[nextSize].src;
      item.w = item.Thumbs[nextSize].w;
      item.h = item.Thumbs[nextSize].h;
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

  static show(ctx, index) {
    if (ctx.loading || !ctx.listen || ctx.viewer.loading || !ctx.results[index]) {
      return false;
    }

    const selected = ctx.results[index];

    if (!ctx.viewer.dirty && ctx.viewer.results && ctx.viewer.results.length > index) {
      // Reuse existing viewer result if possible.
      let i = -1;

      if (ctx.viewer.results[index] && ctx.viewer.results[index].UID === selected.UID) {
        i = index;
      } else {
        i = ctx.viewer.results.findIndex((p) => p.UID === selected.UID);
      }

      if (i > -1 && (((ctx.viewer.complete || ctx.complete) && ctx.viewer.results.length >= ctx.results.length) || i + ctx.viewer.batchSize <= ctx.viewer.results.length)) {
        ctx.$viewer.show(ctx.viewer.results, i);
        return;
      }
    }

    // Fetch photos from server API.
    ctx.viewer.loading = true;

    const params = ctx.searchParams();
    params.count = params.offset + ctx.viewer.batchSize;
    params.offset = 0;

    // Fetch viewer results from API.
    return Api.get("photos/view", { params })
      .then((response) => {
        const count = response && response.data ? response.data.length : 0;
        if (count === 0) {
          ctx.$notify.warn(ctx.$gettext("No pictures found"));
          ctx.viewer.dirty = true;
          ctx.viewer.complete = false;
          return;
        }

        // Process response.
        if (response.headers && response.headers["x-count"]) {
          const c = parseInt(response.headers["x-count"]);
          const l = parseInt(response.headers["x-limit"]);
          ctx.viewer.complete = c < l;
        } else {
          ctx.viewer.complete = ctx.complete;
        }

        let i;

        if (response.data[index] && response.data[index].UID === selected.UID) {
          i = index;
        } else {
          i = response.data.findIndex((p) => p.UID === selected.UID);
        }

        ctx.viewer.results = Thumb.wrap(response.data);

        // Show photos.
        ctx.$viewer.show(ctx.viewer.results, i);
        ctx.viewer.dirty = false;
      })
      .catch(() => {
        ctx.viewer.dirty = true;
        ctx.viewer.complete = false;
      })
      .finally(() => {
        // Unblock.
        ctx.viewer.loading = false;
      });
  }
}

export default Viewer;
