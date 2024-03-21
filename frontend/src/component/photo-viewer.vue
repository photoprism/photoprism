<template>
  <div id="photo-viewer" class="p-viewer pswp" tabindex="-1" role="dialog" aria-hidden="true">
    <div class="pswp__bg"></div>
    <div class="pswp__scroll-wrap">
      <div class="pswp__container" :class="{ slideshow: slideshow.active }">
        <div class="pswp__item"></div>
        <div class="pswp__item"></div>
        <div class="pswp__item"></div>
      </div>

      <div class="pswp__ui pswp__ui--hidden">
        <div class="pswp__top-bar">
          <div class="pswp__taken hidden-xs-only">
            {{ formatDate(item.TakenAtLocal) }}
          </div>

          <div class="pswp__counter"></div>

          <button class="pswp__button pswp__button--close action-close" :title="$gettext('Close')"></button>

          <button v-if="canDownload" class="pswp__button action-download" style="background: none" :title="$gettext('Download')" @click.exact="onDownload">
            <v-icon size="16" color="white">get_app</v-icon>
          </button>

          <button v-if="canEdit" class="pswp__button action-edit hidden-shared-only" style="background: none" :title="$gettext('Edit')" @click.exact="onEdit">
            <v-icon size="16" color="white">edit</v-icon>
          </button>

          <button class="pswp__button action-select" style="background: none" :title="$gettext('Select')" @click.exact="onSelect">
            <v-icon v-if="selection.length && $clipboard.has(item)" size="16" color="white">check_circle</v-icon>
            <v-icon v-else size="16" color="white">radio_button_off</v-icon>
          </button>

          <button v-if="canLike" class="pswp__button action-like hidden-shared-only" style="background: none" :title="$gettext('Like')" @click.exact="onLike">
            <v-icon v-if="item.Favorite" size="16" color="white">favorite</v-icon>
            <v-icon v-else size="16" color="white">favorite_border</v-icon>
          </button>

          <button class="pswp__button pswp__button--fs action-toggle-fullscreen" :title="$gettext('Fullscreen')"></button>

          <button class="pswp__button pswp__button--zoom action-zoom" :title="$gettext('Zoom in/out')"></button>

          <button class="pswp__button action-slideshow" style="background: none" :title="$gettext('Start/Stop Slideshow')" @click.exact="onSlideshow">
            <v-icon v-show="!interval" size="18" color="white">play_arrow</v-icon>
            <v-icon v-show="interval" size="16" color="white">pause</v-icon>
          </button>

          <div class="pswp__preloader">
            <div class="pswp__preloader__icn">
              <div class="pswp__preloader__cut">
                <div class="pswp__preloader__donut"></div>
              </div>
            </div>
          </div>
        </div>

        <div class="pswp__share-modal pswp__share-modal--hidden pswp__single-tap">
          <div class="pswp__share-tooltip"></div>
        </div>

        <button class="pswp__button pswp__button--arrow--left action-previous" title="Previous (arrow left)">
</button>

        <button class="pswp__button pswp__button--arrow--right action-next" title="Next (arrow right)">
</button>

        <div class="pswp__caption" @click="onPlay">
          <div class="pswp__caption__center"></div>
        </div>
      </div>
    </div>
    <div v-if="player.show" class="video-viewer" @click.stop.prevent="closePlayer" @keydown.esc.stop.prevent="closePlayer">
      <p-video-player ref="player" :source="player.source" :poster="player.poster" :height="player.height" :width="player.width" :autoplay="player.autoplay" :loop="player.loop" @close="closePlayer">
</p-video-player>
    </div>
  </div>
</template>

<script>
import "photoswipe/dist/photoswipe.css";
import "photoswipe/dist/default-skin/default-skin.css";
import Event from "pubsub-js";
import Thumb from "model/thumb";
import { Photo, DATE_FULL } from "model/photo";
import Notify from "common/notify";
import { DateTime } from "luxon";

