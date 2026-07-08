<template>
  <div
    ref="root"
    class="p-meta-face-markers"
    :class="{ 'is-edit': isEditMode, 'is-display': !isEditMode }"
    :style="rootStyle"
    @pointerdown="onPointerDown"
    @pointermove="onHoverMove"
    @pointerleave="onHoverLeave"
    @wheel="onWheel"
  >
    <svg v-if="bounds" class="p-meta-face-markers__svg" :style="svgStyle" :viewBox="`0 0 ${bounds.width} ${bounds.height}`">
      <g v-for="m in markers" :key="markerKey(m)">
        <rect
          class="p-meta-face-markers__rect"
          :class="{
            'p-meta-face-markers__rect--named': !!m.Name,
            'p-meta-face-markers__rect--removing': removingMarker && removingMarker.UID === m.UID,
            'p-meta-face-markers__rect--hovered': hoveredUid && hoveredUid === m.UID,
          }"
          :x="m.X * bounds.width"
          :y="m.Y * bounds.height"
          :width="m.W * bounds.width"
          :height="m.H * bounds.height"
        >
          <title v-if="m.Name">{{ m.Name }}</title>
        </rect>
        <text
          v-if="m.Name"
          class="p-meta-face-markers__label"
          text-anchor="middle"
          :x="m.X * bounds.width + (m.W * bounds.width) / 2"
          :y="m.Y * bounds.height + m.H * bounds.height + 16"
        >
          {{ m.Name }}
        </text>
      </g>
      <rect
        v-if="activeDraft"
        class="p-meta-face-markers__rect p-meta-face-markers__rect--draft"
        :x="activeDraft.x"
        :y="activeDraft.y"
        :width="activeDraft.w"
        :height="activeDraft.h"
      ></rect>
      <g v-if="pending && !interaction">
        <circle class="p-meta-face-markers__handle p-meta-face-markers__handle--tl" :cx="pending.x" :cy="pending.y" r="6"></circle>
        <circle class="p-meta-face-markers__handle p-meta-face-markers__handle--tr" :cx="pending.x + pending.w" :cy="pending.y" r="6"></circle>
        <circle class="p-meta-face-markers__handle p-meta-face-markers__handle--bl" :cx="pending.x" :cy="pending.y + pending.h" r="6"></circle>
        <circle class="p-meta-face-markers__handle p-meta-face-markers__handle--br" :cx="pending.x + pending.w" :cy="pending.y + pending.h" r="6"></circle>
      </g>
    </svg>
    <div v-if="pending && bounds && !interaction" class="p-meta-face-markers__confirm" :style="confirmStyle" @pointerdown.stop @pointerup.stop>
      <button type="button" class="p-meta-face-markers__btn p-meta-face-markers__btn--cancel" :title="$gettext('Cancel')" @click.stop="onCancelPending">
        <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
          <path fill="currentColor" d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"></path>
        </svg>
      </button>
      <button
        type="button"
        class="p-meta-face-markers__btn p-meta-face-markers__btn--confirm"
        :class="{ 'is-disabled': busy }"
        :disabled="busy"
        :title="$gettext('Confirm')"
        @click.stop="onConfirmPending"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
          <path fill="currentColor" d="M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"></path>
        </svg>
      </button>
    </div>
    <div v-if="removingMarker && bounds" class="p-meta-face-markers__remove-confirm" :style="removeConfirmStyle" @pointerdown.stop @pointerup.stop>
      <button type="button" class="p-meta-face-markers__btn p-meta-face-markers__btn--cancel" :title="$gettext('Cancel')" @click.stop="onCancelRemove">
        <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
          <path fill="currentColor" d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"></path>
        </svg>
      </button>
      <button
        type="button"
        class="p-meta-face-markers__btn p-meta-face-markers__btn--remove"
        :class="{ 'is-disabled': busy }"
        :disabled="busy"
        :title="$gettext('Remove')"
        @click.stop="onConfirmRemove"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
          <path fill="currentColor" d="M9,3V4H4V6H5V19A2,2 0 0,0 7,21H17A2,2 0 0,0 19,19V6H20V4H15V3H9M7,6H17V19H7V6M9,8V17H11V8H9M13,8V17H15V8H13Z"></path>
        </svg>
      </button>
    </div>
    <button
      type="button"
      class="p-meta-face-markers__btn p-meta-face-markers__btn--back"
      :title="$gettext('Back')"
      :aria-label="$gettext('Back')"
      @click.stop="onBackClick"
      @pointerdown.stop
      @pointerup.stop
    >
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
        <path v-if="$isRtl" fill="currentColor" d="M4 11h12.17l-5.59-5.59L12 4l8 8-8 8-1.41-1.41L16.17 13H4z"></path>
        <path v-else fill="currentColor" d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20z"></path>
      </svg>
    </button>
  </div>
