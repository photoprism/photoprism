<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :scrollable="false"
    :transition="false"
    :close-delay="0"
    :open-delay="0"
    fullscreen
    scrim
    persistent
    tiled
    theme="lightbox"
    class="p-dialog p-lightbox v-dialog--lightbox no-transition"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
    @keydown.space.exact="onKeyDown"
    @keydown.left.exact="onKeyDown"
    @keydown.right.exact="onKeyDown"
    @keydown.esc.stop="close"
    @click.capture="captureDialogClick"
    @pointerdown.capture="captureDialogPointerDown"
  >
    <div class="p-lightbox__underlay no-transition"></div>
    <div ref="container" class="p-lightbox__container no-transition">
      <div
        ref="content"
        tabindex="-1"
        class="p-lightbox__content no-transition"
        :class="{
          'sidebar-visible': info,
          'slideshow-active': slideshow.active,
          'is-fullscreen': isFullscreen(),
          'is-zoomable': isZoomable,
          'is-favorite': model.Favorite,
          'is-playable': model.Playable,
          'is-video': model?.Type === 'video',
          'is-muted': muted,
          'is-selected': $clipboard.has(model),
        }"
      >
        <div ref="lightbox" tabindex="-1" class="p-lightbox__pswp no-transition"></div>
        <div
          v-show="video.controls && controlsShown !== 0"
          ref="controls"
          tabindex="-1"
          class="p-lightbox__controls"
          @click.stop.prevent
        >
          <div :title="video.error" class="video-control video-control--play">
            <v-icon v-if="video.error || video.errorCode > 0" icon="mdi-alert"></v-icon>
            <v-icon v-else-if="video.seeking || video.waiting" icon="mdi-loading" class="animate-loading"></v-icon>
            <v-icon
              v-else-if="video.playing"
              icon="mdi-pause"
              class="clickable"
              @pointerdown.stop.prevent="toggleVideo"
            ></v-icon>
            <v-icon v-else icon="mdi-play" class="clickable" @pointerdown.stop.prevent="toggleVideo"></v-icon>
          </div>
          <div class="video-control video-control--time text-body-2">
            {{ $util.formatSeconds(video.ended ? Math.ceil(video.time) : Math.floor(video.time)) }}
          </div>
          <v-slider
            :model-value="video.time"
            :disabled="!video.seekable"
            :readonly="video.seeking"
            :thumb-size="12"
            :track-size="3"
            hide-details
            :error="video.errorCode > 0"
            :min="0"
            :max="video.duration"
            class="video-control video-control--slider"
            @update:model-value="seekVideo"
            @update:focused="focusContent"
          >
          </v-slider>
          <div class="video-control video-control--duration text-body-2">
            {{ $util.formatRemainingSeconds(video.time, video.duration) }}
          </div>
          <div v-if="featExperimental && video.castable" class="video-control video-control--cast">
            <v-icon
              v-if="video.casting"
              icon="mdi-cast-connected"
              class="clickable"
              @pointerdown.stop.prevent="toggleVideoRemote"
            ></v-icon>
            <v-icon
              v-else
              icon="mdi-cast"
              :disabled="video.remote === 'connecting'"
              class="clickable"
              @pointerdown.stop.prevent="toggleVideoRemote"
            ></v-icon>
          </div>
        </div>
      </div>
      <div v-if="info" ref="sidebar" tabindex="-1" class="p-lightbox__sidebar bg-background">
        <p-sidebar-info v-model="model" :collection="collection" :context="context" @close="hideInfo"></p-sidebar-info>
      </div>
    </div>
    <p-lightbox-menu
      ref="menu"
      :items="menuActions"
      :activator="menuElement"
      attach=".v-dialog--lightbox.v-overlay--active"
      @show="onShowMenu"
      @hide="onHideMenu"
    ></p-lightbox-menu>
  </v-dialog>
</template>

<script>
import PhotoSwipe from "photoswipe";
import Lightbox from "photoswipe/lightbox";
import Captions from "common/captions";
import $api from "common/api";
import $fullscreen from "common/fullscreen";
import Thumb from "model/thumb";
import Collection from "model/collection";
import { Photo } from "model/photo";
import { Album } from "model/album";
import * as media from "common/media";

const VIDEO_EVENT_TYPES = [
  "loadstart",
  "loadedmetadata",
  "loadeddata",
  "progress",
  "stalled",
  "abort",
  "error",
  "play",
  "playing",
  "pause",
  "waiting",
  "ended",
  "seeked",
  "seeking",
  "timeupdate",
  "durationchange",
];

const VIDEO_REMOTE_EVENT_TYPES = ["connect", "connecting", "disconnect"];

import PLightboxMenu from "component/lightbox/menu.vue";
import PSidebarInfo from "component/sidebar/info.vue";

