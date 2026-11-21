<template>
  <section ref="header" class="p-album-header" :style="[tone, { '--album-header-min-height': `${minimumHeight}px` }]" :aria-label="album.Title">
    <div v-if="imageUrl" class="album-header-backdrop" :style="{ backgroundImage: `url(${imageUrl})` }"></div>
    <img v-if="imageUrl" ref="image" :key="imageUrl" class="album-header-image" :src="imageUrl" :style="imageStyle" alt="" decoding="async" @load="onImageLoad" @error="onImageError" />
    <div ref="copy" class="album-header-copy">
      <h1><bdi>{{ $util.balanceBidi(album.Title) }}</bdi></h1>
      <div v-if="album.Description" class="album-header-description" :style="descriptionHeight === null ? {} : { height: `${descriptionHeight}px` }">
        <p ref="description">{{ album.Description }}</p>
      </div>
      <button v-if="overflowing || expanded" type="button" class="album-header-more" :aria-expanded="expanded" @click="toggleDescription">
        {{ expanded ? $gettext("Show Less") : $gettext("Show More") }}
        <v-icon size="16">{{ expanded ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
      </button>
    </div>
  </section>
</template>

<script>
import $api from "common/api";
import { albumCoverFrame, albumCoverTone, albumCollapseProgress } from "common/album-view";

export default {
  name: "PAlbumHeader",
  props: {
    album: { type: Object, required: true },
    revision: { type: Number, default: 0 },
  },
  emits: ["appearance", "collapsed", "collapse-progress"],
  // data initializes independent cover loading and layout state.
  data() {
    return { cover: null, imageUrl: "", frame: {}, tone: albumCoverTone([]), expanded: false, overflowing: false, request: 0, minimumHeight: 0, analyzedHash: "", descriptionHeight: null };
  },
  computed: {
    // coverKey changes whenever the selected album or its published cover changes.
    coverKey() {
      return `${this.album.UID}:${this.album.Thumb}:${this.revision}`;
    },
    // imageStyle positions the foreground independently of the blurred backdrop.
    imageStyle() {
      return Object.fromEntries(Object.entries(this.frame).map(([key, value]) => [key, `${value}px`]));
    },
  },
  watch: {
    // coverKey reloads framing information without blocking the photo query.
    coverKey() {
      this.loadCover();
    },
    // album.Description recalculates truncation after metadata edits.
    "album.Description"() {
      this.expanded = false;
      this.$nextTick(this.measure);
    },
  },
  // mounted observes header visibility and responsive layout changes.
  mounted() {
    this.resizeObserver = new ResizeObserver(this.queueMeasure);
    this.resizeObserver.observe(this.$refs.header);
    this.resizeObserver.observe(this.$refs.copy);
    window.addEventListener("scroll", this.queueScroll, { passive: true, capture: true });
    this.updateCollapse();
    this.loadCover();
  },
  // beforeUnmount disconnects observers and invalidates pending responses.
  beforeUnmount() {
    this.request++;
    this.resizeObserver?.disconnect();
    window.removeEventListener("scroll", this.queueScroll, { capture: true });
    cancelAnimationFrame(this.scrollFrame);
    cancelAnimationFrame(this.layoutFrame);
  },
  methods: {
    // queueMeasure batches layout changes while the description expands.
    queueMeasure() {
      if (!this.layoutFrame) {
        this.layoutFrame = requestAnimationFrame(() => {
          this.layoutFrame = 0;
          this.measure();
        });
      }
    },
    // queueScroll batches scroll geometry reads into a single animation frame.
    queueScroll() {
      if (!this.scrollFrame) {
        this.scrollFrame = requestAnimationFrame(() => {
          this.scrollFrame = 0;
          this.updateCollapse();
        });
      }
    },
    // updateCollapse reports a continuous transition into the normal toolbar.
    updateCollapse() {
      const header = this.$refs.header;
      if (!header) {
        return;
      }
      const toolbar = header.parentElement?.querySelector(".p-album-toolbar .v-toolbar");
      const progress = albumCollapseProgress(header.getBoundingClientRect().bottom, toolbar?.getBoundingClientRect().bottom || 48);
      this.$emit("collapse-progress", progress);
      this.$emit("collapsed", progress >= 1);
    },
    // loadCover resolves the current cover while ignoring responses for earlier revisions.
    loadCover() {
      const request = ++this.request;
      if (!this.album.UID) {
        return;
      }
      return $api.get(`albums/${this.album.UID}/header`).then(({ data }) => {
        if (request !== this.request) {
          return;
        }
        this.cover = data;
        if (!data.Hash) {
          this.onImageError();
          return;
        }
        const width = this.$refs.header?.clientWidth || window.innerWidth;
        const target = width * Math.min(window.devicePixelRatio || 1, 2);
        const size = [720, 1280, 1600, 1920, 2048, 2560, 3840].find((value) => value >= target) || 3840;
        this.imageUrl = `${this.$config.contentUri}/t/${data.Hash}/${this.$config.previewToken}/fit_${size}`;
        this.$nextTick(this.measure);
      }).catch(() => {
        if (request === this.request) {
          this.onImageError();
        }
      });
    },
    // measure fits the cover and checks whether the description needs expansion.
    measure() {
      const header = this.$refs.header;
      const image = this.$refs.image;
      if (!header) {
        return;
      }
      this.updateCollapse();
      const toolbarHeight = header.parentElement?.querySelector(".p-album-toolbar .v-toolbar")?.clientHeight || 48;
      const captionHeight = this.$refs.copy?.clientHeight || 0;
      this.minimumHeight = toolbarHeight + captionHeight + 120;
      this.frame = albumCoverFrame(header.clientWidth, header.clientHeight, image?.naturalWidth || this.cover?.Width, image?.naturalHeight || this.cover?.Height, this.cover?.Faces, { top: toolbarHeight, bottom: captionHeight });
      const description = this.$refs.description;
      if (description) {
        const lineHeight = parseFloat(getComputedStyle(description).lineHeight) || 24;
        const shortenedHeight = Math.min(description.scrollHeight, lineHeight * 2);
        this.overflowing = description.scrollHeight > shortenedHeight + 1;
        this.descriptionHeight = this.expanded ? description.scrollHeight : shortenedHeight;
      }
      if (image?.complete && image.naturalWidth) {
        this.updateTone();
      }
    },
    // onImageLoad measures a decoded image before choosing its text color.
    onImageLoad() {
      this.measure();
    },
    // updateTone chooses one stable black-or-white palette for the current cover.
    updateTone() {
      const image = this.$refs.image;
      const hash = this.cover?.Hash;
      if (!hash || hash === this.analyzedHash || !image?.naturalWidth) {
        return;
      }
      try {
        const canvas = document.createElement("canvas");
        canvas.width = 32;
        canvas.height = 32;
        const context = canvas.getContext("2d", { willReadFrequently: true });
        context.drawImage(image, 0, 0, 32, 32);
        const palette = albumCoverTone(context.getImageData(0, 0, 32, 32).data);
        this.tone = { ...palette, "--album-toolbar-ink": palette["--album-cover-ink"] };
      } catch {
        this.tone = albumCoverTone([]);
      }
      this.analyzedHash = hash;
      this.$emit("appearance", this.tone);
    },
    // onImageError supplies a readable themed placeholder when no cover can be loaded.
    onImageError() {
      this.analyzedHash = "";
      this.imageUrl = "";
      this.tone = albumCoverTone([]);
      this.$emit("appearance", this.tone);
    },
    // toggleDescription expands or shortens the caption and measures the resulting frame.
    toggleDescription() {
      this.expanded = !this.expanded;
      this.$nextTick(this.measure);
    },
  },
};
</script>