</template>

<script>
import { FaceMarkerDisplay, FaceMarkerEdit, isFaceMarkerMode } from "options/face-marker";

// Minimum side length of a drawable square, in screen pixels.
const MIN_DRAW_SIZE = 16;

// Internal pointer-interaction kinds used by the overlay's draw / move /
// resize gestures. Named separately from the public face-marker mode
// constants so the same word "draw" can carry distinct semantics in each
// scope without confusion.
const InteractionDraw = "draw";
const InteractionMove = "move";
const InteractionResize = "resize";

export default {
  name: "PMetaFaceMarkers",
  props: {
    markers: {
      type: Array,
      default: () => [],
    },
    pswp: {
      type: Object,
      default: null,
    },
    mode: {
      type: String,
      default: FaceMarkerDisplay,
      validator: isFaceMarkerMode,
    },
    busy: {
      type: Boolean,
      default: false,
    },
    // hoveredUid is the UID of the marker that should render with the
    // `--hovered` highlight (thicker, accent-colored stroke). Forwarded
    // from `$faceMarkers.hoveredMarkerUid` by the lightbox so sidebar
    // people-row hover and direct rect hover stay in sync.
    hoveredUid: {
      type: String,
      default: "",
    },
  },
  emits: ["create", "cancel", "remove"],
  data() {
    return {
      bounds: null,
      draft: null,
      pending: null,
      interaction: null, // null | InteractionDraw | InteractionMove | InteractionResize
      resizeCorner: null,
      hoverCursor: null,
      pointerId: null,
      dragStart: null,
      rafHandle: null,
      resizeObserver: null,
      // The unnamed marker the user clicked in edit mode. While set, an
      // inline confirm pill anchors below it; ✓ emits `remove`, ✕ clears.
      // Named markers (m.SubjUID truthy) cannot be removed via this path
      // because the backend's `marker.reject()` only accepts unnamed
      // markers — the user has to eject the name first.
      removingMarker: null,
    };
  },
  computed: {
    isEditMode() {
      return this.mode === FaceMarkerEdit;
    },
    // svgStyle positions the absolute SVG overlay over the letterboxed image area.
    svgStyle() {
      if (!this.bounds) {
        return { display: "none" };
      }
      return {
        position: "absolute",
        left: `${this.bounds.left}px`,
        top: `${this.bounds.top}px`,
        width: `${this.bounds.width}px`,
        height: `${this.bounds.height}px`,
      };
    },
    activeDraft() {
      return this.draft || this.pending;
    },
    rootStyle() {
      return this.hoverCursor ? { cursor: this.hoverCursor } : {};
    },
    // confirmStyle positions the confirm pill centered below the pending rect.
    confirmStyle() {
      if (!this.pending || !this.bounds) {
        return { display: "none" };
      }
      const left = this.bounds.left + this.pending.x + this.pending.w / 2;
      const top = this.bounds.top + this.pending.y + this.pending.h;
      return {
        position: "absolute",
        left: `${left}px`,
        top: `${top}px`,
        transform: "translate(-50%, 8px)",
      };
    },
    // Pixel rect of the marker pending removal, in the overlay's local
    // coordinate space. Used to anchor the remove-confirm pill and to
    // highlight the target rectangle.
    removingMarkerRect() {
      if (!this.removingMarker || !this.bounds) {
        return null;
      }
      const m = this.removingMarker;
      return {
        x: m.X * this.bounds.width,
        y: m.Y * this.bounds.height,
        w: m.W * this.bounds.width,
        h: m.H * this.bounds.height,
      };
    },
    // removeConfirmStyle positions the remove-confirm pill centered below the targeted marker.
    removeConfirmStyle() {
      const r = this.removingMarkerRect;
      if (!r || !this.bounds) {
        return { display: "none" };
      }
      const left = this.bounds.left + r.x + r.w / 2;
      const top = this.bounds.top + r.y + r.h;
      return {
        position: "absolute",
        left: `${left}px`,
        top: `${top}px`,
        transform: "translate(-50%, 8px)",
      };
    },
    // markerRects precomputes unnamed-marker pixel rects from the current
    // bounds so hover and pointer-down hit-testing reuse them instead of
    // recomputing every marker's rect on each pointer event. Vue caches it on
    // `bounds` + `markers`, so it only recomputes when those change. Named
    // markers are excluded — they are not removable via this overlay.
    markerRects() {
      if (!this.bounds || !Array.isArray(this.markers)) {
        return [];
      }
      const rects = [];
      for (const m of this.markers) {
        if (!m || m.SubjUID) {
          continue;
        }
        rects.push({
          marker: m,
          x: m.X * this.bounds.width,
          y: m.Y * this.bounds.height,
          w: m.W * this.bounds.width,
          h: m.H * this.bounds.height,
        });
      }
      return rects;
    },
  },
  watch: {
    // mode cancels any active draft and pending remove when leaving edit mode.
    mode(newVal) {
      if (newVal !== FaceMarkerEdit) {
        this.cancelActiveDraft();
        this.removingMarker = null;
      }
    },
    // busy clears the synchronous confirm lock once the parent's save settles —
    // success clears the pending; a failed save keeps it so retry must re-enable.
    busy(newVal) {
      if (!newVal) {
        this._confirming = false;
      }
    },
  },
  mounted() {
    // Non-reactive scratch state for rAF-batched drag flushing and the cached
    // overlay-root rect (kept off `data()` so intermediate rects don't render).
    this._dragRect = null;
    this._dragTarget = null;
    this._dragRaf = null;
    this._dragScheduled = false;
    this._cachedParentRect = null;
    this._parentRectFresh = false;
    // Synchronous in-flight guard so a rapid second confirm cannot emit `create`
    // again before the parent's async `busy` lock propagates back as a prop.
    this._confirming = false;

    this.attachPswpListeners();
    this.attachImageLoadListener();
    this.scheduleUpdate();

    this.onWindowResize = () => this.scheduleUpdate();
    window.addEventListener("resize", this.onWindowResize);

    if (typeof ResizeObserver === "function") {
      this.resizeObserver = new ResizeObserver(() => this.scheduleUpdate());
      if (this.$refs.root) {
        this.resizeObserver.observe(this.$refs.root);
      }
    }
  },
  beforeUnmount() {
    this.detachPswpListeners();
    this.detachImageLoadListener();
    window.removeEventListener("resize", this.onWindowResize);
    window.removeEventListener("pointermove", this.onPointerMove);
    window.removeEventListener("pointerup", this.onPointerUp);
    window.removeEventListener("pointercancel", this.onPointerUp);

    if (this.rafHandle) {
      cancelAnimationFrame(this.rafHandle);
      this.rafHandle = null;
    }

    this.cancelQueuedDrag();

    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
  },
  methods: {
    // imageElement returns the current PhotoSwipe slide's <img> element, or null if none.
    imageElement() {
      const el = this.pswp?.currSlide?.content?.element;
      if (el instanceof HTMLImageElement) {
        return el;
      }
      if (el && typeof el.querySelector === "function") {
        return el.querySelector("img.pswp__image");
      }
      return null;
    },
    // Subscribes to the image's `load` event so updateBounds is called once
    // `naturalWidth/naturalHeight` become available. The letterbox math
    // relies on those intrinsic dimensions, and the <img> for video / live
    // slides is added without explicit dims — so bounds computed before
    // `load` would fall back to the box rect.
    attachImageLoadListener() {
      const img = this.imageElement();
      if (!img) {
        this._loadListenedImg = null;
        return;
      }
      if (this._loadListenedImg === img) {
        return;
      }
      this.detachImageLoadListener();
      this._loadListenedImg = img;
      this._onImgLoad = () => this.scheduleUpdate();
      img.addEventListener("load", this._onImgLoad);
      // A cached image is already complete when we attach, so its `load` event
      // will never fire; trigger a bounds update now so markers initialize.
      if (img.complete && img.naturalWidth > 0) {
        this.scheduleUpdate();
      }
    },
    // detachImageLoadListener removes the image load listener and clears its refs.
    detachImageLoadListener() {
      if (this._loadListenedImg && this._onImgLoad) {
        this._loadListenedImg.removeEventListener("load", this._onImgLoad);
      }
      this._loadListenedImg = null;
      this._onImgLoad = null;
    },
    // attachPswpListeners subscribes to PhotoSwipe zoom/change/resize events to keep bounds in sync.
    attachPswpListeners() {
      if (!this.pswp || typeof this.pswp.on !== "function") {
        return;
      }
      this._onZoomPan = () => this.scheduleUpdate(false);
      this._onChange = () => {
        this.attachImageLoadListener();
        this.scheduleUpdate();
      };
      this._onResize = () => this.scheduleUpdate();
      this.pswp.on("zoomPanUpdate", this._onZoomPan);
      this.pswp.on("change", this._onChange);
      this.pswp.on("resize", this._onResize);
      this.pswp.on("imageClickAction", this._onChange);
    },
    // detachPswpListeners unsubscribes the PhotoSwipe event handlers.
    detachPswpListeners() {
      if (!this.pswp || typeof this.pswp.off !== "function") {
        return;
      }
      if (this._onZoomPan) {
        this.pswp.off("zoomPanUpdate", this._onZoomPan);
      }
      if (this._onChange) {
        this.pswp.off("change", this._onChange);
        this.pswp.off("imageClickAction", this._onChange);
      }
      if (this._onResize) {
        this.pswp.off("resize", this._onResize);
      }
    },
    // scheduleUpdate batches a bounds recompute into the next animation frame.
    // `invalidateParent` (default true) marks the cached overlay-root rect
    // stale; the zoom path passes false since the container doesn't move during
    // pinch-zoom. The flag is set before the rafHandle dedup so dirtiness still
    // accumulates when a frame is coalesced.
    scheduleUpdate(invalidateParent = true) {
      if (invalidateParent) {
        this._parentRectFresh = false;
      }
      if (this.rafHandle) {
        return;
      }
      this.rafHandle = requestAnimationFrame(() => {
        this.rafHandle = null;
        this.updateBounds();
      });
    },
    // parentRect returns the overlay-root rect, cached across pinch-zoom frames
    // where only the image transform (not the container) changes; layout
    // events invalidate it via scheduleUpdate(true).
    parentRect() {
      if (!this._parentRectFresh || !this._cachedParentRect) {
        this._cachedParentRect = this.$refs.root.getBoundingClientRect();
        this._parentRectFresh = true;
      }
      return this._cachedParentRect;
    },
    // updateBounds recomputes the overlay pixel bounds from the image rect, insetting for letterboxed (object-fit: contain) slides.
    updateBounds() {
      const img = this.imageElement();
      if (!img || !this.$refs.root) {
        if (this.bounds !== null) {
          this.bounds = null;
        }
        return;
      }
      const imgRect = img.getBoundingClientRect();
      const parentRect = this.parentRect();
      if (imgRect.width <= 0 || imgRect.height <= 0) {
        if (this.bounds !== null) {
          this.bounds = null;
        }
        return;
      }
      // getBoundingClientRect returns the <img> box, not the letterboxed pixel
      // content (CSS object-fit: contain on video/live/animated slides). Compute
      // the inscribed rect from the natural aspect ratio so marker coords land
      // on the image; for plain image slides this is a no-op.
      let left = imgRect.left - parentRect.left;
      let top = imgRect.top - parentRect.top;
      let width = imgRect.width;
      let height = imgRect.height;
      const nW = img.naturalWidth || 0;
      const nH = img.naturalHeight || 0;
      if (nW > 0 && nH > 0) {
        const naturalRatio = nW / nH;
        const boxRatio = width / height;
        const tol = 0.001;
        if (Math.abs(naturalRatio - boxRatio) > tol) {
          if (naturalRatio > boxRatio) {
            // image wider than box → letterbox top + bottom
            const inscribedH = width / naturalRatio;
            top += (height - inscribedH) / 2;
            height = inscribedH;
          } else {
            // image taller than box → pillarbox left + right
            const inscribedW = height * naturalRatio;
            left += (width - inscribedW) / 2;
            width = inscribedW;
          }
        }
      }
      // Skip the assignment when nothing changed so Vue does not rerender the
      // SVG children on every zoomPanUpdate tick while the image is idle.
      const b = this.bounds;
      if (b && b.left === left && b.top === top && b.width === width && b.height === height) {
        return;
      }
      this.bounds = { left, top, width, height };
    },
    // onPointerDown begins a resize, move, remove-click, or new draft depending on what the pointer lands on (edit mode only).
    onPointerDown(ev) {
      if (!this.isEditMode) {
        return;
      }

      if (!this.bounds) {
        this._parentRectFresh = false;
        this.updateBounds();
        if (!this.bounds) {
          return;
        }
      }

      if (ev.button !== undefined && ev.button !== 0) {
        return;
      }

      const local = this.toLocal(ev.clientX, ev.clientY);
      if (!this.insideBounds(local)) {
        return;
      }

      if (this.pending) {
        const corner = this.hitTestCorner(local, this.pending, this.handleHitRadius(ev));
        if (corner) {
          this.beginResize(corner, ev);
          return;
        }
        if (this.insidePending(local, this.pending)) {
          this.beginMove(local, ev);
          return;
        }
      }

      // Hit-test existing unnamed markers before starting a new draft; a
      // click inside one opens its remove pill. Named markers are skipped —
      // marker.reject() rejects only unnamed markers (eject the name first).
      const target = this.findMarkerAt(local);
      if (target) {
        this.stopEventFromPswp(ev);
        this.removingMarker = target;
        return;
      }

      // Clicking outside a marker cancels any pending remove pill so a
      // fresh draw can start from the same gesture without a prior
      // click "stealing" focus.
      if (this.removingMarker) {
        this.removingMarker = null;
      }

      this.stopEventFromPswp(ev);
      this.pending = null;
      this.interaction = InteractionDraw;
      this.pointerId = ev.pointerId;
      this.dragStart = { clientX: ev.clientX, clientY: ev.clientY, local };
      this.draft = { x: local.x, y: local.y, w: 0, h: 0 };

      this.attachWindowPointerListeners();
    },
    // markerKey returns a stable, unique :key for a marker row. Saved markers
    // use their UID (CropID kept as a legacy fallback); brand-new unsaved
    // markers have no UID yet, so derive a key from the rect geometry — stable
    // across the fresh Marker instances getMarkers() returns each render, and
    // unique per rect, so Vue never collapses keyless rows or reuses/misorders
    // their DOM nodes.
    markerKey(m) {
      if (!m) {
        return "marker-nil";
      }
      if (m.UID) {
        return m.UID;
      }
      if (m.CropID) {
        return m.CropID;
      }
      return `marker-tmp-${m.X}-${m.Y}-${m.W}-${m.H}`;
    },
    // Returns the first unnamed marker whose pixel rect contains the
    // given local point, or null if none. Reuses the precomputed
    // `markerRects` so hit-testing stays O(n) reads, not O(n) multiplies.
    findMarkerAt(local) {
      for (const r of this.markerRects) {
        if (this.insidePending(local, r)) {
          return r.marker;
        }
      }
      return null;
    },
    // queueDrag stores the latest draft/pending rect and flushes it to reactive
    // state at most once per animation frame, collapsing multiple pointermoves
    // per frame into a single SVG re-render. `target` is "draft" or "pending".
    // The boolean guard is set before scheduling so the synchronous rAF stub
    // used in tests still flushes immediately and leaves the guard reset.
    queueDrag(target, rect) {
      this._dragTarget = target;
      this._dragRect = rect;
      if (this._dragScheduled) {
        return;
      }
      this._dragScheduled = true;
      this._dragRaf = requestAnimationFrame(() => {
        this._dragRaf = null;
        this._dragScheduled = false;
        this.applyQueuedDrag();
      });
    },
    // applyQueuedDrag writes the held rect to the reactive draft/pending field.
    applyQueuedDrag() {
      if (!this._dragRect) {
        return;
      }
      if (this._dragTarget === "pending") {
        this.pending = this._dragRect;
      } else {
        this.draft = this._dragRect;
      }
      this._dragRect = null;
    },
    // flushQueuedDrag applies any rAF-batched rect immediately, e.g. on pointer
    // up so the committed rect reflects the final pointer position.
    flushQueuedDrag() {
      if (this._dragRaf) {
        cancelAnimationFrame(this._dragRaf);
        this._dragRaf = null;
      }
      this._dragScheduled = false;
      this.applyQueuedDrag();
    },
    // cancelQueuedDrag drops any queued rect without applying it (cancel,
    // escape, unmount) so a late flush cannot overwrite restored state.
    cancelQueuedDrag() {
      if (this._dragRaf) {
        cancelAnimationFrame(this._dragRaf);
        this._dragRaf = null;
      }
      this._dragScheduled = false;
      this._dragRect = null;
      this._dragTarget = null;
    },
    // ✓ in the remove-confirm pill. Emits `remove` with the marker so
    // the lightbox can call marker.reject() and re-derive the overlay
    // from the updated photo state.
    onConfirmRemove() {
      const m = this.removingMarker;
      if (!m) {
        return;
      }
      this.removingMarker = null;
      this.$emit("remove", m);
    },
    // ✕ in the remove-confirm pill. Dismisses without mutation.
    onCancelRemove() {
      this.removingMarker = null;
    },
    // onPointerMove updates the active draft/move/resize rect from the pointer, keeping it square and inside bounds.
    onPointerMove(ev) {
      if (!this.interaction || !this.dragStart || !this.bounds) {
        return;
      }
      if (this.pointerId !== null && ev.pointerId !== this.pointerId) {
        return;
      }

      const local = this.toLocal(ev.clientX, ev.clientY);
      const cx = Math.max(0, Math.min(this.bounds.width, local.x));
      const cy = Math.max(0, Math.min(this.bounds.height, local.y));

      if (this.interaction === InteractionMove) {
        const origin = this.dragStart.pending;
        if (!origin) {
          return;
        }
        const dx = local.x - this.dragStart.local.x;
        const dy = local.y - this.dragStart.local.y;
        let nx = origin.x + dx;
        let ny = origin.y + dy;
        if (nx < 0) {
          nx = 0;
        }
        if (ny < 0) {
          ny = 0;
        }
        if (nx + origin.w > this.bounds.width) {
          nx = this.bounds.width - origin.w;
        }
        if (ny + origin.h > this.bounds.height) {
          ny = this.bounds.height - origin.h;
        }
        this.queueDrag("pending", { x: nx, y: ny, w: origin.w, h: origin.h });
        return;
      }

      // Square-from-anchor math shared by draw (anchor = pointerdown) and
      // resize (anchor = opposite corner). The larger axis wins so the
      // rect stays visually square regardless of drag direction.
      const dx = cx - this.dragStart.local.x;
      const dy = cy - this.dragStart.local.y;

      let side = Math.max(Math.abs(dx), Math.abs(dy));
      const signX = dx < 0 ? -1 : 1;
      const signY = dy < 0 ? -1 : 1;

      if (this.interaction === InteractionResize && side < MIN_DRAW_SIZE) {
        side = MIN_DRAW_SIZE;
      }

      let sx = this.dragStart.local.x;
      let sy = this.dragStart.local.y;
      let sw = side;
      let sh = side;

      if (signX < 0) {
        sx = this.dragStart.local.x - side;
      }
      if (signY < 0) {
        sy = this.dragStart.local.y - side;
      }

      if (sx < 0) {
        sw += sx;
        sh += sx;
        sx = 0;
      }
      if (sy < 0) {
        sw += sy;
        sh += sy;
        sy = 0;
      }
      if (sx + sw > this.bounds.width) {
        const over = sx + sw - this.bounds.width;
        sw -= over;
        sh -= over;
      }
      if (sy + sh > this.bounds.height) {
        const over = sy + sh - this.bounds.height;
        sw -= over;
        sh -= over;
      }

      if (sw < 0) {
        sw = 0;
      }
      if (sh < 0) {
        sh = 0;
      }

      if (this.interaction === InteractionResize) {
        this.queueDrag("pending", { x: sx, y: sy, w: sw, h: sh });
      } else {
        this.queueDrag("draft", { x: sx, y: sy, w: sw, h: sh });
      }
    },
    // onPointerUp ends the gesture, promoting a large-enough draw draft into a confirmable pending rect.
    onPointerUp(ev) {
      if (!this.interaction) {
        return;
      }
      if (this.pointerId !== null && ev && ev.pointerId !== this.pointerId) {
        return;
      }

      this.detachWindowPointerListeners();
      // Apply any rAF-batched move so the committed rect reflects the final
      // pointer position even if the frame flush hasn't fired yet.
      this.flushQueuedDrag();

      const wasInteraction = this.interaction;
      const draft = this.draft;

      this.interaction = null;
      this.resizeCorner = null;
      this.pointerId = null;
      this.dragStart = null;
      this.draft = null;

      // Move/resize have already written the up-to-date `pending`; only
      // the draw path needs to promote its draft into pending.
      if (wasInteraction !== InteractionDraw) {
        return;
      }

      if (!draft || !this.bounds || draft.w < MIN_DRAW_SIZE || draft.h < MIN_DRAW_SIZE) {
        return;
      }

      // A freshly promoted pending is a new marker — allow it to be confirmed.
      this._confirming = false;
      this.pending = draft;
    },
    // onConfirmPending emits the normalized pending rect as a create event once, guarded against double-confirm.
    onConfirmPending() {
      if (this.busy || this._confirming) {
        return;
      }

      const pending = this.pending;
      const bounds = this.bounds;
      if (!pending || !bounds) {
        return;
      }

      this._confirming = true;

      const nx = pending.x / bounds.width;
      const ny = pending.y / bounds.height;
      const nw = pending.w / bounds.width;
      const nh = pending.h / bounds.height;

      this.$emit("create", {
        X: this.clamp01(nx),
        Y: this.clamp01(ny),
        W: this.clamp01(nw),
        H: this.clamp01(nh),
      });
    },
    // onCancelPending discards the pending rect without exiting draw mode.
    onCancelPending() {
      this.pending = null;
      this.hoverCursor = null;
      this._confirming = false;
    },
    // Back-button click. Signals the parent lightbox to exit face-marker
    // mode entirely (display or draw). Uses the existing `cancel` emit
    // so the lightbox's `@cancel="exitFaceMarkerMode"` wiring catches
    // it without a new listener. Distinct from `onCancelPending` —
    // that one discards a draft rect without exiting draw mode.
    onBackClick() {
      this.cancelActiveDraft();
      this.$emit("cancel");
    },
    // Called by the parent only after a successful save — on failure the
    // parent leaves the rect on screen so the user can retry.
    clearPending() {
      this.pending = null;
      this.hoverCursor = null;
      this._confirming = false;
    },
    // cancelActiveDraft tears down any in-progress gesture and clears draft/pending state.
    cancelActiveDraft() {
      if (this.interaction) {
        this.detachWindowPointerListeners();
      }
      this.cancelQueuedDrag();
      this.interaction = null;
      this.resizeCorner = null;
      this.pointerId = null;
      this.dragStart = null;
      this.draft = null;
      this.pending = null;
      this.hoverCursor = null;
      this._confirming = false;
    },
    // handleEnter mirrors a ✓ click; no-op during draft / drag / remove-confirm.
    handleEnter() {
      if (this.busy || this.interaction || this.removingMarker || !this.pending) {
        return;
      }
      this.onConfirmPending();
    },
    // handleEscape cancels in-progress draw/move/resize or clears the pending
    // rect without exiting draw mode; returns true when the overlay consumed it.
    handleEscape() {
      if (this.interaction === InteractionDraw) {
        this.cancelQueuedDrag();
        this.interaction = null;
        this.pointerId = null;
        this.dragStart = null;
        this.draft = null;
        this.detachWindowPointerListeners();
        return true;
      }
      if (this.interaction === InteractionMove || this.interaction === InteractionResize) {
        this.cancelQueuedDrag();
        const snapshot = this.dragStart && this.dragStart.pending;
        if (snapshot) {
          this.pending = { ...snapshot };
        }
        this.interaction = null;
        this.resizeCorner = null;
        this.pointerId = null;
        this.dragStart = null;
        this.detachWindowPointerListeners();
        return true;
      }
      if (this.pending) {
        this.pending = null;
        this._confirming = false;
        return true;
      }
      if (this.removingMarker) {
        this.removingMarker = null;
        return true;
      }
      return false;
    },
    // stopEventFromPswp stops propagation and default so PhotoSwipe doesn't treat the gesture as a pan/zoom.
    stopEventFromPswp(ev) {
      if (typeof ev.stopPropagation === "function") {
        ev.stopPropagation();
      }
      if (typeof ev.preventDefault === "function" && ev.cancelable !== false) {
        ev.preventDefault();
      }
    },
    // attachWindowPointerListeners tracks pointer move/up/cancel on window for the duration of a drag.
    attachWindowPointerListeners() {
      window.addEventListener("pointermove", this.onPointerMove);
      window.addEventListener("pointerup", this.onPointerUp);
      window.addEventListener("pointercancel", this.onPointerUp);
    },
    // detachWindowPointerListeners removes the window-level drag pointer listeners.
    detachWindowPointerListeners() {
      window.removeEventListener("pointermove", this.onPointerMove);
      window.removeEventListener("pointerup", this.onPointerUp);
      window.removeEventListener("pointercancel", this.onPointerUp);
    },
    // handleHitRadius returns the corner grab radius for the event's pointer
    // type: larger for touch/pen (coarse) pointers, where the visible handle is
    // hard to hit; the mouse radius is unchanged so existing behavior holds.
    handleHitRadius(ev) {
      const coarse = !!ev && (ev.pointerType === "touch" || ev.pointerType === "pen");
      return coarse ? 22 : 14;
    },
    // hitTestCorner returns the resize-handle corner key within the given radius of point p, or null.
    hitTestCorner(p, rect, radius = 14) {
      const r = radius;
      const corners = {
        tl: { x: rect.x, y: rect.y },
        tr: { x: rect.x + rect.w, y: rect.y },
        bl: { x: rect.x, y: rect.y + rect.h },
        br: { x: rect.x + rect.w, y: rect.y + rect.h },
      };
      for (const key of Object.keys(corners)) {
        const c = corners[key];
        if (Math.hypot(p.x - c.x, p.y - c.y) <= r) {
          return key;
        }
      }
      return null;
    },
    // insidePending reports whether point p lies within the given rect.
    insidePending(p, rect) {
      return p.x >= rect.x && p.y >= rect.y && p.x <= rect.x + rect.w && p.y <= rect.y + rect.h;
    },
    // The opposite corner becomes the fixed anchor so the square-from-anchor
    // math in onPointerMove works the same way as for the draw path.
    beginResize(corner, ev) {
      const p = this.pending;
      if (!p) {
        return;
      }
      let anchor;
      if (corner === "tl") {
        anchor = { x: p.x + p.w, y: p.y + p.h };
      } else if (corner === "tr") {
        anchor = { x: p.x, y: p.y + p.h };
      } else if (corner === "bl") {
        anchor = { x: p.x + p.w, y: p.y };
      } else {
        anchor = { x: p.x, y: p.y };
      }

      this.stopEventFromPswp(ev);
      this.interaction = InteractionResize;
      this.resizeCorner = corner;
      this.pointerId = ev.pointerId;
      this.dragStart = {
        clientX: ev.clientX,
        clientY: ev.clientY,
        local: anchor,
        pending: { ...p },
      };
      this.attachWindowPointerListeners();
    },
    // onHoverMove sets the cursor to reflect the resize/move/remove affordance under the pointer (edit mode only).
    onHoverMove(ev) {
      if (!this.isEditMode || this.interaction) {
        return;
      }
      if (!this.bounds) {
        if (this.hoverCursor !== null) {
          this.hoverCursor = null;
        }
        return;
      }
      const local = this.toLocal(ev.clientX, ev.clientY);
      if (!this.insideBounds(local)) {
        if (this.hoverCursor !== null) {
          this.hoverCursor = null;
        }
        return;
      }
      if (this.pending) {
        const corner = this.hitTestCorner(local, this.pending, this.handleHitRadius(ev));
        if (corner) {
          const c = corner === "tl" || corner === "br" ? "nwse-resize" : "nesw-resize";
          if (this.hoverCursor !== c) {
            this.hoverCursor = c;
          }
          return;
        }
        if (this.insidePending(local, this.pending)) {
          if (this.hoverCursor !== "move") {
            this.hoverCursor = "move";
          }
          return;
        }
      }
      // Hovering an unnamed marker rect: signal it is clickable for
      // removal. Named markers fall through to the default cursor.
      if (this.findMarkerAt(local)) {
        if (this.hoverCursor !== "pointer") {
          this.hoverCursor = "pointer";
        }
        return;
      }
      if (this.hoverCursor !== null) {
        this.hoverCursor = null;
      }
    },
    // onHoverLeave clears the hover cursor when the pointer leaves the overlay.
    onHoverLeave() {
      if (this.hoverCursor !== null) {
        this.hoverCursor = null;
      }
    },
    // onWheel re-dispatches wheel events on PhotoSwipe's element while in edit
    // mode (overlay's pointer-events: auto would otherwise swallow zoom gestures).
    onWheel(ev) {
      if (!this.isEditMode) {
        return;
      }
      const pswpEl = this.pswp?.element;
      if (!pswpEl) {
        return;
      }
      if (typeof ev.preventDefault === "function" && ev.cancelable !== false) {
        ev.preventDefault();
      }
      pswpEl.dispatchEvent(
        new WheelEvent("wheel", {
          deltaX: ev.deltaX,
          deltaY: ev.deltaY,
          deltaZ: ev.deltaZ,
          deltaMode: ev.deltaMode,
          clientX: ev.clientX,
          clientY: ev.clientY,
          bubbles: true,
          cancelable: true,
          ctrlKey: ev.ctrlKey,
          shiftKey: ev.shiftKey,
          altKey: ev.altKey,
          metaKey: ev.metaKey,
        })
      );
    },
    // beginMove starts dragging the whole pending rect from the given local anchor point.
    beginMove(local, ev) {
      const p = this.pending;
      if (!p) {
        return;
      }
      this.stopEventFromPswp(ev);
      this.interaction = InteractionMove;
      this.resizeCorner = null;
      this.pointerId = ev.pointerId;
      this.dragStart = {
        clientX: ev.clientX,
        clientY: ev.clientY,
        local,
        pending: { ...p },
      };
      this.attachWindowPointerListeners();
    },
    // toLocal converts client coordinates to the overlay's image-local coordinate space.
    toLocal(clientX, clientY) {
      if (!this.bounds || !this.$refs.root) {
        return { x: 0, y: 0 };
      }
      const rect = this.$refs.root.getBoundingClientRect();
      return {
        x: clientX - rect.left - this.bounds.left,
        y: clientY - rect.top - this.bounds.top,
      };
    },
    // insideBounds reports whether local point p lies within the image bounds.
    insideBounds(p) {
      return this.bounds && p.x >= 0 && p.y >= 0 && p.x <= this.bounds.width && p.y <= this.bounds.height;
    },
    // clamp01 clamps a normalized coordinate to [0, 1), capping at 0.999999 to stay below 1.
    clamp01(v) {
      if (v < 0) {
        return 0;
      }
      if (v >= 1) {
        return 0.999999;
      }
      return v;
    },
  },
};
</script>