export default {
  name: "PLightbox",
  components: [PLightboxMenu, PSidebarInfo],
  emits: ["enter", "leave"],
  expose: ["onShortCut"],
  data() {
    const debug = this.$config.get("debug");
    const trace = this.$config.get("trace");
    const features = this.$config.getSettings().features;
    return {
      debug,
      trace,
      visible: false,
      busy: false,
      closing: false,
      info: localStorage.getItem("lightbox.info") === "true",
      menuElement: null,
      menuBgColor: "#252525",
      menuVisible: false,
      lightbox: null, // Current PhotoSwipe lightbox instance.
      captionPlugin: null, // Current PhotoSwipe caption plugin instance.
      muted: window.sessionStorage.getItem("lightbox.muted") === "true",
      hasTouch: this.$util.hasTouch(),
      shortVideoDuration: 5, // Duration in seconds for videos that are short enough to automatically loop.
      playControlHideDelay: 1000, // Hide the lightbox controls after this time in ms when a video starts playing.
      defaultControlHideDelay: 5000, // Automatically hide lightbox controls this time in ms, TODO: add custom settings.
      idleTimer: false,
      controlsShown: -1, // -1 or a positive timestamp indicates that the controls are shown (0 means hidden).
      canEdit: this.$config.allow("photos", "update") && features.edit,
      canLike: this.$config.allow("photos", "manage") && features.favorites,
      canDownload: this.$config.allow("photos", "download") && features.download,
      canArchive: this.$config.allow("photos", "delete") && features.archive,
      canManageAlbums: this.$config.allow("albums", "manage"),
      canFullscreen: $fullscreen.isSupported() && (!this.$isMobile || this.$config.featExperimental()), // see https://developer.mozilla.org/en-US/docs/Web/API/Document/fullscreenEnabled
      wasFullscreen: $fullscreen.isEnabled(),
      isZoomable: true,
      mobileBreakpoint: 600, // Minimum viewport width for large screens.
      featExperimental: this.$config.featExperimental(), // Enables features that may be incomplete or unstable.
      featDevelop: this.$config.featDevelop(), // Enables new features that are still under development.
      selection: this.$clipboard.selection,
      config: this.$config.values,
      collection: null,
      context: "",
      model: new Thumb(), // Current slide.
      models: [], // Slide models.
      index: 0, // Current slide index in models.
      subscriptions: [], // Event subscriptions.
      // Video properties for rendering the controls.
      video: {
        controls: false,
        src: "",
        error: "",
        errorCode: 0,
        state: 0,
        time: 0,
        duration: 0,
        seeking: false,
        seekable: false,
        waiting: false,
        playing: false,
        paused: false,
        ended: false,
        castable: false,
        casting: false,
        remote: "",
      },
      // Slideshow properties.
      slideshow: {
        active: false,
        interval: false,
        wait: 5000,
        waitAfterVideo: 2500,
        next: -1,
      },
      touchStartListener: (ev) => this.onTouchStartOnce(ev),
      mouseMoveListener: (ev) => this.onMouseMoveOnce(ev),
      lightboxPointerListener: (ev) => this.onLightboxPointerEvent(ev),
      videoEventListener: (ev) => this.onVideoEvent(ev),
      videoRemoteListener: (ev) => this.onVideoRemote(ev),
      videoAvailabilityListener: (castable) => {
        if (typeof this.video === "object") {
          this.video.castable = castable;
        }
      },
    };
  },
  created() {
    this.subscriptions.push(this.$event.subscribe("lightbox.open", this.openLightbox.bind(this)));
    this.subscriptions.push(this.$event.subscribe("lightbox.pause", this.pauseLightbox.bind(this)));
    this.subscriptions.push(this.$event.subscribe("lightbox.close", this.onClose.bind(this)));
  },
  beforeUnmount() {
    // Exit fullscreen mode if enabled, has no effect otherwise.
    $fullscreen.exit();

    // Remove timeouts.
    this.clearTimeouts();
    this.removeEventListeners();

    // Pause slideshow and videos.
    this.pauseLightbox();

    // Destroy PhotoSwipe.
    this.destroyLightbox();

    // Remove event listeners.
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    // Opens and initializes the lightbox with the given options.
    openLightbox(ev, data) {
      if (!data) {
        return;
      }

      if (data.view) {
        this.showView(data.view, data.index);
      } else {
        this.showThumbs(data.models, data.index, data);
      }
    },
    // Pauses the lightbox slideshow and any videos that are playing.
    pauseLightbox() {
      this.pausePlaying();
      this.pauseSlideshow();
    },
    // Returns true if a blocking event is currently being processed.
    isBusy(action) {
      if (this.busy) {
        this.log(`still busy, cannot ${action ? action : "do this"}`);
        return true;
      }

      return false;
    },
    // Triggered before the lightbox content is initialized.
    showDialog() {
      this.$view.enter(this, this.$refs?.content);
      this.busy = true;
      this.closing = false;
      this.visible = true;
      this.wasFullscreen = $fullscreen.isEnabled();
      this.info = localStorage.getItem("lightbox.info") === "true";

      // Publish init event.
      this.$event.publish("lightbox.init");
    },
    // Hides the lightbox and restores the scrollbar state.
    hideDialog() {
      // Reset component state.
      this.onReset();

      // Hide sidebar.
      this.info = false;

      // Remove lightbox focus and hide lightbox.
      if (this.visible) {
        this.visible = false;
      }

      this.busy = false;
      this.closing = false;

      // Publish event to be consumed by other components.
      this.$event.publish("lightbox.closed");
    },
    // Triggered when the dialog has been fully opened.
    afterEnter() {
      this.$event.publish("lightbox.enter");
      this.$emit("enter");
    },
    // Triggered when the dialog has closed.
    afterLeave() {
      // Publish enter event.
      this.visible = false;
      this.busy = false;
      this.closing = false;
      this.$view.leave(this);
      this.$event.publish("lightbox.leave");
      this.$emit("leave");
    },
    focusContent(ev) {
      if (
        this.$refs.content &&
        this.$refs.content instanceof HTMLElement &&
        document.activeElement !== this.$refs.content
      ) {
        this.$refs.content.focus();

        if (this.debug && ev) {
          this.log(`set focus to content`, { ev });
        }
      }
    },
    log(ev, data) {
      if (!ev) {
        return;
      }
      if (data) {
        console.log(`%clightbox: ${ev}`, "color: #039be5;", data);
      } else {
        console.log(`%clightbox: ${ev}`, "color: #039be5;");
      }
    },
    // Returns the PhotoSwipe content element.
    getLightboxElement() {
      if (!this.$refs.lightbox) {
        this.log("lightbox element is not visible");
        return null;
      }

      return this.$refs.lightbox;
    },
    // Returns the metadata sidebar element.
    getSidebarElement() {
      if (!this.$refs.sidebar) {
        this.log("sidebar element is not visible");
        return null;
      }

      return this.$refs.sidebar;
    },
    // Returns the PhotoSwipe config options, see https://photoswipe.com/options/.
    getOptions() {
      return {
        appendToEl: this.getLightboxElement(),
        pswpModule: PhotoSwipe,
        index: this.index,
        mouseMovePan: true,
        arrowPrev: true,
        arrowNext: true,
        loop: false,
        zoom: true,
        close: false,
        escKey: false,
        arrowKeys: false,
        pinchToClose: false,
        counter: false,
        trapFocus: false,
        returnFocus: false,
        allowPanToNext: false,
        closeOnVerticalDrag: false,
        initialZoomLevel: "fit",
        secondaryZoomLevel: "fill",
        showHideAnimationType: "none",
        hideAnimationDuration: 0,
        showAnimationDuration: 0,
        wheelToZoom: true,
        maxZoomLevel: 8,
        bgOpacity: 1,
        preload: [1, 1],
        mainClass: "p-lightbox__pswp",
        tapAction: (point, ev) => this.onContentTap(ev),
        imageClickAction: (point, ev) => this.onContentClick(ev),
        bgClickAction: (point, ev) => this.onBgClick(ev),
        paddingFn: (viewport, data) => this.getPadding(viewport, data),
        getViewportSizeFn: () => this.getViewport(),
        closeTitle: this.$gettext("Close"),
        zoomTitle: this.$gettext("Zoom in/out"),
        arrowPrevTitle: this.$gettext("Previous"),
        arrowNextTitle: this.$gettext("Next"),
        errorMsg: this.$gettext("Error"),
      };
    },
    // Displays the thumbnail images and/or videos that belong to the specified models in the lightbox.
    showThumbs(models, index = 0, ctx = {}) {
      if (this.isBusy("show thumbs")) {
        return Promise.reject();
      }

      // Check if at least one model was passed, as otherwise no content can be displayed.
      if (!Array.isArray(models) || models.length === 0 || index >= models.length) {
        this.log("model list passed to lightbox is empty:", models);
        return Promise.reject();
      }

      // Show and initialize the component.
      this.$event.subscribeOnce("lightbox.enter", () => {
        this.renderLightbox(models, index, ctx)
          .then(() => {
            this.busy = false;
          })
          .catch(() => {
            this.busy = false;
            this.close();
          });
      });

      this.showDialog();

      return Promise.resolve();
    },
    // Loads the pictures that belong to a component and displays them in the lightbox.
    showView(view, index) {
      if (this.isBusy("show context")) {
        return Promise.reject();
      }

      if (view.loading || !view.listen || view.lightbox.loading || !view.results[index]) {
        return Promise.reject();
      }

      // Get collection model from view, if any.
      const collection = view.model && view.model instanceof Collection ? view.model : null;
      const context = view.getContext && typeof view.getContext === "function" ? view.getContext() : "";
      const selected = view.results[index];

      if (!view.lightbox.dirty && view.lightbox.results && view.lightbox.results.length > index) {
        // Reuse existing lightbox result if possible.
        let i = -1;

        if (view.lightbox.results[index] && view.lightbox.results[index].UID === selected.UID) {
          i = index;
        } else {
          i = view.lightbox.results.findIndex((p) => p.UID === selected.UID);
        }

        if (
          i > -1 &&
          (((view.lightbox.complete || view.complete) && view.lightbox.results.length >= view.results.length) ||
            i + view.lightbox.batchSize <= view.lightbox.results.length)
        ) {
          return this.showThumbs(view.lightbox.results, i, { collection, context });
        }
      }

      // Fetch photos from server API.
      view.lightbox.loading = true;

      const params = view.searchParams();
      params.count = params.offset + view.lightbox.batchSize;
      params.offset = 0;

      // Fetch lightbox results from API.
      return $api
        .get("photos/view", { params })
        .then((response) => {
          const count = response && response.data ? response.data.length : 0;
          if (count === 0) {
            view.$notify.warn(view.$gettext("No pictures found"));
            view.lightbox.dirty = true;
            view.lightbox.complete = false;
            return;
          }

          // Process response.
          if (response.headers && response.headers["x-count"]) {
            const c = parseInt(response.headers["x-count"]);
            const l = parseInt(response.headers["x-limit"]);
            view.lightbox.complete = c < l;
          } else {
            view.lightbox.complete = view.complete;
          }

          let i;

          if (response.data[index] && response.data[index].UID === selected.UID) {
            i = index;
          } else {
            i = response.data.findIndex((p) => p.UID === selected.UID);
          }

          view.lightbox.results = Thumb.wrap(response.data);

          // Show pictures.
          this.showThumbs(view.lightbox.results, i, { collection, context });
          view.lightbox.dirty = false;
        })
        .catch(() => {
          view.lightbox.dirty = true;
          view.lightbox.complete = false;
        })
        .finally(() => {
          // Unblock.
          view.lightbox.loading = false;
        });
    },
    getNumItems() {
      return this.models.length;
    },
    getItemData(el, i) {
      // Get the slide model.
      const model = this.models[i];

      // Get the estimated slide (viewport) size in real pixels.
      const pixels = this.getSlidePixels(model);

      // Find thumbnail size that best matches the current slide size and zoom level.
      const thumb = this.$util.thumb(model.Thumbs, pixels.width, pixels.height);

      // Set thumbnail image URL, width, and height.
      const img = {
        src: thumb.src,
        width: thumb.w,
        height: thumb.h,
        alt: model?.Title,
        model: model,
        loading: false,
      };

      // Check if content is playable and return the data needed to render it in "contentLoad".
      if (model?.Playable && model?.Hash) {
        /*
          TODO: The server should (additionally) provide a video/animation still from time index 0 that can be used as
                poster (the current thumbnail is taken later for longer videos, since the first frame is often black).
         */

        // Check the duration so that short videos can be looped, unless a slideshow is playing.
        const isShort = model?.Duration
          ? model.Duration > 0 && model.Duration <= this.shortVideoDuration * 1000000000
          : false;

        // Set the slide data needed to render and play the video.
        const video = {
          type: "html", // Render custom HTML.
          html: `<div class="pswp__html"></div>`, // Replaced with the <video> element.
          model: model, // Content model.
          duration: model.Duration > 0 ? model.Duration / 1000000000 : 0,
          format: this.$util.videoFormat(model?.Codec, model?.Mime), // Content format.
          loop: model?.Type !== media.Live && (isShort || model?.Type === media.Animated), // If possible, loop these types.
          msrc: img.src, // Image URL.
          loading: true,
        };

        if (model?.Type === media.Live) {
          video.width = img.width;
          video.height = img.height;
        }

        return video;
      }

      // Return the image data so that PhotoSwipe can render it in the lightbox,
      // see https://photoswipe.com/data-sources/#dynamically-generated-data.
      return img;
    },
    isContentZoomable(isContentZoomable, content) {
      if (content.data?.model?.Type === media.Live) {
        isContentZoomable = true;
      }

      return isContentZoomable;
    },
    onContentLoad(ev) {
      const { content } = ev;
      if (content.data?.type === "html") {
        // Prevent default loading behavior.
        ev.preventDefault();

        try {
          // Create pswp__media element.
          const mediaElement = document.createElement("div");
          mediaElement.setAttribute("class", "pswp__media");
          mediaElement.classList.add(`pswp__media--${content.data.model.Type}`);

          // Create and append video player.
          mediaElement.appendChild(this.createVideoElement(content, false, false, false));

          // Create and append cover image.
          if (content.data.msrc) {
            const imageElement = document.createElement("img");
            imageElement.setAttribute("src", content.data.msrc);
            imageElement.setAttribute("class", "pswp__image");
            mediaElement.appendChild(imageElement);
          }

          // Create pswp__play button element.
          const buttonElement = document.createElement("i");
          buttonElement.setAttribute("class", "pswp__play mdi-play mdi v-icon v-theme--default clickable");
          mediaElement.appendChild(buttonElement);

          content.element = mediaElement;
          content.state = "loading";
          content.data.loading = false;
          content.onLoaded();
        } catch (err) {
          this.log("failed to load video", err);
        }
      }
    },
    onContentDestroy(ev) {
      if (typeof ev?.content?.data?.events === "object") {
        const data = ev.content.data;

        if (this.debug) {
          this.log(`content.destroy`, data);
        }

        // Remove video event listeners.
        data.events?.abort();
        data.events = null;
      }
    },
    // Creates an HTMLMediaElement for playing videos, animations, and live photos.
    createVideoElement(content, autoplay = false, loop = false, mute = false) {
      const data = content.data;
      const model = data.model;
      const format = data.format;
      const posterSrc = data.msrc;

      // See https://developer.mozilla.org/en-US/docs/Web/API/HTMLMediaElement.
      const video = document.createElement("video");

      // Check if a slideshow is running.
      const slideshow = this.slideshow.active;

      let preload = "none";

      if (autoplay) {
        preload = "auto";
      } else if (slideshow || loop) {
        preload = "metadata";
      }

      // Set HTMLMediaElement properties.
      video.className = "pswp__video";
      video.poster = posterSrc;
      video.autoplay = Boolean(autoplay);
      video.loop = Boolean(loop && !slideshow);
      video.muted = Boolean(mute || this.muted);
      video.preload = preload;
      video.setAttribute("playsinline", ""); // iOS requires attribute
      video.playsInline = true;
      video.disableRemotePlayback = false;
      video.controls = false;
      video.dir = document.dir ? document.dir : this.$config.dir(this.$isRtl);

      // Create AbortController instance to clean up the event handlers.
      const ctrl = new AbortController();

      // Abort any existing controller.
      data.events?.abort();
      data.events = ctrl;

      // Attach video event handlers.
      VIDEO_EVENT_TYPES.forEach((ev) => {
        video.addEventListener(ev, this.videoEventListener, { signal: ctrl.signal });
      });

      // Create and append video source elements, depending on file format support.
      if (
        format !== media.FormatAvc &&
        model?.Mime &&
        model.Mime !== media.ContentTypeMp4AvcMain &&
        video.canPlayType(model.Mime)
      ) {
        const nativeSource = document.createElement("source");
        nativeSource.type = model.Mime;
        nativeSource.src = this.$util.videoFormatUrl(model.Hash, format);
        video.appendChild(nativeSource);
      } else {
        const avcSource = document.createElement("source");
        avcSource.type = media.ContentTypeMp4AvcMain;
        avcSource.src = this.$util.videoFormatUrl(model.Hash, media.FormatAvc);
        video.appendChild(avcSource);
      }

      // If we set preload programmatically, kick Safari to honor it.
      if (preload !== "none") {
        try {
          video.load();
        } catch (err) {
          if (this.debug) {
            this.log("video.load", { err });
          }
        }
      }

      // Check if remote playback is supported by this browser.
      if (this.featExperimental && video.remote && video.remote instanceof RemotePlayback) {
        if (!this.video.castable) {
          const cancel = () => {
            video.remote
              .cancelWatchAvailability?.(this.videoAvailabilityListener)
              .catch(this.trace ? this.log : () => {});
          };

          ctrl.signal.addEventListener("abort", cancel, { once: true });
          video.remote.watchAvailability(this.videoAvailabilityListener).catch(this.trace ? this.log : () => {});
        }

        // Attach video remote event handlers.
        VIDEO_REMOTE_EVENT_TYPES.forEach((ev) => {
          video.addEventListener(ev, this.videoRemoteListener, { signal: ctrl.signal });
        });
      }

      // Return HTMLMediaElement.
      return video;
    },
    onVideoEvent(ev) {
      const { video, data } = this.getContent();

      if (!video || !data) {
        return;
      } else if (ev && ev.target.src !== video.src) {
        return;
      }

      return this.setVideo(video, data, ev);
    },
    toggleVideoRemote() {
      const { video, data } = this.getContent();

      if (!video || !data) {
        return;
      } else if (!video.remote || !(video.remote instanceof RemotePlayback)) {
        return;
      }

      if (video.remote.state === "connected" || video.remote.state === "disconnected") {
        try {
          video.remote.prompt().catch((err) => {
            this.onVideoRemoteError(err);
          });
        } catch (err) {
          this.onVideoRemoteError(err);
        }
      }
    },
    onVideoRemoteError(err) {
      if (!err) {
        return;
      }

      if (err instanceof DOMException) {
        switch (err.name) {
          case "NotSupportedError":
            this.$notify.error(this.$gettext("Not supported"));
            break;
          case "NotFoundError":
            this.$notify.error(this.$gettext("Not found"));
            break;
          case "NotAllowedError":
            this.$notify.error(this.$gettext("Not allowed"));
            break;
          default:
            this.$notify.error(err.message);
        }
      } else {
        this.log("video.remote", { err });
      }
    },
    onVideoRemote(ev) {
      const { video, data } = this.getContent();

      if (!video || !data) {
        return;
      } else if (ev && ev.target.src !== video.src) {
        return;
      }

      this.video.casting = ev.state === "connected";
      this.video.remote = ev.state;
    },
    setVideo(video, data, ev) {
      if (!data) {
        return;
      } else if (!video) {
        this.resetVideo();
        return;
      }

      if (video.src !== this.video.src) {
        this.resetVideo();
      }

      let isPlaying =
        video.readyState && !video.paused && !video.ended && !video.waiting && (!video.error || video.error.code === 0);

      if (ev && ev.type) {
        switch (ev.type) {
          case "playing":
            // Automatically hide the lightbox controls after a video has started playing.
            this.hideControlsWithDelay(this.playControlHideDelay);
            this.video.waiting = false;
            isPlaying = true;
            break;
          case "ended":
          case "pause":
            this.video.waiting = false;
            video.parentElement.classList.remove("is-playing");
            video.parentElement.classList.remove("is-waiting");
            break;
          case "abort":
          case "error":
            this.video.waiting = false;
            video.parentElement.classList.add("is-broken");
            video.parentElement.classList.remove("is-playing");
            video.parentElement.classList.remove("is-waiting");
            break;
          case "timeupdate":
          case "loadeddata":
          case "loadedmetadata":
            this.video.waiting = false;
            video.parentElement.classList.remove("is-waiting");
            break;
          case "waiting":
            this.video.waiting = true;
            video.parentElement.classList.add("is-waiting");
        }

        // Automatically hide the lightbox controls after a video has started playing.
        if (ev.type === "ended") {
          if (!this.slideshow.active) {
            this.showControls();
          } else {
            this.clearSlideshowInterval();
            this.onSlideshowNext();
            this.setSlideshowInterval();
          }
        }
      }

      // URL of the currently playing video.
      this.video.src = video.src;

      // Loop short videos of 5 seconds or less, even if the server does not know the duration.
      if (
        !data.loop &&
        video.duration &&
        video.duration <= this.shortVideoDuration &&
        data.model?.Type !== media.Live
      ) {
        data.loop = true;
        video.loop = data.loop && !this.slideshow.active;
      }

      // Do not display video controls if a slideshow is running,
      // or the video belongs to an animation or live photo.
      this.video.controls =
        !this.slideshow.active && data.model?.Type !== media.Animated && data.model?.Type !== media.Live;

      // Get video playback error, if any:
      // https://developer.mozilla.org/de/docs/Web/API/HTMLMediaElement/error
      if (video.error && video.error instanceof MediaError && video.error.code > 0) {
        if (this.debug) {
          this.log("video.error", video.error);
        }

        switch (video.error.code) {
          case 1:
            this.$notify.error(this.$gettext("Something went wrong, try again"));
            break;
          case 2:
            this.video.error = this.$gettext("Request failed - are you offline?");
            break;
          case 3:
          case 4:
            this.video.error = this.$gettext("Not supported");
            break;
          default:
            this.video.error = video.error.message;
        }

        video.parentElement.classList.add("is-broken");
        this.video.errorCode = video.error.code;
      } else {
        video.parentElement.classList.remove("is-broken");
        this.video.error = "";
        this.video.errorCode = 0;
      }

      this.video.state = video.readyState;

      if (this.video.time !== video.currentTime) {
        this.video.time = video.currentTime;
      }

      this.video.duration = video.duration ? video.duration : data.duration;
      this.video.seeking = video.seeking === true;

      // Enable seeking if the video has a seekable time range.
      if (video.seekable && video.seekable instanceof TimeRanges) {
        this.video.seekable = video.readyState > 0 && video.seekable.length > 0;
      }

      // Disable seeking if video doesn't load or there is an error.
      if (video.readyState < 1 || this.video.errorCode > 0) {
        this.video.seekable = false;
      }

      this.video.paused = video.paused;
      this.video.ended = video.ended;
      this.video.playing = isPlaying;

      if (this.video.playing) {
        video.parentElement.classList.add("is-playing");
        video.parentElement.classList.remove("is-waiting");
        video.parentElement.classList.remove("is-broken");
      }
    },
    resetVideo(showControls = false) {
      this.video = {
        controls: !!showControls,
        src: "",
        error: "",
        errorCode: 0,
        state: 0,
        time: 0,
        duration: 0,
        seeking: false,
        seekable: false,
        waiting: false,
        playing: false,
        paused: false,
        ended: false,
        castable: this.video.castable,
        casting: false,
        remote: "",
      };
    },
    // Initializes and opens the PhotoSwipe lightbox with the
    // images and/or videos that belong to the specified models.
    renderLightbox(models, index = 0, ctx) {
      // Check if at least one model was passed, as otherwise no content can be displayed.
      if (!Array.isArray(models) || models.length === 0 || index >= models.length) {
        this.log("model list is empty", models);
        return Promise.reject();
      }

      // Set collection model (e.g. album, label) and view context, if any.
      const collectionModel = ctx?.collection ?? ctx?.album ?? null;
      this.collection = collectionModel instanceof Collection ? collectionModel : null;
      this.context = ctx?.context ?? "";

      // Set the model list and start index.
      // TODO: In the future, additional models should be dynamically loaded when the index reaches the end of the list.
      if (this.$isRtl) {
        // Reverse the slide direction for right-to-left languages.
        this.models = models.toReversed();
        this.index = models.length - (index + 1);
      } else {
        // Keep direction for left-to-right languages.
        this.models = models.slice();
        this.index = index;
      }

      // Get PhotoSwipe lightbox config options, see https://photoswipe.com/options/.
      const options = this.getOptions();

      if (!options.appendToEl) {
        this.log("content element not found", options);
        return Promise.reject();
      }

      // Focus content element.
      this.focusContent();

      // Create PhotoSwipe instance.
      let lightbox = new Lightbox(options);
      let firstPicture = true;

      // Keep reference to PhotoSwipe instance.
      this.lightbox = lightbox;
      this.idleTimer = false;

      // Use dynamic caption plugin,
      // see https://github.com/dimsemenov/photoswipe-dynamic-caption-plugin.
      this.captionPlugin = new Captions(this.lightbox, {
        type: "below",
        mobileLayoutBreakpoint: 1024,
        captionContent: (slide) => {
          if (!slide || !this.models || slide?.index < 0) {
            return "";
          }

          const model = this.models[slide.index];

          if (model) {
            return this.formatCaption(model);
          }

          return "";
        },
      });

      // Register animation event handlers to prevent user actions during animations,
      // see https://photoswipe.com/events/#opening-or-closing-transition-events.
      this.lightbox.on("openingAnimationStart", () => {
        if (this.debug) {
          this.log("start opening animation");
        }
        this.busy = true;
      });
      this.lightbox.on("openingAnimationEnd", () => {
        this.busy = false;
        if (this.debug) {
          this.log("end opening animation");
        }
      });
      this.lightbox.on("closingAnimationStart", () => {
        if (this.debug) {
          this.log("start closing animation");
        }
        this.busy = true;
      });
      this.lightbox.on("closingAnimationEnd", () => {
        this.busy = false;
        if (this.debug) {
          this.log("end closing animation");
        }
      });

      // Add a custom pointer event handler to prevent the default
      // action when events are triggered on an HTMLMediaElement.
      this.lightbox.on("pointerUp", this.lightboxPointerListener);
      this.lightbox.on("pointerDown", this.lightboxPointerListener);
      // this.lightbox.on("pointerMove", this.lightboxPointerListener);

      // Add PhotoSwipe lightbox controls,
      // see https://photoswipe.com/adding-ui-elements/.
      this.addLightboxControls();

      // Handle zoom level changes to load higher quality thumbnails
      // when image size changes
      this.lightbox.on("imageSizeChange", ({ slide }) => {
        if (slide.isActive) {
          this.onImageSizeChange();
        }
      });

      // Trigger onChange() event handler when slide is changed and on initialization,
      // see https://photoswipe.com/events/#initialization-events.
      this.lightbox.on("change", this.onChange.bind(this));
      // this.lightbox.on("loadComplete", this.onLoadComplete.bind(this));

      // Processes model data for rendering slides with PhotoSwipe,
      // see https://photoswipe.com/filters/#itemdata.
      this.lightbox.addFilter("numItems", this.getNumItems.bind(this));
      this.lightbox.addFilter("itemData", this.getItemData.bind(this));
      this.lightbox.addFilter("isContentZoomable", this.isContentZoomable.bind(this));

      // Renders content when a slide starts to load (can be default prevented),
      // see https://photoswipe.com/events/#slide-content-events.
      this.lightbox.on("contentLoad", this.onContentLoad.bind(this));
      // this.lightbox.on("contentResize", this.onContentResize.bind(this));
      // this.lightbox.on("contentRemove", this.onContentRemove.bind(this));
      this.lightbox.on("contentDestroy", this.onContentDestroy.bind(this));

      // Pauses videos, animations, and live photos when slide content becomes active (can be default prevented),
      // see https://photoswipe.com/events/#slide-content-events.
      this.lightbox.on("contentActivate", (ev) => {
        const { content } = ev;

        if (!content) {
          return;
        }

        const data = typeof content?.data === "object" ? content?.data : {};

        if (!data) {
          return;
        }

        switch (data.model?.Type) {
          case media.Video:
          case media.Animated:
            this.isZoomable = false;
            break;
          default:
            this.isZoomable = true;
        }

        let video;

        // Get <video> element, if any.
        if (content?.element && content?.element.firstElementChild instanceof HTMLMediaElement) {
          video = content.element.firstElementChild;
        } else {
          video = false;
        }

        this.setVideo(video, data);

        // Automatically play video on this slide if it's the first item,
        // a slideshow is active, or it's an animation or live photo.
        if (video) {
          if (data.loop || this.slideshow.active || firstPicture) {
            this.playVideo(content.element.firstElementChild, data.loop);
          }
        }

        // Flag first picture as being displayed/activated.
        if (firstPicture) {
          firstPicture = false;
        }
      });

      // Pauses videos, animations, and live photos when content becomes active (can be default prevented),
      // see https://photoswipe.com/events/#slide-content-events.
      this.lightbox.on("contentDeactivate", (ev) => {
        const { content } = ev;

        // Stop any video currently playing on this slide.
        if (content?.element && content?.element.firstElementChild instanceof HTMLMediaElement) {
          this.pauseVideo(content.element.firstElementChild);
        }
      });

      // Add a close event handler to destroy the lightbox after use,
      // see https://photoswipe.com/events/#closing-events.
      this.lightbox.on("close", this.onLightboxClose.bind(this));

      // Add a destroy event handler to complete the closing of the lightbox,
      // see https://photoswipe.com/events/#closing-events.
      this.lightbox.on("destroy", this.onLightboxDestroyed.bind(this));

      // Add an init event handler that is triggered when PhotoSwipe is fully initialized,
      // see https://photoswipe.com/events/#closing-events.
      this.lightbox.on("afterInit", this.onLightboxOpened.bind(this));

      // Init PhotoSwipe.
      this.lightbox.init();

      // Show first image.
      this.lightbox.loadAndOpen(this.index);
      this.busy = false;

      return Promise.resolve();
    },
    // Adds PhotoSwipe lightbox controls, see https://photoswipe.com/adding-ui-elements/.
    addLightboxControls() {
      const lightbox = this.lightbox;

      // Add a sidebar toggle button only if the window is large enough.
      // TODO: Proof-of-concept only, the sidebar needs to be fully implemented before removing the featDevelop check.
      // TODO: Once this is fully implemented, remove the "this.experimental" flag check below.
      // IDEA: We can later try to add styles that display the sidebar at the bottom
      //       instead of on the side, to allow use on mobile devices.
      lightbox.on("uiRegister", () => {
        // Add close button.
        lightbox.pswp.ui.registerElement({
          name: "close-button",
          className: "pswp__button--close-button", // Sets the icon style/size in lightbox.css.
          title: this.$gettext("Close"),
          ariaLabel: this.$gettext("Close"),
          order: 1,
          isButton: true,
          html: {
            isCustomSVG: true,
            inner: `<path d="M24 10l-2-2-6 6-6-6-2 2 6 6-6 6 2 2 6-6 6 6 2-2-6-6z" id="pswp__icn-close-button" />`,
            outlineID: "pswp__icn-close-button", // Add this to the <path> in the inner property.
            size: 32, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
          },
          onClick: (ev) =>
            this.onControlClick(ev, () => {
              if (this.debug) {
                this.log("pswp.ui.close", ev);
              }

              this.close();
            }),
        });

        // Add information toggle button.
        if (window.innerWidth > this.mobileBreakpoint) {
          lightbox.pswp.ui.registerElement({
            name: "sidebar-button",
            className: "pswp__button--info-button pswp__button--mdi", // Sets the icon style/size in lightbox.css.
            title: this.$gettext("Information"),
            ariaLabel: this.$gettext("Information"),
            order: 9,
            isButton: true,
            html: {
              isCustomSVG: true,
              inner:
                '<path d="M11 7V9H13V7H11M14 17V15H13V11H10V13H11V15H10V17H14M22 12C22 17.5 17.5 22 12 22C6.5 22 2 17.5 2 12C2 6.5 6.5 2 12 2C17.5 2 22 6.5 22 12M20 12C20 7.58 16.42 4 12 4C7.58 4 4 7.58 4 12C4 16.42 7.58 20 12 20C16.42 20 20 16.42 20 12Z" id="pswp__icn-info"/>',
              outlineID: "pswp__icn-info", // Add this to the <path> in the inner property.
              size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
            },
            onClick: (ev) => this.onControlClick(ev, this.toggleInfo),
          });
        }

        // Add sound mute/unmute control for videos.
        lightbox.pswp.ui.registerElement({
          name: "sound-toggle",
          className: "pswp__button--sound-toggle pswp__button--mdi", // Sets the icon style/size in lightbox.css.
          title: this.$gettext("Mute"),
          ariaLabel: this.$gettext("Mute"),
          order: 10,
          isButton: true,
          html: {
            isCustomSVG: true,
            inner: `<use class="pswp__icn-shadow pswp__icn-sound-on" xlink:href="#pswp__icn-sound-on"></use><path d="M14,3.23V5.29C16.89,6.15 19,8.83 19,12C19,15.17 16.89,17.84 14,18.7V20.77C18,19.86 21,16.28 21,12C21,7.72 18,4.14 14,3.23M16.5,12C16.5,10.23 15.5,8.71 14,7.97V16C15.5,15.29 16.5,13.76 16.5,12M3,9V15H7L12,20V4L7,9H3Z" id="pswp__icn-sound-on" class="pswp__icn-sound-on" /><use class="pswp__icn-shadow pswp__icn-sound-off" xlink:href="#pswp__icn-sound-off"></use><path d="M12,4L9.91,6.09L12,8.18M4.27,3L3,4.27L7.73,9H3V15H7L12,20V13.27L16.25,17.53C15.58,18.04 14.83,18.46 14,18.7V20.77C15.38,20.45 16.63,19.82 17.68,18.96L19.73,21L21,19.73L12,10.73M19,12C19,12.94 18.8,13.82 18.46,14.64L19.97,16.15C20.62,14.91 21,13.5 21,12C21,7.72 18,4.14 14,3.23V5.29C16.89,6.15 19,8.83 19,12M16.5,12C16.5,10.23 15.5,8.71 14,7.97V10.18L16.45,12.63C16.5,12.43 16.5,12.21 16.5,12Z" id="pswp__icn-sound-off" class="pswp__icn-sound-off" />`,
            size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
          },
          onClick: (ev) => this.onControlClick(ev, this.toggleMute),
        });

        // Add slideshow play/pause toggle control.
        lightbox.pswp.ui.registerElement({
          name: "slideshow-toggle",
          className: "pswp__button--slideshow-toggle pswp__button--mdi", // Sets the icon style/size in lightbox.css.
          title: this.$gettext("Slideshow"),
          ariaLabel: this.$gettext("Slideshow"),
          order: 10,
          isButton: true,
          html: {
            isCustomSVG: true,
            inner: `<use class="pswp__icn-shadow pswp__icn-slideshow-on" xlink:href="#pswp__icn-slideshow-on"></use><path d="M14,19H18V5H14M6,19H10V5H6V19Z" id="pswp__icn-slideshow-on" class="pswp__icn-slideshow-on" /><use class="pswp__icn-shadow pswp__icn-slideshow-off" xlink:href="#pswp__icn-slideshow-off"></use><path d="M8,5.14V19.14L19,12.14L8,5.14Z" id="pswp__icn-slideshow-off" class="pswp__icn-slideshow-off" />`,
            size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
          },
          onClick: (ev) => this.onControlClick(ev, this.toggleSlideshow),
        });

        // Add fullscreen mode toggle control.
        if (this.canFullscreen) {
          lightbox.pswp.ui.registerElement({
            name: "fullscreen-toggle",
            className: "pswp__button--fullscreen-toggle pswp__button--mdi", // Sets the icon style/size in lightbox.css.
            title: this.$gettext("Fullscreen"),
            ariaLabel: this.$gettext("Fullscreen"),
            order: 10,
            isButton: true,
            html: {
              isCustomSVG: true,
              inner: `<use class="pswp__icn-shadow pswp__icn-fullscreen-on" xlink:href="#pswp__icn-fullscreen-on"></use><path d="M14,14H19V16H16V19H14V14M5,14H10V19H8V16H5V14M8,5H10V10H5V8H8V5M19,8V10H14V5H16V8H19Z" id="pswp__icn-fullscreen-on" class="pswp__icn-fullscreen-on" /><use class="pswp__icn-shadow pswp__icn-fullscreen-off" xlink:href="#pswp__icn-fullscreen-off"></use><path d="M5,5H10V7H7V10H5V5M14,5H19V10H17V7H14V5M17,14H19V19H14V17H17V14M10,17V19H5V14H7V17H10Z" id="pswp__icn-fullscreen-off" class="pswp__icn-fullscreen-off" />`,
              size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
            },
            onClick: (ev) => this.onControlClick(ev, this.toggleFullscreen),
          });
        }

        // Add favorite toggle control if user has permission to use it.
        if (this.canLike) {
          lightbox.pswp.ui.registerElement({
            name: "favorite-toggle",
            className: "pswp__button--favorite-toggle pswp__button--mdi hidden-shared-only", // Sets the icon style/size in lightbox.css.
            title: this.$gettext("Like"),
            ariaLabel: this.$gettext("Like"),
            order: 10,
            isButton: true,
            html: {
              isCustomSVG: true,
              inner: `<use class="pswp__icn-shadow pswp__icn-favorite-on" xlink:href="#pswp__icn-favorite-on"></use><path d="M12,17.27L18.18,21L16.54,13.97L22,9.24L14.81,8.62L12,2L9.19,8.62L2,9.24L7.45,13.97L5.82,21L12,17.27Z" id="pswp__icn-favorite-on" class="pswp__icn-favorite-on" /><use class="pswp__icn-shadow pswp__icn-favorite-off" xlink:href="#pswp__icn-favorite-off"></use><path d="M12,15.39L8.24,17.66L9.23,13.38L5.91,10.5L10.29,10.13L12,6.09L13.71,10.13L18.09,10.5L14.77,13.38L15.76,17.66M22,9.24L14.81,8.63L12,2L9.19,8.63L2,9.24L7.45,13.97L5.82,21L12,17.27L18.18,21L16.54,13.97L22,9.24Z" id="pswp__icn-favorite-off" class="pswp__icn-favorite-off" />`,
              size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
            },
            onClick: (ev) => this.onControlClick(ev, this.toggleLike),
          });
        }

        // Add selection toggle control.
        lightbox.pswp.ui.registerElement({
          name: "select-toggle",
          className: "pswp__button--select-toggle pswp__button--mdi", // Sets the icon style/size in lightbox.css.
          title: this.$gettext("Select"),
          ariaLabel: this.$gettext("Select"),
          order: 10,
          isButton: true,
          html: {
            isCustomSVG: true,
            inner: `<use class="pswp__icn-shadow pswp__icn-select-on" xlink:href="#pswp__icn-select-on"></use><path d="M12 2C6.5 2 2 6.5 2 12S6.5 22 12 22 22 17.5 22 12 17.5 2 12 2M10 17L5 12L6.41 10.59L10 14.17L17.59 6.58L19 8L10 17Z" id="pswp__icn-select-on" class="pswp__icn-select-on" /><use class="pswp__icn-shadow pswp__icn-select-off" xlink:href="#pswp__icn-select-off"></use><path d="M12,20A8,8 0 0,1 4,12A8,8 0 0,1 12,4A8,8 0 0,1 20,12A8,8 0 0,1 12,20M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2Z" id="pswp__icn-select-off" class="pswp__icn-select-off" />`,
            size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
          },
          onClick: (ev) => this.onControlClick(ev, this.toggleSelect),
        });

        // Add edit button control if user has permission to use it.
        if (this.canEdit) {
          lightbox.pswp.ui.registerElement({
            name: "edit-button",
            className: "pswp__button--edit-button pswp__button--mdi hidden-shared-only", // Sets the icon style/size in lightbox.css.
            title: this.$gettext("Edit"),
            ariaLabel: this.$gettext("Edit"),
            order: 10,
            isButton: true,
            html: {
              isCustomSVG: true,
              inner: `<path d="M20.71,7.04C21.1,6.65 21.1,6 20.71,5.63L18.37,3.29C18,2.9 17.35,2.9 16.96,3.29L15.12,5.12L18.87,8.87M3,17.25V21H6.75L17.81,9.93L14.06,6.18L3,17.25Z" id="pswp__icn-edit" />`,
              outlineID: "pswp__icn-edit", // Add this to the <path> in the inner property.
              size: 24, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
            },
            onClick: (ev) => this.onControlClick(ev, this.onEdit),
          });
        }

        // Add an action menu with additional options if there's at least one menu item.
        if (this.menuActions().filter((action) => action.visible).length > 0) {
          lightbox.pswp.ui.registerElement({
            name: "menu-button",
            className: "pswp__button--menu-button pswp__button--mdi", // Sets the icon style/size in lightbox.css.
            ariaLabel: this.$gettext("More options"),
            order: 10,
            isButton: true,
            html: {
              isCustomSVG: true,
              inner: `<path d="M9.5 13a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0zm0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0zm0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0z" id="pswp__icn-menu-button" />`,
              outlineID: "pswp__icn-menu-button", // Add this to the <path> in the inner property.
              size: 16, // Depends on the original SVG viewBox, e.g. use 24 for viewBox="0 0 24 24".
            },
            onInit: (el) => {
              this.menuElement = el;
            },
          });
        }
      });
    },
    // Returns the available menu actions.
    menuActions() {
      return [
        {
          name: "cover",
          icon: "mdi-image-album",
          text: this.$gettext("Set as Album Cover"),
          disabled: !this.model,
          visible:
            this.canManageAlbums &&
            this.collection &&
            this.collection instanceof Collection &&
            !this.model?.Removed &&
            !this.model?.Archived,
          click: () => {
            this.onSetCollectionCover();
          },
        },
        {
          name: "remove",
          icon: "mdi-eject",
          text: this.$gettext("Remove from Album"),
          visible:
            this.canManageAlbums &&
            this.collection &&
            this.collection instanceof Album &&
            this.collection?.Type === "album" &&
            !this.model?.Removed &&
            !this.model?.Archived,
          click: () => {
            this.onRemoveFromAlbum();
          },
        },
        {
          name: "archive",
          icon: "mdi-archive",
          text: this.$pgettext("Verb", "Archive"),
          shortcut: "Ctrl-A",
          disabled: !this.model,
          visible:
            this.canArchive &&
            this.context !== "hidden" &&
            ((this.context !== "archive" && !this.model?.Archived) || this.model?.Archived === false),
          click: () => {
            this.onArchive();
          },
        },
        {
          name: "restore",
          icon: "mdi-archive-arrow-up",
          text: this.$gettext("Restore"),
          shortcut: "Ctrl-A",
          disabled: !this.model,
          visible:
            this.canArchive &&
            this.context !== "hidden" &&
            (this.model?.Archived || (this.context === "archive" && this.model?.Archived !== false)),
          click: () => {
            this.onRestore();
          },
        },
        {
          name: "download",
          icon: "mdi-download",
          text: this.$gettext("Download"),
          shortcut: "Ctrl-D",
          disabled: !this.model,
          visible: this.canDownload,
          click: () => {
            this.onDownload();
          },
        },
      ];
    },
    onShowMenu() {
      this.pauseSlideshow();
      this.menuVisible = true;
    },
    onHideMenu() {
      this.menuVisible = false;
    },
    close() {
      if (this.closing) {
        return new Promise((resolve) => {
          this.$event.subscribeOnce("lightbox.leave", resolve);
        });
      }

      this.closing = true;

      if (this.lightbox) {
        return new Promise((resolve) => {
          this.$event.subscribeOnce("lightbox.leave", resolve);
          setTimeout(() => {
            this.destroyLightbox();
          }, 150);
        });
      }

      return new Promise((resolve) => {
        this.$event.subscribeOnce("lightbox.leave", resolve);
        this.hideDialog();
      });
    },
    onLightboxOpened() {
      this.addEventListeners();
      this.$event.publish("lightbox.opened");
    },
    onLightboxClose() {
      this.$event.publish("lightbox.pause");
      this.$event.publish("lightbox.close");
    },
    // Destroys the PhotoSwipe lightbox instance after use, see onClose().
    destroyLightbox() {
      this.$nextTick(() => {
        if (this.lightbox) {
          this.lightbox.destroy();
          return;
        }

        this.hideDialog();
      });
    },
    onLightboxDestroyed() {
      // Remove lightbox reference.
      this.lightbox = null;

      // Hide lightbox and sidebar.
      this.$nextTick(() => {
        this.hideDialog();
      });
    },
    // Returns the picture (model) caption as sanitized HTML, if any.
    formatCaption(model) {
      if (!model) {
        return "";
      }

      let caption = "";

      if (model.Title) {
        caption += `<h4>${this.$util.encodeHTML(model.Title.trim())}</h4>`;
      }

      /*
        TODO: Find a good position for the date information that works for all screen sizes and image dimensions.
              We MAY postpone this and display it along with other metadata in the new sidebar.
       */
      /* if (model.TakenAtLocal) {
         caption += `<div>${this.$util.formatDate(model.TakenAtLocal)}</div>`;
      } */

      if (model.Description && !model.Caption) {
        model.Caption = model.Description;
      }

      let text = typeof model.Caption === "string" ? model.Caption.trim() : "";

      if (text) {
        if (!caption && text.split("\n").length < 2) {
          // Render large caption if there is no title and it has only one line.
          caption += `<h4>${this.$util.encodeHTML(text)}</h4>`;
        } else {
          // Render small caption otherwise.
          caption += `<p>${this.$util.encodeHTML(text)}</p>`;
        }
      }

      return this.$util.sanitizeHtml(caption);
    },
    // Removes any event listeners before the lightbox is fully closed.
    onClose() {
      // Exit full screen mode only if it was not previously enabled.
      if (!this.wasFullscreen) {
        this.exitFullscreen();
      }

      this.clearTimeouts();
      this.removeEventListeners();
      this.closing = true;
    },
    // Resets the component state after closing the lightbox.
    onReset() {
      this.resetControls();
      this.resetModels();
    },
    // Resets the state of the lightbox controls.
    resetControls() {
      this.controlsShown = -1;
    },
    // Reset the lightbox models and index.
    resetModels() {
      this.collection = null;
      this.context = "";
      this.model = new Thumb();
      this.models = [];
      this.index = 0;
    },
    // Returns the active PhotoSwipe instance, if any.
    // Be sure to check the result before using it!
    pswp() {
      return this.lightbox?.pswp;
    },
    // Called when the slide is changed and on initialization,
    // see https://photoswipe.com/events/#initialization-events.
    onChange() {
      // Get active PhotoSwipe instance.
      const pswp = this.pswp();

      if (!pswp) {
        return;
      }

      // Hide action menu when slide changes.
      if (this.$refs.menu) {
        this.$refs.menu.hide();
      }

      // Set current slide (model) list index.
      if (typeof pswp.currIndex === "number") {
        this.index = pswp.currIndex;
      }

      // Set current slide model.
      if (this.index >= 0 && this.models.length > 0 && this.index < this.models.length) {
        this.model = this.models[this.index];
      }

      // Pause the slideshow if the index of the next slide does not match.
      if (this.slideshow.next !== this.index) {
        this.pauseSlideshow();
      }

      // Ensure that content is focused.
      this.focusContent();
    },
    // Called when the user clicks on the PhotoSwipe lightbox background,
    // see https://photoswipe.com/click-and-tap-actions.
    onBgClick(ev) {
      if (!ev) {
        return;
      }

      if (this.debug) {
        this.log(`background.${ev?.type}`, { ev });
      }

      if (this.controlsVisible()) {
        this.close();
      } else {
        this.showControls();
      }
    },
    // Returns the type of control if the event originates
    // from a PhotoSwipe UI control, like the close button.
    pswpControl(ev) {
      if (!ev) {
        return false;
      }

      let target;

      if (ev.originalEvent?.target) {
        target = ev.originalEvent.target;
      } else if (ev.target) {
        target = ev.target;
      } else {
        return false;
      }

      if (typeof target.closest === "function") {
        if (target.closest(".pswp__button--close-button")) {
          if (this.debug) {
            this.log(`${ev?.type} on close`, { ev });
          }

          return "close";
        }

        if (target.closest(".pswp__button")) {
          if (this.debug) {
            this.log(`${ev?.type} on button`, { ev });
          }

          return "button";
        }

        if (target.closest(".pswp__top-bar")) {
          if (this.debug) {
            this.log(`${ev?.type} on top-bar`, { ev });
          }

          return "top-bar";
        }
      }

      return false;
    },
    // Called when the lightbox receives a pointer down or up event.
    // Move events are ignored for now.
    onLightboxPointerEvent(ev, action) {
      if (!ev || !ev.originalEvent?.target) {
        return;
      }

      const target = ev.originalEvent.target;

      if (this.debug) {
        this.log(`pointer.${ev.type}`, { ev, target, action });
      }

      // Close the lightbox when the user clicks the close button if it is visible.
      const pswpControl = this.pswpControl(ev);
      if (pswpControl === "close") {
        if (this.controlsVisible()) {
          ev.preventDefault();
          this.close();
        }
        return;
      }

      if (target.closest(".pswp__dynamic-caption")) {
        ev.preventDefault();
      }
    },
    // Handle user clicks on a control. Does not reliably work for the close button.
    onControlClick(ev, action) {
      if (!ev) {
        return;
      }

      if (this.debug) {
        this.log(`control.${ev.type}`, { ev, action });
      }

      if (ev && ev.cancelable) {
        ev.stopPropagation();
        ev.preventDefault();
      }

      if (typeof action === "function") {
        if (this.isBusy(action.name)) {
          return false;
        } else if (this.controlsVisible()) {
          action();
          return true;
        } else {
          this.log(`controls not visible, will not call ${action.name}`);
        }
      }

      return false;
    },
    // Capture click events on the dialog component.
    captureDialogClick(ev) {
      if (!ev) {
        return;
      }

      if (this.debug) {
        this.log(`dialog.capture.${ev.type}`, { ev, target: ev.target });
      }

      // Reveal the controls when the user clicks or touches the top of the screen,
      // where they are located when visible.
      if (ev.y <= 128) {
        if (!this.controlsVisible()) {
          ev.stopPropagation();
          ev.preventDefault();
          this.clearIdleTimeout();
          this.showControls();
        }
      } else if (ev.target instanceof HTMLMediaElement) {
        ev.stopPropagation();
        ev.preventDefault();
      }
    },
    // Capture pointer down events on the dialog component.
    captureDialogPointerDown(ev) {
      if (!ev) {
        return;
      }

      if (this.debug) {
        this.log(`dialog.capture.${ev.type}`, { ev, target: ev.target });
      }

      // Handle the click and touch events on custom content.
      if (
        ev.target instanceof HTMLMediaElement ||
        (ev.target instanceof HTMLElement &&
          (ev.target.classList.contains("pswp__image") || ev.target.classList.contains("pswp__play")))
      ) {
        // Always stop slideshow after user interaction with the content.
        if (this.slideshow.active) {
          this.pauseSlideshow();
        }

        // On touch devices, trigger the default event on the sides and when content is zoomed.
        if (this.hasTouch) {
          const { slide } = this.getContent();

          if (slide.currZoomLevel !== slide.zoomLevels.initial) {
            return;
          } else if (ev.clientX && window.innerWidth) {
            const x = ev.clientX / window.innerWidth;

            // Let PhotoSwipe handle the left and right 30% of the screen.
            if (x <= 0.3 || x >= 0.7) {
              return;
            }
          }
        }

        // Toggle video playback.
        this.clearIdleTimeout();
        ev.stopPropagation();
        ev.preventDefault();
        this.toggleVideo();
      }
    },
    // Handle user clicks on an image slide in the lightbox.
    onContentClick(ev) {
      if (!ev) {
        return;
      }

      if (this.debug) {
        this.log(`content.${ev.type}`, { ev, target: ev.target, originalTarget: ev.originalEvent?.target });
      }

      if (this.slideshow.active) {
        this.pauseSlideshow();
      }

      const pswp = this.pswp();

      const isZoomable = pswp.currSlide.isZoomable();

      if (isZoomable) {
        pswp.currSlide.toggleZoom();
      }
    },
    // Handle user taps on an image slide in the lightbox.
    onContentTap(ev) {
      if (!ev) {
        return;
      }

      if (this.debug) {
        this.log(`content.${ev.type}`, { ev, target: ev.target, originalTarget: ev.originalEvent?.target });
      }

      if (ev.target instanceof HTMLMediaElement) {
        // Do nothing.
      } else {
        ev.stopPropagation();
        ev.preventDefault();
        this.toggleControls();
      }
    },
    // Toggles fullscreen mode.
    toggleFullscreen() {
      if ($fullscreen.isEnabled()) {
        this.exitFullscreen();
      } else {
        this.requestFullscreen();
      }
    },
    // Returns true if fullscreen mode is enabled.
    isFullscreen() {
      // see https://developer.mozilla.org/en-US/docs/Web/API/Document/fullscreenElement
      return $fullscreen.isEnabled();
    },
    // Exits fullscreen mode if enabled.
    exitFullscreen() {
      $fullscreen
        .exit()
        .then(() => {
          this.resize(true);
        })
        .catch((err) => console.error(err));
    },
    // Switches to fullscreen mode if not already enabled.
    requestFullscreen() {
      $fullscreen.request().then(() => {
        this.resize(true);
      });
    },
    // Toggles the favorite flag of the current picture.
    toggleLike() {
      this.model.toggleLike();
    },
    // Toggles the selection of the current picture in the global photo clipboard.
    toggleSelect() {
      this.$clipboard.toggle(this.model);
    },
    // Returns the active HTMLMediaElement element in the lightbox, if any.
    getContent() {
      const result = { slide: null, content: null, data: null, video: null };
      const pswp = this.pswp();

      if (!pswp) {
        return result;
      }

      result.slide = pswp?.currSlide;
      result.content = pswp?.currSlide?.content;

      if (!result.slide || !result.content) {
        return result;
      }

      result.data = typeof result.content.data === "object" ? result.content.data : {};

      // Get <video> element, if any.
      if (result.content.element && result.content.element.firstElementChild instanceof HTMLMediaElement) {
        result.video = result.content.element.firstElementChild;
      }

      return result;
    },
    // Finds and pauses an actively playing video, e.g. before closing the lightbox.
    pausePlaying() {
      // Get active video element, if any.
      const { video } = this.getContent();

      if (!video) {
        return;
      }

      // Calling pause() before a play promise has been resolved may result in an error,
      // see https://github.com/flutter/flutter/issues/136309 (we'll ignore this for now).
      if (!video.paused) {
        try {
          video.pause();
        } catch (err) {
          if (this.debug) {
            this.log("video.pause", { err });
          }
        }
        video.parentElement?.classList.remove("is-playing");
      }
    },
    // Starts playback on the specified video element, if any.
    playVideo(video, loop) {
      if (!video || !(video instanceof HTMLMediaElement)) {
        return;
      }

      if (video.error && video.error instanceof MediaError && video.error.code > 0) {
        return;
      }

      if (video.preload === "none") {
        video.preload = "auto";
        try {
          video.load();
        } catch (err) {
          if (this.debug) {
            this.log("video.load", { err });
          }
        }
      }

      video.loop = loop && !this.slideshow.active;
      video.muted = this.muted;

      if (this.muted) {
        video.setAttribute("muted", "");
      } else {
        video.removeAttribute("muted");
      }

      if (video.paused || video.ended) {
        try {
          requestAnimationFrame(() => {
            requestAnimationFrame(async () => {
              const playPromise = video.play();
              if (playPromise !== undefined) {
                playPromise.catch((err) => {
                  if (this.trace && err && err.message) {
                    this.log("video.play", { err });
                  }
                });
              }
            });
          });
        } catch {
          // Ignore.
        }
      }
    },
    // Handles Ctrl/Cmd + key combinations.
    onShortCut(ev) {
      if (this.trace) {
        this.log("shortcut", { ev });
      }

      switch (ev.code) {
        case "Escape":
          this.close();
          return true;
        case "Period":
          this.onShowMenu();
          this.toggleSelect();
          return true;
        case "KeyA":
          if (this.canArchive && this.context !== "hidden") {
            if (this.model.Archived || (this.context === "archive" && this.model?.Archived !== false)) {
              this.onRestore();
            } else {
              this.onArchive();
            }
          }
          return true;
        case "KeyD":
          if (this.canDownload) {
            this.onDownload();
          }
          return true;
        case "KeyE":
          if (this.canEdit) {
            this.onEdit();
          }
          return true;
        case "KeyF":
          if (this.canFullscreen) {
            this.toggleFullscreen();
          }
          break;
        case "KeyI":
          this.toggleInfo();
          return true;
        case "KeyL":
          this.onShowMenu();
          if (this.canLike) {
            this.toggleLike();
          }
          return true;
        case "KeyM":
          this.toggleMute();
          return true;
        case "KeyS":
          this.toggleSlideshow();
          return true;
      }
    },
    // Handles other key events.
    onKeyDown(ev) {
      if (!ev || !ev.code || !this.visible || !this.$view.isActive(this)) {
        return;
      }

      if (
        this.info &&
        (document.activeElement instanceof HTMLInputElement || document.activeElement instanceof HTMLTextAreaElement)
      ) {
        return;
      }

      if (this.trace) {
        this.log("key.down", { ev });
      }

      this.pauseSlideshow();

      // Handle space and escape key events.
      switch (ev.code) {
        case "ArrowLeft":
          ev.preventDefault();
          ev.stopPropagation();
          if (this.model?.Playable && this.video.controls && this.video.playing) {
            this.seekVideoSeconds(this.$isRtl ? 10 : -10);
          } else if (this.index > 0) {
            this.pswp().prev();
          }
          break;
        case "ArrowRight":
          ev.preventDefault();
          ev.stopPropagation();

          if (this.model?.Playable && this.video.controls && this.video.playing) {
            this.seekVideoSeconds(this.$isRtl ? -10 : 10);
          } else if (this.models.length > this.index + 1) {
            this.pswp().next();
          }
          break;
        case "Space":
          ev.preventDefault();
          ev.stopPropagation();

          // Get active video element, if any.
          const { video } = this.getContent();

          if (video) {
            this.toggleVideo();
          } else {
            this.toggleControls();
          }
          break;
      }
    },
    // Toggles video playback on the current video element, if any.
    toggleVideo() {
      // Get active video element, if any.
      const { data, video } = this.getContent();

      if (!video) {
        return;
      }

      // Play video if it is currently paused and pause it otherwise.
      if (video.paused || video.ended) {
        this.playVideo(video, data.loop);
      } else {
        this.pauseVideo(video);
      }
    },
    // Jumps to the specified time index when a video is loaded and seekable.
    seekVideo(seekTo) {
      if (Number.isNaN(seekTo)) {
        return false;
      }

      // Get active video element, if any.
      const { video } = this.getContent();

      if (!video) {
        return false;
      } else if (!video?.readyState || video.readyState < 1 || !video.duration || !this.video.seekable) {
        return;
      }

      // If possible, use the fastSeek() method to quickly jump to the new time index:
      // https://developer.mozilla.org/en-US/docs/Web/API/HTMLMediaElement/fastSeek
      if (typeof video.fastSeek === "function") {
        if (seekTo >= video.duration - 0.01) {
          video.loop = false;
          video.fastSeek(video.duration);
          this.pauseVideo(video);
        } else if (seekTo <= 0) {
          video.loop = false;
          video.fastSeek(0);
          this.pauseVideo(video);
        } else {
          video.fastSeek(seekTo);
        }
      } else {
        if (seekTo >= video.duration - 0.01) {
          video.loop = false;
          video.currentTime = video.duration;
          this.pauseVideo(video);
        } else if (seekTo <= 0) {
          video.loop = false;
          video.currentTime = 0;
          this.pauseVideo(video);
        } else {
          video.currentTime = seekTo;
        }
      }

      return true;
    },
    // Skips the specified number of seconds when a video is loaded and seekable.
    seekVideoSeconds(seconds) {
      if (!seconds || Number.isNaN(seconds)) {
        return false;
      } else if (!this.video.playing) {
        return false;
      }

      // Get active video element, if any.
      const { video } = this.getContent();

      if (!video || !video.currentTime) {
        return false;
      }

      this.seekVideo(video.currentTime + seconds);

      return true;
    },
    // Stops playback on the specified video element, if any.
    pauseVideo(video) {
      if (!video || !(video instanceof HTMLMediaElement)) {
        return;
      }

      if (!video.paused) {
        try {
          video.pause();
        } catch (err) {
          if (this.debug) {
            this.log("video.pause", { err });
          }
        }
        video.parentElement?.classList.remove("is-playing");
        this.showControls();
      }
    },
    // Mutes/unmutes the sound for videos.
    toggleMute() {
      this.muted = !this.muted;

      window.sessionStorage.setItem("lightbox.muted", this.muted.toString());

      const { video } = this.getContent();

      if (!video) {
        return;
      }

      video.muted = this.muted;

      if (this.muted) {
        video.setAttribute("muted", "");
      } else {
        video.removeAttribute("muted");
      }
    },
    // Starts/stops a slideshow so that the next slide opens automatically at regular intervals.
    toggleSlideshow() {
      if (this.slideshow.active || this.slideshow.interval) {
        this.pauseSlideshow();
      } else {
        this.playSlideshow();
      }
    },
    // Starts a slideshow, if not already active.
    playSlideshow() {
      // Return if already playing.
      if (this.slideshow.active) {
        return;
      }

      // Flag slideshow as active.
      this.slideshow.active = true;

      const { video } = this.getContent();

      // Play video, if any, but without looping.
      if (video) {
        this.playVideo(video, false);
      }

      // Show next slide at regular intervals.
      this.setSlideshowInterval();
    },
    setSlideshowInterval() {
      this.clearSlideshowInterval();

      if (!this.slideshow.active) {
        return;
      }

      this.slideshow.interval = setInterval(() => {
        this.onSlideshowNext();
      }, this.slideshow.wait);
    },
    clearSlideshowInterval() {
      if (this.slideshow.interval) {
        clearInterval(this.slideshow.interval);
        this.slideshow.interval = false;
      }
    },
    onSlideshowNext() {
      // Get PhotoSwipe instance.
      const pswp = this.pswp();

      if (!pswp || typeof pswp.next !== "function" || !pswp.currSlide?.content) {
        this.pauseSlideshow();
        return;
      }

      const { video } = this.getContent();

      if (video && !video.paused) {
        // Do nothing if a video is still playing.
      } else if (!this.$isRtl && this.models.length > this.index + 1) {
        // Show the next slide.
        this.slideshow.next = this.index + 1;
        pswp.next();
      } else if (this.$isRtl && this.index > 0) {
        // Reverse slideshow direction for right-to-left languages.
        this.slideshow.next = this.index - 1;
        pswp.prev();
      } else {
        // Pause slideshow if this is the end.
        this.pauseSlideshow();
      }
    },
    // Pauses the slideshow, if currently active.
    pauseSlideshow() {
      if (this.slideshow.active) {
        this.slideshow.active = false;
      }

      this.clearSlideshowInterval();

      this.slideshow.next = -1;

      this.showControls();
    },
    // Updates the collection cover, if a collection model exists.
    onSetCollectionCover() {
      if (!this.canManageAlbums || !(this.collection instanceof Collection)) {
        return;
      }

      this.pauseSlideshow();

      if (!this.model || !this.model.Hash) {
        this.log("viewer: could not update collection cover because the file hash is missing");
        return;
      }

      if (!this.collection || !this.collection?.UID) {
        this.log("viewer: could not update collection cover because the collection is not defined");
        return;
      }

      this.collection.setCover(this.model.Hash).then(() => {
        this.$notify.success(this.$gettext("Changes successfully saved"));
      });
    },
    onRemoveFromAlbum() {
      if (!this.canManageAlbums || !(this.collection instanceof Album)) {
        return;
      }

      this.pauseSlideshow();

      if (!this.model || !this.model?.UID) {
        this.log("viewer: could not remove picture from album because the model UID is not defined");
        return;
      }

      if (!this.collection || !this.collection?.UID) {
        this.log("viewer: could not remove picture from album because the album is not defined");
        return;
      }

      this.model.Removed = true;

      $api.delete(`albums/${this.collection.UID}/photos`, { data: { photos: [this.model.UID] } }).catch(() => {
        this.model.Removed = false;
      });
    },
    onArchive() {
      if (!this.canArchive) {
        return;
      }

      this.pauseSlideshow();

      if (!this.model || !this.model.UID) {
        this.log("viewer: could not move photo to archive because model UID is unknown");
        return;
      }

      this.model.Archived = true;

      return $api.post("batch/photos/archive", { photos: [this.model.UID] }).then(() => {
        this.$notify.success(this.$gettext("Archived"));
      });
    },
    onRestore() {
      if (!this.canArchive) {
        return;
      }

      this.pauseSlideshow();

      if (!this.model || !this.model.UID) {
        this.log("viewer: could remove photo from archive because model UID is unknown");
        return;
      }

      this.model.Archived = false;

      $api.post("batch/photos/restore", { photos: [this.model.UID] }).then(() => {
        this.$notify.success(this.$gettext("Restored"));
      });
    },
    // Downloads the original files of the current picture.
    onDownload() {
      if (!this.canDownload) {
        return;
      }

      this.pauseSlideshow();

      /*
        TODO: Once all the lightbox's core functionality has been restored, add a file size/type
              selection dialog so the user can choose which format and quality to download.
       */

      if (!this.model || !this.model.DownloadUrl) {
        this.log("viewer: no download url");
        return;
      }

      this.$notify.success(this.$gettext("Downloading…"));

      new Photo().find(this.model.UID).then((p) => p.downloadAll());
    },
    onEdit() {
      this.pauseLightbox();

      let index = 0;

      // remove duplicates
      let filtered = this.models?.filter(function (p, i, s) {
        return !(i > 0 && p.UID === s[i - 1].UID);
      });

      let selection = filtered.map((p, i) => {
        if (this.model.UID === p.UID) {
          index = i;
        }

        return p.UID;
      });

      let album = null;

      // Close lightbox and open edit dialog when closed.
      this.close().then(() => {
        this.$event.publish("dialog.edit", { selection, album, index });
      });
    },
    resize(force) {
      this.$nextTick(() => {
        if (this.visible && this.getLightboxElement() && !this.isBusy("resize")) {
          const pswp = this.pswp();
          if (pswp && pswp?.updateSize) {
            pswp.updateSize(force);
          }
        }
      });
    },
    toggleInfo() {
      if (!this.visible) {
        return;
      }

      if (this.info) {
        this.hideInfo();
      } else {
        this.showInfo();
      }
    },
    // Shows the lightbox sidebar, if hidden.
    showInfo() {
      if (!this.visible || this.info) {
        return;
      }

      this.info = true;

      localStorage.setItem("lightbox.info", `${this.info.toString()}`);

      // Resize and focus content element.
      this.$nextTick(() => {
        this.resize(true);
        this.focusContent();
      });
    },
    // Hides the lightbox sidebar, if visible.
    hideInfo() {
      if (!this.visible || !this.info) {
        return;
      }

      this.info = false;

      localStorage.setItem("lightbox.info", `${this.info.toString()}`);

      // Resize and focus content element.
      this.$nextTick(() => {
        this.resize(true);
        this.focusContent();
      });
    },
    toggleControls() {
      if (!this.visible) {
        return;
      }

      if (this.pswp() && this.pswp().element) {
        const el = this.pswp().element;
        if (el.classList.contains("pswp--ui-visible")) {
          this.hideControls();
        } else {
          this.showControls();
        }
      }
    },
    showControls() {
      if (!this.visible) {
        return;
      }

      this.showLightboxControls();
      this.startTimer();
    },
    showLightboxControls() {
      this.controlsShown = Date.now();
      this.showPswpControls();
    },
    showPswpControls() {
      const pswp = this.pswp();
      if (pswp && pswp.element?.classList?.add) {
        pswp.element.classList.add("pswp--ui-visible");
      }
    },
    hideControls() {
      if (!this.visible) {
        return;
      }

      this.hideLightboxControls();
    },
    hideLightboxControls() {
      if (this.menuVisible) {
        return;
      }

      this.controlsShown = 0;
      this.hidePswpControls();
    },
    hidePswpControls() {
      const pswp = this.pswp();
      if (pswp && pswp.element?.classList?.remove) {
        pswp.element.classList.remove("pswp--ui-visible");
      }
    },
    hideControlsWithDelay(delay) {
      if (!delay || delay < 1) {
        return;
      }

      this.clearIdleTimeout();
      this.idleTimer = window.setTimeout(() => {
        this.hideControls();
      }, delay);
    },
    controlsVisible() {
      return this.controlsShown !== 0;
    },
    onTouchStartOnce() {
      this.clearIdleTimeout();
      this.hasTouch = true;
    },
    onMouseMoveOnce() {
      this.showControls();
    },
    // Removes any touch and mouse event handlers.
    removeEventListeners() {
      document.removeEventListener("touchstart", this.touchStartListener, false);
      document.removeEventListener("mousemove", this.mouseMoveListener, false);
    },
    // Attaches touch and mouse event handlers to automatically hide controls.
    addEventListeners() {
      document.addEventListener("touchstart", this.touchStartListener, { once: true });
      document.addEventListener("mousemove", this.mouseMoveListener, { once: true });
    },
    startTimer() {
      if (this.hasTouch) {
        return;
      }

      this.hideControlsWithDelay(this.defaultControlHideDelay);

      document.addEventListener("mousemove", this.mouseMoveListener, { once: true });
    },
    clearTimeouts() {
      this.clearIdleTimeout();
    },
    // Clears the idle timer used to automatically hide the lightbox controls.
    clearIdleTimeout() {
      if (this.idleTimer) {
        window.clearTimeout(this.idleTimer);
        this.idleTimer = false;
      }
    },
    // Returns the viewport size without sidebar, if visible.
    getViewport() {
      const el = this.$refs?.content;

      if (el) {
        return {
          x: el.clientWidth,
          y: el.clientHeight,
        };
      } else {
        return {
          x: window.innerWidth,
          y: window.innerHeight,
        };
      }
    },
    getSlidePixels(model) {
      // Get viewport size without sidebar, if visible.
      const viewport = this.getViewport();

      // Subtract viewport padding to get estimated slide size if it is an image or vector graphic.
      if (model && (model.Type === media.Image || model.Type === media.Raw || model.Type === media.Vector)) {
        const padding = this.getPadding(viewport, { width: model.Width, height: model.Height });
        viewport.x = viewport.x - padding.left - padding.right;
        viewport.y = viewport.y - padding.top - padding.bottom;
      }

      // Calculate estimated slide size based on viewport size and device pixel ratio.
      return {
        width: viewport.x * window.devicePixelRatio,
        height: viewport.y * window.devicePixelRatio,
      };
    },
    // Calculates viewport padding based on screen and image size.
    getPadding(viewport, data) {
      let top = 0,
        bottom = 0,
        left = 0,
        right = 0;

      // No lightbox padding if content width or height is not specified.
      if (!viewport || !data?.width || !data?.height) {
        return { top, bottom, left, right };
      }

      // Add padding based on content and viewport size, except on small mobile screens.
      if (viewport.x > this.mobileBreakpoint) {
        // Large screens.
        if (data.width % viewport.x !== 0 && viewport.x > viewport.y) {
          left = 48;
          right = 48;
        }

        if (data.height % viewport.y === 0) {
          top = 48;
          bottom = 48;
          left = 48;
          right = 48;
        } else if (data.height > data.width) {
          top = 48;
          bottom = 48;
        } else {
          top = 72;
          bottom = 64;
        }
      }

      return { top, bottom, left, right };
    },
    // Updates the thumbnail when the zoom level changes and a different resolution may be required.
    onImageSizeChange() {
      if (this.isBusy("change image size")) {
        return;
      }

      const { slide, content, video, data } = this.getContent();

      if (!slide) {
        return;
      }

      // Continue only if the content is an <img> element and not e.g. a <video>.
      if (video || data?.type === "html" || data?.loading) {
        return;
      } else if (!content || !content.element || !(content.element instanceof HTMLImageElement)) {
        return;
      }

      // Get current zoom level and image model.
      const zoomLevel = slide.currZoomLevel;
      const model = data.model;

      // Do not proceed if the model is missing or incomplete.
      if (!model || !model.Thumbs) {
        return;
      }

      // Do not proceed unless the image is zoomed to near its intrinsic (natural) size.
      if (zoomLevel < 0.95) {
        return;
      }

      // Calculate slide width and height in real pixels based on zoom level and pixel ratio.
      const slideWidth = Math.ceil(slide.width * zoomLevel * window.devicePixelRatio);
      const slideHeight = Math.ceil(slide.height * zoomLevel * window.devicePixelRatio);

      // Find thumbnail size that best matches the current slide size and zoom level.
      const thumb = this.$util.thumb(model.Thumbs, slideWidth, slideHeight);

      // Do not change image if no matching thumbnail size was found or is available.
      if (!thumb || !thumb.src || !thumb.w || !thumb.h) {
        return;
      }

      // Get the thumbnail URL of the currently displayed image.
      const currentSrc = data.src;

      // Do not proceed if the thumbnail URL remains the same.
      if (currentSrc === thumb.src) {
        return;
      }

      // Create HTMLImageElement to load thumbnail image in the matching size.
      try {
        const image = new Image();

        // Decode the image synchronously for atomic presentation with other content:
        // https://developer.mozilla.org/en-US/docs/Web/API/HTMLImageElement/decoding
        image.decoding = "sync";

        // Tell the browser to load the new image as quickly as possible:
        // https://developer.mozilla.org/en-US/docs/Web/API/HTMLImageElement/loading
        image.loading = "eager";

        // Flag the new image as loading in the content data.
        data.loading = true;

        // Attach an onload event handler to swap the thumbnail when the new image is loaded.
        const onImageLoad = (ev) => {
          if (!ev || !ev.target) {
            return;
          }

          // Remove loading flag.
          data.loading = false;

          if (this.trace) {
            this.log(`image.${ev.type}`, { ev, target: ev.target });
          }

          // Abort if image URL is empty or the current slide is undefined.
          if (!content || !ev.target.currentSrc || !ev.target.naturalHeight || !ev.target.naturalWidth) {
            if (this.debug) {
              this.log(`failed to replace thumbnail with ${thumb.src}`, { element: content.element, image: ev.target });
            }
            data.loading = false;
            return;
          }

          if (content.element.src === ev.target.currentSrc) {
            this.log(`old and new thumbnail are the same ${ev.target.currentSrc}`);
            data.loading = false;
            return;
          }

          if (this.debug) {
            this.log(`loaded thumbnail ${thumb.size} from ${ev.target.currentSrc}`);
          }

          // Update the slide's HTMLImageElement to use the new thumbnail image.
          content.element.src = ev.target.currentSrc;
          content.element.width = ev.target.width;
          content.element.height = ev.target.height;

          // Update PhotoSwipe's slide data.
          data.src = thumb.src;
          data.width = thumb.w;
          data.height = thumb.h;
          data.loading = false;
        };

        image.addEventListener("load", onImageLoad, { once: true });

        // Set thumbnail src to load the new image.
        image.src = thumb.src;
      } catch (err) {
        this.log(`failed to load image size ${thumb.size}`, { err });
        data.loading = false;
      }
    },
  },
};
</script>