export default {
  name: "PPhotoViewer",
  data() {
    return {
      canEdit: this.$config.allow("photos", "update") && this.$config.feature("edit"),
      canLike: this.$config.allow("photos", "manage") && this.$config.feature("favorites"),
      canDownload: this.$config.allow("photos", "download") && this.$config.feature("download"),
      selection: this.$clipboard.selection,
      config: this.$config.values,
      item: new Thumb(),
      subscriptions: [],
      interval: false,
      slideshow: {
        active: false,
        next: 0,
      },
      player: {
        show: false,
        loop: false,
        autoplay: true,
        source: "",
        poster: "",
        width: 640,
        height: 480,
      },
    };
  },
  created() {
    this.subscriptions["viewer.change"] = Event.subscribe("viewer.change", this.onChange);
    this.subscriptions["viewer.pause"] = Event.subscribe("viewer.pause", this.onPause);
    this.subscriptions["viewer.show"] = Event.subscribe("viewer.show", this.onShow);
    this.subscriptions["viewer.hide"] = Event.subscribe("viewer.hide", this.onHide);
  },
  destroyed() {
    this.onPause();

    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    formatDate(s) {
      if (!s || !s.length) {
        return s;
      }

      const l = s.length;

      if (l !== 20 || s[l - 1] !== "Z") {
        return s;
      }

      return DateTime.fromISO(s, { zone: "UTC" }).toLocaleString(DATE_FULL);
    },
    onShow() {
      this.$scrollbar.hide();
    },
    onHide() {
      this.closePlayer();
      this.onPause();
      this.$scrollbar.show();
    },
    onChange(ev, data) {
      const psp = this.$viewer.gallery;

      if (psp && this.slideshow.next !== psp.getCurrentIndex()) {
        this.onPause();
      }

      if (data.item && this.item && this.item.UID !== data.item.UID) {
        this.closePlayer();
      }

      this.item = data.item;
    },
    onLike() {
      this.item.toggleLike();
    },
    onSelect() {
      this.$clipboard.toggle(this.item);
    },
    onPlay() {
      if (this.item && this.item.Playable) {
        new Photo().find(this.item.UID).then((video) => this.openPlayer(video));
      }
    },
    openPlayer(video) {
      if (!video) {
        this.$notify.error(this.$gettext("No video selected"));
        return;
      }

      const params = video.videoParams();

      if (params.error) {
        this.$notify.error(params.error);
        return;
      }

      // Set video parameters.
      this.player.loop = params.loop;
      this.player.width = params.width;
      this.player.height = params.height;
      this.player.poster = params.poster;
      this.player.source = params.uri;

      // Play video.
      this.player.show = true;
    },
    closePlayer() {
      if (this.$refs.player) {
        this.$refs.player.stop();
      }

      this.player.show = false;
    },
    onPause() {
      this.slideshow.active = false;

      if (this.interval) {
        clearInterval(this.interval);
        this.interval = false;
      }
    },
    onSlideshow() {
      if (this.interval) {
        this.onPause();
        return;
      }

      this.slideshow.active = true;

      const self = this;
      const psp = this.$viewer.gallery;

      self.interval = setInterval(() => {
        if (psp && typeof psp.next === "function") {
          psp.next();
          this.slideshow.next = psp.getCurrentIndex();
        } else {
          this.onPause();
        }
      }, 5000);
    },
    onDownload() {
      this.onPause();

      if (!this.item || !this.item.DownloadUrl) {
        console.warn("photo viewer: no download url");
        return;
      }

      Notify.success(this.$gettext("Downloading…"));

      new Photo().find(this.item.UID).then((p) => p.downloadAll());
    },
    onEdit() {
      this.onPause();

      const g = this.$viewer.gallery; // Gallery
      let index = 0;

      // remove duplicates
      let filtered = g.items.filter(function (p, i, s) {
        return !(i > 0 && p.UID === s[i - 1].UID);
      });

      let selection = filtered.map((p, i) => {
        if (g.currItem.UID === p.UID) {
          index = i;
        }

        return p.UID;
      });

      let album = null;

      g.close(); // Close Gallery

      Event.publish("dialog.edit", { selection, album, index }); // Open Edit Dialog
    },
  },
};
</script>
