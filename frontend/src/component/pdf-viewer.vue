<template>
  <div
    class="p-pdf-viewer"
    :class="{ 'is-loading': loading, 'is-error': !!errorMessage, 'is-thumbs-hidden': !thumbsVisible, 'controls-visible': controlsVisible }"
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
      <div ref="thumbs" class="p-pdf-viewer__thumbs hidden-xs">
        <button
          v-for="n in pageCount"
          :key="'thumb-' + n"
          type="button"
          class="p-pdf-viewer__thumb"
          :class="{ 'is-active': n === currentPage }"
          :aria-label="$gettext('Go to Page %{n}', { n })"
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
        @mousedown="onPagesMouseDown"
        @keydown.left.exact.prevent.stop="$emit('media-prev')"
        @keydown.right.exact.prevent.stop="$emit('media-next')"
        @touchstart="onPagesTouchStart"
        @touchmove="onPagesTouchMove"
        @touchend="onPagesTouchEnd"
        @touchcancel="onPagesTouchEnd"
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
          class="hidden-xs"
          :icon="thumbsVisible ? 'mdi-view-grid' : 'mdi-view-grid-outline'"
          variant="text"
          density="comfortable"
          :title="$gettext('Toggle Thumbnails')"
          :aria-label="$gettext('Toggle Thumbnails')"
          @click="toggleThumbs"
        ></v-btn>
        <span class="p-pdf-viewer__sep"></span>
        <v-btn
          icon="mdi-chevron-left"
          variant="text"
          density="comfortable"
          :disabled="currentPage <= 1"
          :title="$gettext('Previous Page')"
          :aria-label="$gettext('Previous Page')"
          @click="prevPage"
        ></v-btn>
        <div class="p-pdf-viewer__pageinfo">
          <input
            class="p-pdf-viewer__pageinput"
            type="text"
            inputmode="numeric"
            :value="pageInput"
            :aria-label="$gettext('Go to Page')"
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
          :title="$gettext('Next Page')"
          :aria-label="$gettext('Next Page')"
          @click="nextPage"
        ></v-btn>
        <span class="p-pdf-viewer__sep"></span>
        <v-btn
          icon="mdi-magnify-minus-outline"
          variant="text"
          density="comfortable"
          :disabled="zoom <= minZoom"
          :title="$gettext('Zoom Out')"
          :aria-label="$gettext('Zoom Out')"
          @click="zoomOut"
        ></v-btn>
        <v-btn
          icon="mdi-magnify-plus-outline"
          variant="text"
          density="comfortable"
          :disabled="zoom >= maxZoom"
          :title="$gettext('Zoom In')"
          :aria-label="$gettext('Zoom In')"
          @click="zoomIn"
        ></v-btn>
      </div>
      <button
        v-if="hasPrev"
        type="button"
        class="p-pdf-viewer__nav p-pdf-viewer__nav--prev"
        :title="$gettext('Previous')"
        :aria-label="$gettext('Previous')"
        @click="$emit('media-prev')"
      >
        <svg class="p-pdf-viewer__nav-icn" viewBox="0 0 60 60" aria-hidden="true">
          <path d="M29 43l-3 3-16-16 16-16 3 3-13 13 13 13z"></path>
        </svg>
      </button>
      <button
        v-if="hasNext"
        type="button"
        class="p-pdf-viewer__nav p-pdf-viewer__nav--next"
        :title="$gettext('Next')"
        :aria-label="$gettext('Next')"
        @click="$emit('media-next')"
      >
        <svg class="p-pdf-viewer__nav-icn" viewBox="0 0 60 60" aria-hidden="true">
          <path d="M29 43l-3 3-16-16 16-16 3 3-13 13 13 13z"></path>
        </svg>
      </button>
    </template>
  </div>
</template>

<script>
import { markRaw } from "vue";
import { loadPdfDocument, renderPdfPage, getPdfPageSize, destroyPdfDocument, isRenderCancelled } from "common/pdf";

// Scale used for the small thumbnail-strip previews; independent of the main zoom.
const ThumbScale = 0.2;

// Padding (px) on each side of the page column; kept in sync with the CSS so the
// fit-to-width scale measures the real available width. See css/pdf-viewer.css.
const PagePadding = 16;

