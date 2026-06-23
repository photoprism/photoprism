<template>
  <div
    class="p-pdf-viewer"
    :class="{ 'is-loading': loading, 'is-error': !!errorMessage, 'is-thumbs-hidden': !thumbsVisible }"
    @wheel.stop
    @pointerdown.stop
    @touchstart.stop
    @touchmove.stop
  >
    <div v-if="errorMessage" class="p-pdf-viewer__error">
      <v-icon icon="mdi-file-alert-outline" size="48"></v-icon>
      <div class="text-body-1">{{ errorMessage }}</div>
    </div>
    <template v-else>
      <div ref="thumbs" class="p-pdf-viewer__thumbs">
        <button
          v-for="n in pageCount"
          :key="'thumb-' + n"
          type="button"
          class="p-pdf-viewer__thumb"
          :class="{ 'is-active': n === currentPage }"
          :aria-label="$gettext('Go to page %{n}', { n })"
          @click="goToPage(n)"
        >
          <canvas ref="thumb" :data-page="n"></canvas>
          <span class="p-pdf-viewer__thumb-label">{{ n }}</span>
        </button>
      </div>
      <div
        ref="scroll"
        class="p-pdf-viewer__pages"
        tabindex="0"
        @scroll="onScroll"
        @keydown.left.exact.prevent="$emit('media-prev')"
        @keydown.right.exact.prevent="$emit('media-next')"
      >
        <div v-for="n in pageCount" :key="'page-' + n" ref="page" class="p-pdf-viewer__page" :data-page="n">
          <canvas ref="canvas"></canvas>
        </div>
        <div v-if="loading" class="p-pdf-viewer__spinner">
          <v-progress-circular indeterminate color="primary"></v-progress-circular>
        </div>
      </div>
      <div class="p-pdf-viewer__controls">
        <v-btn
          :icon="thumbsVisible ? 'mdi-view-grid' : 'mdi-view-grid-outline'"
          variant="text"
          density="comfortable"
          :aria-label="$gettext('Toggle thumbnails')"
          @click="toggleThumbs"
        ></v-btn>
        <span class="p-pdf-viewer__sep"></span>
        <v-btn
          icon="mdi-chevron-left"
          variant="text"
          density="comfortable"
          :disabled="currentPage <= 1"
          :aria-label="$gettext('Previous page')"
          @click="prevPage"
        ></v-btn>
        <div class="p-pdf-viewer__pageinfo">
          <input
            class="p-pdf-viewer__pageinput"
            type="text"
            inputmode="numeric"
            :value="pageInput"
            :aria-label="$gettext('Go to page')"
            @input="pageInput = $event.target.value"
            @keyup.enter="submitPageInput"
            @blur="submitPageInput"
          />
          <span class="p-pdf-viewer__pagecount">/ {{ pageCount }}</span>
        </div>
        <v-btn
          icon="mdi-chevron-right"
          variant="text"
          density="comfortable"
          :disabled="currentPage >= pageCount"
          :aria-label="$gettext('Next page')"
          @click="nextPage"
        ></v-btn>
        <span class="p-pdf-viewer__sep"></span>
        <v-btn
          icon="mdi-magnify-minus-outline"
          variant="text"
          density="comfortable"
          :disabled="scale <= minScale"
          :aria-label="$gettext('Zoom out')"
          @click="zoomOut"
        ></v-btn>
        <v-btn
          icon="mdi-magnify-plus-outline"
          variant="text"
          density="comfortable"
          :disabled="scale >= maxScale"
          :aria-label="$gettext('Zoom in')"
          @click="zoomIn"
        ></v-btn>
      </div>
    </template>
  </div>
</template>

<script>
import { markRaw } from "vue";
import { loadPdfDocument, renderPdfPage, destroyPdfDocument, isRenderCancelled } from "common/pdf";

// Scale used for the small thumbnail-strip previews; independent of the main zoom.
const ThumbScale = 0.2;

