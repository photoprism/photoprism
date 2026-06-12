<template>
  <div
    class="p-pdf-viewer"
    :class="{ 'is-loading': loading, 'is-error': !!errorMessage }"
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
  emits: ["loaded", "page-changed", "error", "media-prev", "media-next"],
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
    this.visibleRatios = {};
    this.destroyed = false;
  },
  mounted() {
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
      this.visibleRatios = {};

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
        threshold: [0, 0.25, 0.5, 0.75, 1],
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
        this.visibleRatios[n] = entry.isIntersecting ? entry.intersectionRatio : 0;

        if (entry.isIntersecting) {
          this.renderPage(n);
        }
      }

      let best = 0;
      let bestRatio = 0;

      for (const key in this.visibleRatios) {
        if (this.visibleRatios[key] > bestRatio) {
          bestRatio = this.visibleRatios[key];
          best = Number(key);
        }
      }

      if (best > 0) {
        this.setCurrentPage(best);
      }
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
    // rerenderVisible re-renders the currently visible pages at the active scale,
    // cancelling any in-flight renders first (used after a zoom change).
    rerenderVisible() {
      for (const key in this.visibleRatios) {
        const n = Number(key);

        if (this.visibleRatios[key] > 0) {
          if (this.renderTasks[n]) {
            this.renderTasks[n].cancel();
            delete this.renderTasks[n];
          }

          this.renderedPages[n] = false;
          this.renderPage(n);
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
    goToPage(n) {
      const target = Math.min(Math.max(1, Math.floor(n) || 1), this.pageCount || 1);
      const changed = target !== this.currentPage;

      this.currentPage = target;
      this.pageInput = String(target);

      const el = (this.$refs.page || [])[target - 1];

      if (el && typeof el.scrollIntoView === "function") {
        el.scrollIntoView({ block: "start", behavior: "smooth" });
      }

      this.renderPage(target);
      this.scrollActiveThumbIntoView();

      if (changed) {
        this.$emit("page-changed", target);
      }
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