// Upper bound (px) for a rendered page's backing-store width. Caps the device-pixel
// multiplier so a zoomed page never exceeds the mobile-Safari canvas limit, which
// would otherwise render blank at high zoom on high-density screens.
const MaxCanvasPx = 4096;

// Edge-swipe tuning: a one-finger swipe that starts within EdgeSwipeZone px of the
// viewport's left or right edge and travels at least EdgeSwipeThreshold px horizontally
// switches documents, leaving swipes in the page body free to scroll it.
const EdgeSwipeZone = 44;
const EdgeSwipeThreshold = 56;

export default {
  name: "PPdfViewer",
  props: {
    // src is the inline-PDF URL (see $util.pdfUrl).
    src: {
      type: String,
      default: "",
    },
    // pages is an optional page-count hint; the loaded document is authoritative.
    pages: {
      type: Number,
      default: 0,
    },
    // hasPrev / hasNext enable the toolbar's media-navigation buttons, which switch to
    // the previous / next lightbox item; the lightbox owns the bounds via its slide index.
    hasPrev: {
      type: Boolean,
      default: false,
    },
    hasNext: {
      type: Boolean,
      default: false,
    },
    // controlsVisible mirrors the lightbox chrome visibility so the desktop overlay
    // navigation arrows fade in and out together with the top bar, like photo slides.
    controlsVisible: {
      type: Boolean,
      default: true,
    },
  },
  emits: ["loaded", "page-changed", "error", "media-prev", "media-next"],
  data() {
    return {
      pdf: null,
      pageCount: this.pages || 0,
      currentPage: 1,
      pageInput: "1",
      // zoom is a fit-to-width multiplier (1.0 = fit width): >1 enlarges with horizontal
      // scrolling, <1 shrinks so a tall portrait page fits the height. An absolute pdf.js
      // scale would make zoom a no-op on phones (native width already exceeds the column).
      zoom: 1.0,
      minZoom: 0.25,
      maxZoom: 4.0,
      zoomStep: 0.25,
      loading: false,
      errorMessage: "",
      thumbsVisible: true,
    };
  },
  watch: {
    zoom() {
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
    // Natural page sizes { width, height } (CSS px at scale 1), cached to derive the
    // fit-to-width scale and the fit-to-page initial zoom.
    this.pageSizes = {};
    // Re-fits pages when the column width changes (window resize, thumbnail toggle).
    this.resizeObserver = null;
    this.resizePending = false;
    // Pinch-to-zoom bookkeeping; the preview scales via CSS during the gesture and the
    // final zoom is committed (re-rendered) once on release.
    this.pinching = false;
    this.pinchStartDist = 0;
    this.pinchStartZoom = 1;
    this.pinchLiveZoom = 0;
    // Active one-finger edge-swipe ({ edge, startX, startY, decided, active }) or null.
    this.edgeSwipe = null;
    // Mouse drag-to-pan bookkeeping; the window-level move/up listeners are bound once
    // so they can be removed after the drag (a mouse has no way to pan otherwise).
    this.panStart = null;
    this.boundPanMove = (ev) => this.onPanMove(ev);
    this.boundPanEnd = () => this.onPanEnd();
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
    if (this.src) {
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
      if (!this.src || this.pdf || this.loading) {
        return;
      }

      this.loading = true;
      this.errorMessage = "";

      try {
        const { pdf, pageCount } = await loadPdfDocument(this.src);

        if (this.destroyed) {
          destroyPdfDocument(pdf);
          return;
        }

        // Keep the pdf.js document non-reactive; a Vue reactive Proxy breaks its
        // private class fields (getPage would throw "object is not the right class").
        this.pdf = markRaw(pdf);
        this.pageCount = pageCount;
        this.$emit("loaded", pageCount);
        await this.$nextTick();
        // Pick the initial zoom before observing/rendering so pages first paint at it.
        await this.setInitialZoom();
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
    // teardown cancels in-flight renders, disconnects observers, and releases the
    // document so an unmounted slide holds no resources.
    teardown() {
      this.onPanEnd();
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

      if (this.resizeObserver) {
        this.resizeObserver.disconnect();
        this.resizeObserver = null;
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

      // Re-fit visible pages when the column width changes so a rotation or split-view
      // resize keeps them at fit-to-width instead of the width captured on first render.
      if (typeof ResizeObserver !== "undefined" && this.$refs.scroll) {
        this.resizeObserver = new ResizeObserver(() => this.onResize());
        this.resizeObserver.observe(this.$refs.scroll);
      }
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
    // onPagesMouseDown starts a grab-and-drag pan when the column overflows (a plain mouse
    // has no horizontal wheel). Window listeners keep the drag alive off the viewer.
    onPagesMouseDown(ev) {
      if (ev.button !== 0) {
        return;
      }

      const el = this.$refs.scroll;

      if (!el || (el.scrollWidth <= el.clientWidth && el.scrollHeight <= el.clientHeight)) {
        return;
      }

      ev.preventDefault();
      el.focus({ preventScroll: true });
      this.panStart = { x: ev.clientX, y: ev.clientY, left: el.scrollLeft, top: el.scrollTop };
      el.classList.add("is-panning");
      window.addEventListener("mousemove", this.boundPanMove);
      window.addEventListener("mouseup", this.boundPanEnd);
    },
    // onPanMove scrolls the column opposite the cursor travel so the page follows the drag.
    onPanMove(ev) {
      const el = this.$refs.scroll;

      if (!el || !this.panStart) {
        return;
      }

      el.scrollLeft = this.panStart.left - (ev.clientX - this.panStart.x);
      el.scrollTop = this.panStart.top - (ev.clientY - this.panStart.y);
    },
    // onPanEnd ends a drag-to-pan and detaches the window listeners.
    onPanEnd() {
      if (this.$refs.scroll) {
        this.$refs.scroll.classList.remove("is-panning");
      }

      this.panStart = null;
      window.removeEventListener("mousemove", this.boundPanMove);
      window.removeEventListener("mouseup", this.boundPanEnd);
    },
    // onThumbsIntersect lazily renders thumbnails as they scroll into view.
    onThumbsIntersect(entries) {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          this.renderThumb(Number(entry.target.dataset.page));
        }
      }
    },
    // renderPage renders page n onto its canvas at the fit-to-width zoom, ignoring
    // benign cancellations. The slot is reserved before the first await so a second
    // intersection callback for the same page cannot start a duplicate render.
    async renderPage(n) {
      if (!this.pdf || this.renderTasks[n] || this.renderedPages[n]) {
        return;
      }

      const canvas = (this.$refs.canvas || [])[n - 1];

      if (!canvas) {
        return;
      }

      const slot = { canceled: false, task: null, cancel() {
        this.canceled = true;
        if (this.task) {
          this.task.cancel();
        }
      } };
      this.renderTasks[n] = slot;

      try {
        const { scale, displayWidth } = await this.pageScale(n);

        if (this.destroyed || slot.canceled || !canvas) {
          return;
        }

        const handle = renderPdfPage(this.pdf, n, canvas, scale, this.renderDpr(displayWidth));
        slot.task = handle;
        await handle.promise;

        if (!slot.canceled) {
          this.renderedPages[n] = true;
        }
      } catch (e) {
        if (!isRenderCancelled(e)) {
          console.error("pdf: failed to render page", n, e);
        }
      } finally {
        if (this.renderTasks[n] === slot) {
          delete this.renderTasks[n];
        }
      }
    },
    // naturalSize returns page n's natural { width, height } in CSS px (viewport at scale 1),
    // fetched once and cached.
    async naturalSize(n) {
      if (!this.pageSizes[n]) {
        const size = await getPdfPageSize(this.pdf, n);
        this.pageSizes[n] = { width: size.width || 0, height: size.height || 0 };
      }

      return this.pageSizes[n];
    },
    // pageScale resolves the fit-to-width scale for page n at the current zoom and the
    // resulting display width. The page fits the column at zoom 1, so 1 divides the
    // available width by the page's natural width; higher zoom multiplies from there.
    async pageScale(n) {
      const width = (await this.naturalSize(n)).width;
      const avail = this.availableWidth() || width;
      const fit = width > 0 ? avail / width : 1;
      const scale = Math.max(0.01, fit * this.zoom);

      return { scale, displayWidth: width * scale };
    },
    // availableWidth returns the page column's inner content width in CSS pixels, or 0
    // when it cannot be measured yet (callers then fall back to the page's own width).
    availableWidth() {
      const el = this.$refs.scroll;

      if (!el) {
        return 0;
      }

      const w = el.clientWidth - 2 * PagePadding;

      return w > 0 ? w : 0;
    },
    // renderDpr returns the device-pixel multiplier for the backing store, capped so a
    // zoomed page never exceeds the mobile-Safari canvas limit (blank render above it).
    renderDpr(displayWidth) {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);

      if (!displayWidth || displayWidth <= 0) {
        return dpr;
      }

      return Math.max(1, Math.min(dpr, MaxCanvasPx / displayWidth));
    },
    // onResize re-fits the visible pages after a column-width change, throttled to one
    // pass per animation frame and suppressed mid-pinch so the gesture stays smooth.
    onResize() {
      if (this.resizePending) {
        return;
      }

      this.resizePending = true;
      requestAnimationFrame(() => {
        this.resizePending = false;

        if (this.destroyed || this.pinching) {
          return;
        }

        this.rerenderVisible();
      });
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
    // prevPage steps to the previous page (toolbar button beside the page number).
    prevPage() {
      this.goToPage(this.currentPage - 1);
    },
    // nextPage steps to the next page (toolbar button beside the page number).
    nextPage() {
      this.goToPage(this.currentPage + 1);
    },
    // submitPageInput jumps to the page typed into the jump-to-page field. Enter and
    // the blur it triggers both fire this, so it no-ops once the field already shows
    // the current page to avoid a redundant second jump.
    submitPageInput() {
      const n = parseInt(this.pageInput, 10);

      if (isNaN(n) || n === this.currentPage) {
        this.pageInput = String(this.currentPage);
        return;
      }

      this.goToPage(n);
    },
    // zoomIn increases the zoom level by one step.
    zoomIn() {
      this.setZoom(this.zoom + this.zoomStep);
    },
    // zoomOut decreases the zoom level by one step.
    zoomOut() {
      this.setZoom(this.zoom - this.zoomStep);
    },
    // setZoom clamps and applies a new zoom multiplier.
    setZoom(value) {
      const clamped = this.clampZoom(value);

      if (clamped !== this.zoom) {
        this.zoom = clamped;
      }
    },
    // clampZoom rounds and constrains a zoom multiplier to the supported range.
    clampZoom(value) {
      return Math.min(this.maxZoom, Math.max(this.minZoom, Math.round(value * 100) / 100));
    },
    // setInitialZoom picks the opening zoom once per open: fit-to-page when a fit-to-width
    // first page overflows the viewport height, else fit-to-width (see the related spec).
    async setInitialZoom() {
      const natural = await this.naturalSize(1);

      if (this.destroyed) {
        return;
      }

      const el = this.$refs.scroll;
      const contentHeight = el ? el.clientHeight - 2 * PagePadding : 0;
      this.zoom = this.fitPageZoom(natural, this.availableWidth(), contentHeight);
    },
    // fitPageZoom returns the zoom that fits the page into the column width and viewport
    // height, clamped to [minZoom, 1.0]; 1.0 (fit-width) when the inputs can't be measured.
    fitPageZoom(natural, avail, contentHeight) {
      if (!natural || !natural.width || !natural.height || avail <= 0 || contentHeight <= 0) {
        return 1;
      }

      const pageHeightAtFitWidth = avail * (natural.height / natural.width);

      return this.clampZoom(Math.min(1, contentHeight / pageHeightAtFitWidth));
    },
    // touchDistance returns the pixel distance between the first two active touches.
    touchDistance(touches) {
      const dx = touches[0].clientX - touches[1].clientX;
      const dy = touches[0].clientY - touches[1].clientY;

      return Math.hypot(dx, dy);
    },
    // pinchZoomFor derives the clamped zoom for a pinch from its start distance/zoom.
    pinchZoomFor(startZoom, startDist, dist) {
      if (!startDist) {
        return startZoom;
      }

      return this.clampZoom(startZoom * (dist / startDist));
    },
    // onPagesTouchStart starts a two-finger pinch, or arms a one-finger edge-swipe when the
    // touch lands within EdgeSwipeZone of an edge (so a page-body drag still scrolls).
    onPagesTouchStart(ev) {
      if (ev.touches && ev.touches.length === 2) {
        this.pinching = true;
        this.pinchStartDist = this.touchDistance(ev.touches);
        this.pinchStartZoom = this.zoom;
        this.pinchLiveZoom = this.zoom;
        this.edgeSwipe = null;
        return;
      }

      if (ev.touches && ev.touches.length === 1) {
        const t = ev.touches[0];
        const w = window.innerWidth || 0;
        const nearLeft = t.clientX <= EdgeSwipeZone;
        const nearRight = w > 0 && t.clientX >= w - EdgeSwipeZone;
        this.edgeSwipe = nearLeft || nearRight ? { edge: nearLeft ? "left" : "right", startX: t.clientX, startY: t.clientY, decided: false, active: false } : null;
      }
    },
    // onPagesTouchMove previews the pinch zoom (CSS transform, no re-render), or locks a
    // horizontal edge-swipe once its axis is clear so it navigates instead of scrolling.
    onPagesTouchMove(ev) {
      if (this.pinching && ev.touches && ev.touches.length === 2) {
        ev.preventDefault();
        this.pinchLiveZoom = this.pinchZoomFor(this.pinchStartZoom, this.pinchStartDist, this.touchDistance(ev.touches));
        this.applyPinchPreview(this.pinchLiveZoom);
        return;
      }

      if (this.edgeSwipe && ev.touches && ev.touches.length === 1) {
        const t = ev.touches[0];
        const dx = t.clientX - this.edgeSwipe.startX;
        const dy = t.clientY - this.edgeSwipe.startY;

        if (!this.edgeSwipe.decided && (Math.abs(dx) > 10 || Math.abs(dy) > 10)) {
          this.edgeSwipe.decided = true;
          this.edgeSwipe.active = Math.abs(dx) > Math.abs(dy);
        }

        if (this.edgeSwipe.active) {
          ev.preventDefault();
        }
      }
    },
    // onPagesTouchEnd commits a pinch zoom, or switches documents when an edge-swipe traveled
    // far enough (inward from the left edge = previous item, from the right edge = next).
    onPagesTouchEnd(ev) {
      if (this.pinching && (!ev.touches || ev.touches.length < 2)) {
        this.pinching = false;
        this.clearPinchPreview();

        const target = this.pinchLiveZoom;
        this.pinchLiveZoom = 0;

        if (target && target !== this.zoom) {
          this.setZoom(target);
        }

        return;
      }

      if (this.edgeSwipe && this.edgeSwipe.active) {
        const t = ev.changedTouches && ev.changedTouches[0];
        const dx = t ? t.clientX - this.edgeSwipe.startX : 0;

        if (Math.abs(dx) >= EdgeSwipeThreshold) {
          if (this.edgeSwipe.edge === "left" && dx > 0 && this.hasPrev) {
            this.$emit("media-prev");
          } else if (this.edgeSwipe.edge === "right" && dx < 0 && this.hasNext) {
            this.$emit("media-next");
          }
        }
      }

      this.edgeSwipe = null;
    },
    // applyPinchPreview scales the rendered pages via CSS relative to the committed zoom.
    applyPinchPreview(live) {
      const el = this.$refs.scroll;

      if (!el || !this.zoom) {
        return;
      }

      el.style.setProperty("--pdf-pinch", String(live / this.zoom));
      el.classList.add("is-pinching");
    },
    // clearPinchPreview removes the pinch transform so the committed render lays out normally.
    clearPinchPreview() {
      const el = this.$refs.scroll;

      if (!el) {
        return;
      }

      el.classList.remove("is-pinching");
      el.style.removeProperty("--pdf-pinch");
    },
    // toggleThumbs shows or hides the page-thumbnail strip (closed by default on
    // small screens). Thumbnails render lazily, so the strip catches up when shown;
    // the pages re-fit to the width freed or taken by the strip.
    toggleThumbs() {
      this.thumbsVisible = !this.thumbsVisible;
      this.$nextTick(() => this.rerenderVisible());
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