export default {
  name: "PPdfViewer",
  props: {
    // src is the inline-PDF URL (see $util.pdfUrl).
    src: {
      type: String,
      default: "",
    },
    // active gates loading and rendering so preloaded-but-hidden slides stay idle.
    active: {
      type: Boolean,
      default: true,
    },
    // pages is an optional page-count hint; the loaded document is authoritative.
    pages: {
      type: Number,
      default: 0,
    },
  },
  emits: ["loaded", "page-changed", "error", "media-prev", "media-next", "thumbs-visible"],
  data() {
    return {
      pdf: null,
      pageCount: this.pages || 0,
      currentPage: 1,
      pageInput: "1",
      scale: 1.0,
      minScale: 0.5,
      maxScale: 4.0,
      scaleStep: 0.25,
      loading: false,
      errorMessage: "",
      thumbsVisible: true,
    };
  },
  watch: {
    active(value) {
      if (value) {
        this.open();
      } else {
        this.teardown();
      }
    },
    src() {
      this.reopen();
    },
    scale() {
      this.rerenderVisible();
    },
  },
  created() {
    // Non-reactive bookkeeping for observers, in-flight render handles, and the
    // per-page visibility ratios that drive the current-page indicator.
    this.pageObserver = null;
    this.thumbObserver = null;
    this.renderTasks = {};
    this.thumbTasks = {};
    this.renderedPages = {};
    this.renderedThumbs = {};
    this.intersecting = {};
    this.scrollPending = false;
    // Target page of an in-flight programmatic alignment (0 when idle); set so
    // updateCurrentPage defers to the requested page until the jump converges.
    this.scrollingTo = 0;
    this.destroyed = false;
    // Hide the thumbnail strip by default on small screens, where it would take
    // most of the width; it can be toggled back on from the controls.
    if (this.$vuetify && this.$vuetify.display && this.$vuetify.display.smAndDown) {
      this.thumbsVisible = false;
    }
  },
  mounted() {
    // Publish the initial strip visibility so the lightbox can place the prev
    // navigation arrow to the right of the thumbnail strip (or at the far edge
    // when the strip is hidden).
    this.$emit("thumbs-visible", this.thumbsVisible);

    if (this.active && this.src) {
      this.open();
    }
  },
  beforeUnmount() {
    this.destroyed = true;
    this.teardown();
  },
  methods: {
    // open loads the document, sizes the page placeholders, and starts observing
    // pages and thumbnails for lazy, progressive rendering.
    async open() {
      if (!this.active || !this.src || this.pdf || this.loading) {
        return;
      }

      this.loading = true;
      this.errorMessage = "";

      try {
        const { pdf, pageCount } = await loadPdfDocument(this.src);

        if (this.destroyed || !this.active) {
          destroyPdfDocument(pdf);
          return;
        }

        // Keep the pdf.js document non-reactive; a Vue reactive Proxy breaks its
        // private class fields (getPage would throw "object is not the right class").
        this.pdf = markRaw(pdf);
        this.pageCount = pageCount;
        this.$emit("loaded", pageCount);
        await this.$nextTick();
        this.observeAll();
        // Focus the scroll area so Up/Down/PageDown scroll the document and
        // Left/Right are captured here for media navigation.
        if (this.$refs.scroll && typeof this.$refs.scroll.focus === "function") {
          this.$refs.scroll.focus({ preventScroll: true });
        }
      } catch (e) {
        this.errorMessage = this.$gettext("Failed to load the PDF document.");
        this.$emit("error", e);
      } finally {
        this.loading = false;
      }
    },
    // reopen tears down and reloads after the source changes.
    reopen() {
      this.teardown();

      if (this.active && this.src) {
        this.$nextTick(() => this.open());
      }
    },
    // teardown cancels in-flight renders, disconnects observers, and releases the
    // document so an inactive or unmounted slide holds no resources.
    teardown() {
      Object.values(this.renderTasks).forEach((t) => t && t.cancel());
      Object.values(this.thumbTasks).forEach((t) => t && t.cancel());
      this.renderTasks = {};
      this.thumbTasks = {};
      this.renderedPages = {};
      this.renderedThumbs = {};
      this.intersecting = {};

      if (this.pageObserver) {
        this.pageObserver.disconnect();
        this.pageObserver = null;
      }
      if (this.thumbObserver) {
        this.thumbObserver.disconnect();
        this.thumbObserver = null;
      }

      if (this.pdf) {
        destroyPdfDocument(this.pdf);
        this.pdf = null;
      }
    },
    // observeAll attaches IntersectionObservers to the page and thumbnail elements.
    observeAll() {
      this.pageObserver = new IntersectionObserver((entries) => this.onPagesIntersect(entries), {
        root: this.$refs.scroll,
        rootMargin: "100% 0px",
        threshold: 0,
      });
      (this.$refs.page || []).forEach((el) => this.pageObserver.observe(el));

      this.thumbObserver = new IntersectionObserver((entries) => this.onThumbsIntersect(entries), {
        root: this.$refs.thumbs,
        rootMargin: "200px",
        threshold: 0,
      });
      (this.$refs.thumb || []).forEach((el) => this.thumbObserver.observe(el));
    },
    // onPagesIntersect renders pages entering the viewport band and updates the
    // current page to the one with the largest visible area (the single tracker
    // that keeps the indicator and thumbnail highlight in sync with scrolling).
    onPagesIntersect(entries) {
      for (const entry of entries) {
        const n = Number(entry.target.dataset.page);
        this.intersecting[n] = entry.isIntersecting;

        if (entry.isIntersecting) {
          this.renderPage(n);
        }
      }

      this.updateCurrentPage();
    },
    // updateCurrentPage sets the current page to the one occupying the most of
    // the actual viewport, computed from geometry. The observer's
    // intersectionRatio cannot be used here: the render-ahead rootMargin makes
    // several pages report a full ratio at once, so the lowest-numbered one
    // would always win and the indicator would lag a page behind. Only the pages
    // the observer currently reports as intersecting can be in view, so the scan
    // stays O(visible) instead of measuring every page on each scroll frame.
    updateCurrentPage() {
      const scroll = this.$refs.scroll;

      if (!scroll || !this.pageCount || this.scrollingTo) {
        return;
      }

      const view = scroll.getBoundingClientRect();
      const pages = this.$refs.page || [];
      let best = 0;
      let bestVisible = 0;

      for (const key in this.intersecting) {
        if (!this.intersecting[key]) {
          continue;
        }

        const el = pages[Number(key) - 1];

        if (!el) {
          continue;
        }

        const r = el.getBoundingClientRect();
        const visible = Math.min(r.bottom, view.bottom) - Math.max(r.top, view.top);

        if (visible > bestVisible) {
          bestVisible = visible;
          best = Number(key);
        }
      }

      if (best > 0) {
        this.setCurrentPage(best);
      }
    },
    // onScroll keeps the current page in sync while scrolling, throttled to one
    // computation per animation frame.
    onScroll() {
      if (this.scrollPending) {
        return;
      }

      this.scrollPending = true;
      requestAnimationFrame(() => {
        this.scrollPending = false;
        this.updateCurrentPage();
      });
    },
    // onThumbsIntersect lazily renders thumbnails as they scroll into view.
    onThumbsIntersect(entries) {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          this.renderThumb(Number(entry.target.dataset.page));
        }
      }
    },
    // renderPage renders page n onto its canvas, ignoring benign cancellations.
    async renderPage(n) {
      if (!this.pdf || this.renderTasks[n] || this.renderedPages[n]) {
        return;
      }

      const canvas = (this.$refs.canvas || [])[n - 1];

      if (!canvas) {
        return;
      }

      const handle = renderPdfPage(this.pdf, n, canvas, this.scale);
      this.renderTasks[n] = handle;

      try {
        await handle.promise;
        this.renderedPages[n] = true;
      } catch (e) {
        if (!isRenderCancelled(e)) {
          console.error("pdf: failed to render page", n, e);
        }
      } finally {
        delete this.renderTasks[n];
      }
    },
    // renderThumb renders a small preview for page n in the thumbnail strip.
    async renderThumb(n) {
      if (!this.pdf || this.thumbTasks[n] || this.renderedThumbs[n]) {
        return;
      }

      const canvas = (this.$refs.thumb || [])[n - 1];

      if (!canvas) {
        return;
      }

      const handle = renderPdfPage(this.pdf, n, canvas, ThumbScale);
      this.thumbTasks[n] = handle;

      try {
        await handle.promise;
        this.renderedThumbs[n] = true;
      } catch (e) {
        if (!isRenderCancelled(e)) {
          console.error("pdf: failed to render thumbnail", n, e);
        }
      } finally {
        delete this.thumbTasks[n];
      }
    },
    // rerenderVisible re-renders pages at the active scale after a zoom change.
    // Zoom applies to every page, so all rendered pages are invalidated — the
    // currently visible ones re-render now, and off-screen ones re-render when
    // scrolled into view. Re-rendering only the visible pages would leave the
    // rest sized at the previous scale, so pages would not scale uniformly.
    rerenderVisible() {
      Object.values(this.renderTasks).forEach((t) => t && t.cancel());
      this.renderTasks = {};
      this.renderedPages = {};

      for (const key in this.intersecting) {
        if (this.intersecting[key]) {
          this.renderPage(Number(key));
        }
      }
    },
    // setCurrentPage updates the indicator state when the page in view changes.
    setCurrentPage(n) {
      if (n === this.currentPage) {
        return;
      }

      this.currentPage = n;
      this.pageInput = String(n);
      this.$emit("page-changed", n);
      this.scrollActiveThumbIntoView();
    },
    // goToPage scrolls to and renders the target page, clamped to valid bounds.
    // Adjacent hops scroll smoothly; longer jumps re-align via alignToPage,
    // because lazy rendering grows pages from their placeholder height as they
    // enter the viewport and shifts the target's offset — a single scroll would
    // land on a nearby page on large documents.
    goToPage(n) {
      const target = Math.min(Math.max(1, Math.floor(n) || 1), this.pageCount || 1);
      const changed = target !== this.currentPage;
      const stepwise = Math.abs(target - this.currentPage) <= 1;

      this.currentPage = target;
      this.pageInput = String(target);
      this.renderPage(target);
      this.scrollActiveThumbIntoView();

      const el = (this.$refs.page || [])[target - 1];

      if (stepwise && el && typeof el.scrollIntoView === "function") {
        // Cancel any in-flight alignment so it does not fight the smooth scroll.
        this.scrollingTo = 0;
        el.scrollIntoView({ block: "start", behavior: "smooth" });
      } else if (el) {
        this.alignToPage(target);
      }

      if (changed) {
        this.$emit("page-changed", target);
      }
    },
    // alignToPage scrolls the target page to the top, correcting until it
    // settles. Pages render lazily and grow from their placeholder height as
    // they enter the viewport, which shifts the target's offset after the first
    // scroll; on large documents a single jump lands on a nearby page. The
    // scrollingTo guard makes updateCurrentPage defer to the requested page
    // until alignment converges or the attempt budget runs out.
    alignToPage(target) {
      const scroll = this.$refs.scroll;
      const el = (this.$refs.page || [])[target - 1];

      if (!scroll || !el) {
        return;
      }

      this.scrollingTo = target;
      let attempts = 0;

      const align = () => {
        if (this.destroyed || this.scrollingTo !== target) {
          return;
        }

        const delta = el.getBoundingClientRect().top - scroll.getBoundingClientRect().top;

        if (Math.abs(delta) <= 1 || attempts >= 16) {
          this.scrollingTo = 0;
          this.updateCurrentPage();
          return;
        }

        attempts += 1;
        scroll.scrollTop += delta;
        requestAnimationFrame(align);
      };

      align();
    },
    prevPage() {
      this.goToPage(this.currentPage - 1);
    },
    nextPage() {
      this.goToPage(this.currentPage + 1);
    },
    // submitPageInput jumps to the page typed into the jump-to-page field.
    submitPageInput() {
      const n = parseInt(this.pageInput, 10);

      if (isNaN(n)) {
        this.pageInput = String(this.currentPage);
        return;
      }

      this.goToPage(n);
    },
    zoomIn() {
      this.setScale(this.scale + this.scaleStep);
    },
    zoomOut() {
      this.setScale(this.scale - this.scaleStep);
    },
    // setScale clamps and applies a new zoom level.
    setScale(value) {
      const clamped = Math.min(this.maxScale, Math.max(this.minScale, Math.round(value * 100) / 100));

      if (clamped !== this.scale) {
        this.scale = clamped;
      }
    },
    // toggleThumbs shows or hides the page-thumbnail strip (closed by default on
    // small screens). Thumbnails render lazily, so the strip catches up when shown.
    toggleThumbs() {
      this.thumbsVisible = !this.thumbsVisible;
      this.$emit("thumbs-visible", this.thumbsVisible);
    },
    // scrollActiveThumbIntoView keeps the highlighted thumbnail visible in the strip.
    scrollActiveThumbIntoView() {
      const el = (this.$refs.thumb || [])[this.currentPage - 1];

      if (el && typeof el.scrollIntoView === "function") {
        el.scrollIntoView({ block: "nearest", inline: "nearest" });
      }
    },
  },
};
</script>

<style src="../css/pdf-viewer.css"></style>
