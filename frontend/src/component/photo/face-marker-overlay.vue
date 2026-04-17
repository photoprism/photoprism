<template>
  <div
    ref="root"
    class="p-face-markers"
    :class="{ 'is-drawing': isDrawMode, 'is-display': !isDrawMode }"
    :style="rootStyle"
    @pointerdown="onPointerDown"
    @pointermove="onHoverMove"
    @pointerleave="onHoverLeave"
  >
    <svg v-if="bounds" class="p-face-markers__svg" :style="svgStyle" :viewBox="`0 0 ${bounds.width} ${bounds.height}`">
      <template v-for="m in markers" :key="m.UID || m.CropID">
        <rect
          class="p-face-markers__rect"
          :class="{ 'p-face-markers__rect--named': !!m.Name }"
          :x="m.X * bounds.width"
          :y="m.Y * bounds.height"
          :width="m.W * bounds.width"
          :height="m.H * bounds.height"
        >
          <title v-if="m.Name">{{ m.Name }}</title>
        </rect>
        <text
          v-if="m.Name && !isDrawMode"
          class="p-face-markers__label"
          text-anchor="middle"
          :x="m.X * bounds.width + (m.W * bounds.width) / 2"
          :y="m.Y * bounds.height + m.H * bounds.height + 16"
        >
          {{ m.Name }}
        </text>
      </template>
      <rect
        v-if="activeDraft"
        class="p-face-markers__rect p-face-markers__rect--draft"
        :x="activeDraft.x"
        :y="activeDraft.y"
        :width="activeDraft.w"
        :height="activeDraft.h"
      ></rect>
      <g v-if="pending && !interaction">
        <circle class="p-face-markers__handle" :cx="pending.x" :cy="pending.y" r="6"></circle>
        <circle class="p-face-markers__handle" :cx="pending.x + pending.w" :cy="pending.y" r="6"></circle>
        <circle class="p-face-markers__handle" :cx="pending.x" :cy="pending.y + pending.h" r="6"></circle>
        <circle class="p-face-markers__handle" :cx="pending.x + pending.w" :cy="pending.y + pending.h" r="6"></circle>
      </g>
    </svg>
    <div v-if="pending && bounds && !interaction" class="p-face-markers__confirm" :style="confirmStyle" @pointerdown.stop @pointerup.stop>
      <button
        type="button"
        class="p-face-markers__btn p-face-markers__btn--confirm"
        :class="{ 'is-disabled': busy }"
        :disabled="busy"
        :title="$gettext('Confirm')"
        @click.stop="onConfirmPending"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
          <path fill="currentColor" d="M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"></path>
        </svg>
      </button>
      <button type="button" class="p-face-markers__btn p-face-markers__btn--cancel" :title="$gettext('Cancel')" @click.stop="onCancelPending">
        <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
          <path fill="currentColor" d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"></path>
        </svg>
      </button>
    </div>
  </div>
</template>

<script>
const MIN_DRAW_SIZE = 16;

export default {
  name: "PFaceMarkerOverlay",
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
      default: "display",
      validator: (value) => value === "display" || value === "draw",
    },
    busy: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["create", "cancel"],
  data() {
    return {
      bounds: null,
      draft: null,
      pending: null,
      interaction: null,
      resizeCorner: null,
      hoverCursor: null,
      pointerId: null,
      dragStart: null,
      rafHandle: null,
      resizeObserver: null,
    };
  },
  computed: {
    isDrawMode() {
      return this.mode === "draw";
    },
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
  },
  watch: {
    mode(value) {
      if (value !== "draw") {
        this.cancelActiveDraft();
      }
    },
  },
  mounted() {
    this.attachPswpListeners();
    this.scheduleUpdate();
    this.onWindowResize = () => this.scheduleUpdate();
    window.addEventListener("resize", this.onWindowResize);
    window.addEventListener("keydown", this.onKeyDown);

    if (typeof ResizeObserver === "function") {
      this.resizeObserver = new ResizeObserver(() => this.scheduleUpdate());

      if (this.$refs.root) {
        this.resizeObserver.observe(this.$refs.root);
      }
    }
  },
  beforeUnmount() {
    this.detachPswpListeners();
    window.removeEventListener("resize", this.onWindowResize);
    window.removeEventListener("keydown", this.onKeyDown);
    window.removeEventListener("pointermove", this.onPointerMove);
    window.removeEventListener("pointerup", this.onPointerUp);
    window.removeEventListener("pointercancel", this.onPointerUp);

    if (this.rafHandle) {
      cancelAnimationFrame(this.rafHandle);
      this.rafHandle = null;
    }

    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
  },
  methods: {
    imageElement() {
      const element = this.pswp?.currSlide?.content?.element;

      if (element instanceof HTMLImageElement) {
        return element;
      }

      if (element && typeof element.querySelector === "function") {
        return element.querySelector("img.pswp__image");
      }

      return null;
    },
    attachPswpListeners() {
      if (!this.pswp || typeof this.pswp.on !== "function") {
        return;
      }

      this._onZoomPan = () => this.scheduleUpdate();
      this._onChange = () => this.scheduleUpdate();
      this._onResize = () => this.scheduleUpdate();
      this.pswp.on("zoomPanUpdate", this._onZoomPan);
      this.pswp.on("change", this._onChange);
      this.pswp.on("resize", this._onResize);
    },
    detachPswpListeners() {
      if (!this.pswp || typeof this.pswp.off !== "function") {
        return;
      }

      if (this._onZoomPan) {
        this.pswp.off("zoomPanUpdate", this._onZoomPan);
      }

      if (this._onChange) {
        this.pswp.off("change", this._onChange);
      }

      if (this._onResize) {
        this.pswp.off("resize", this._onResize);
      }
    },
    scheduleUpdate() {
      if (this.rafHandle) {
        return;
      }

      this.rafHandle = requestAnimationFrame(() => {
        this.rafHandle = null;
        this.updateBounds();
      });
    },
    updateBounds() {
      const image = this.imageElement();

      if (!image || !this.$refs.root) {
        this.bounds = null;
        return;
      }

      const imageRect = image.getBoundingClientRect();
      const rootRect = this.$refs.root.getBoundingClientRect();

      if (imageRect.width <= 0 || imageRect.height <= 0) {
        this.bounds = null;
        return;
      }

      const bounds = {
        left: imageRect.left - rootRect.left,
        top: imageRect.top - rootRect.top,
        width: imageRect.width,
        height: imageRect.height,
      };

      if (
        this.bounds &&
        this.bounds.left === bounds.left &&
        this.bounds.top === bounds.top &&
        this.bounds.width === bounds.width &&
        this.bounds.height === bounds.height
      ) {
        return;
      }

      this.bounds = bounds;
    },
    onPointerDown(event) {
      if (!this.isDrawMode) {
        return;
      }

      if (!this.bounds) {
        this.updateBounds();

        if (!this.bounds) {
          return;
        }
      }

      if (event.button !== undefined && event.button !== 0) {
        return;
      }

      const local = this.toLocal(event.clientX, event.clientY);

      if (!this.insideBounds(local)) {
        return;
      }

      if (this.pending) {
        const corner = this.hitTestCorner(local, this.pending);

        if (corner) {
          this.beginResize(corner, event);
          return;
        }

        if (this.insidePending(local, this.pending)) {
          this.beginMove(local, event);
          return;
        }
      }

      this.stopEvent(event);
      this.pending = null;
      this.interaction = "draw";
      this.pointerId = event.pointerId;
      this.dragStart = { local };
      this.draft = { x: local.x, y: local.y, w: 0, h: 0 };
      this.attachWindowPointerListeners();
    },
    onPointerMove(event) {
      if (!this.interaction || !this.dragStart || !this.bounds) {
        return;
      }

      if (this.pointerId !== null && event.pointerId !== this.pointerId) {
        return;
      }

      const local = this.toLocal(event.clientX, event.clientY);
      const currentX = Math.max(0, Math.min(this.bounds.width, local.x));
      const currentY = Math.max(0, Math.min(this.bounds.height, local.y));

      if (this.interaction === "move") {
        const origin = this.dragStart.pending;

        if (!origin) {
          return;
        }

        const deltaX = local.x - this.dragStart.local.x;
        const deltaY = local.y - this.dragStart.local.y;
        let nextX = origin.x + deltaX;
        let nextY = origin.y + deltaY;

        if (nextX < 0) {
          nextX = 0;
        }

        if (nextY < 0) {
          nextY = 0;
        }

        if (nextX + origin.w > this.bounds.width) {
          nextX = this.bounds.width - origin.w;
        }

        if (nextY + origin.h > this.bounds.height) {
          nextY = this.bounds.height - origin.h;
        }

        this.pending = { x: nextX, y: nextY, w: origin.w, h: origin.h };
        return;
      }

      const deltaX = currentX - this.dragStart.local.x;
      const deltaY = currentY - this.dragStart.local.y;
      const signX = deltaX < 0 ? -1 : 1;
      const signY = deltaY < 0 ? -1 : 1;
      let side = Math.max(Math.abs(deltaX), Math.abs(deltaY));

      if (this.interaction === "resize" && side < MIN_DRAW_SIZE) {
        side = MIN_DRAW_SIZE;
      }

      let startX = this.dragStart.local.x;
      let startY = this.dragStart.local.y;
      let width = side;
      let height = side;

      if (signX < 0) {
        startX = this.dragStart.local.x - side;
      }

      if (signY < 0) {
        startY = this.dragStart.local.y - side;
      }

      if (startX < 0) {
        width += startX;
        height += startX;
        startX = 0;
      }

      if (startY < 0) {
        width += startY;
        height += startY;
        startY = 0;
      }

      if (startX + width > this.bounds.width) {
        const overflow = startX + width - this.bounds.width;
        width -= overflow;
        height -= overflow;
      }

      if (startY + height > this.bounds.height) {
        const overflow = startY + height - this.bounds.height;
        width -= overflow;
        height -= overflow;
      }

      width = Math.max(0, width);
      height = Math.max(0, height);

      if (this.interaction === "resize") {
        this.pending = { x: startX, y: startY, w: width, h: height };
      } else {
        this.draft = { x: startX, y: startY, w: width, h: height };
      }
    },
    onPointerUp(event) {
      if (!this.interaction) {
        return;
      }

      if (this.pointerId !== null && event && event.pointerId !== this.pointerId) {
        return;
      }

      const interaction = this.interaction;
      const draft = this.draft;
      this.detachWindowPointerListeners();
      this.interaction = null;
      this.resizeCorner = null;
      this.pointerId = null;
      this.dragStart = null;
      this.draft = null;

      if (interaction !== "draw") {
        return;
      }

      if (!draft || !this.bounds || draft.w < MIN_DRAW_SIZE || draft.h < MIN_DRAW_SIZE) {
        return;
      }

      this.pending = draft;
    },
    onConfirmPending() {
      if (this.busy || !this.pending || !this.bounds) {
        return;
      }

      this.$emit("create", {
        X: this.clamp01(this.pending.x / this.bounds.width),
        Y: this.clamp01(this.pending.y / this.bounds.height),
        W: this.clamp01(this.pending.w / this.bounds.width),
        H: this.clamp01(this.pending.h / this.bounds.height),
      });
    },
    onCancelPending() {
      this.pending = null;
      this.hoverCursor = null;
    },
    clearPending() {
      this.pending = null;
      this.hoverCursor = null;
    },
    cancelActiveDraft() {
      if (this.interaction) {
        this.detachWindowPointerListeners();
      }

      this.interaction = null;
      this.resizeCorner = null;
      this.pointerId = null;
      this.dragStart = null;
      this.draft = null;
      this.pending = null;
      this.hoverCursor = null;
    },
    onKeyDown(event) {
      if (event.key !== "Escape") {
        return;
      }

      if (this.interaction === "draw") {
        this.interaction = null;
        this.pointerId = null;
        this.dragStart = null;
        this.draft = null;
        this.detachWindowPointerListeners();
        return;
      }

      if (this.interaction === "move" || this.interaction === "resize") {
        if (this.dragStart?.pending) {
          this.pending = { ...this.dragStart.pending };
        }

        this.interaction = null;
        this.resizeCorner = null;
        this.pointerId = null;
        this.dragStart = null;
        this.detachWindowPointerListeners();
        return;
      }

      if (this.pending) {
        this.pending = null;
        return;
      }

      this.$emit("cancel");
    },
    stopEvent(event) {
      if (typeof event.stopPropagation === "function") {
        event.stopPropagation();
      }

      if (typeof event.preventDefault === "function" && event.cancelable !== false) {
        event.preventDefault();
      }
    },
    attachWindowPointerListeners() {
      window.addEventListener("pointermove", this.onPointerMove);
      window.addEventListener("pointerup", this.onPointerUp);
      window.addEventListener("pointercancel", this.onPointerUp);
    },
    detachWindowPointerListeners() {
      window.removeEventListener("pointermove", this.onPointerMove);
      window.removeEventListener("pointerup", this.onPointerUp);
      window.removeEventListener("pointercancel", this.onPointerUp);
    },
    hitTestCorner(point, rect) {
      const radius = 14;
      const corners = {
        tl: { x: rect.x, y: rect.y },
        tr: { x: rect.x + rect.w, y: rect.y },
        bl: { x: rect.x, y: rect.y + rect.h },
        br: { x: rect.x + rect.w, y: rect.y + rect.h },
      };

      for (const key of Object.keys(corners)) {
        const corner = corners[key];

        if (Math.hypot(point.x - corner.x, point.y - corner.y) <= radius) {
          return key;
        }
      }

      return null;
    },
    insidePending(point, rect) {
      return point.x >= rect.x && point.y >= rect.y && point.x <= rect.x + rect.w && point.y <= rect.y + rect.h;
    },
    beginResize(corner, event) {
      if (!this.pending) {
        return;
      }

      let anchor = { x: this.pending.x, y: this.pending.y };

      if (corner === "tl") {
        anchor = { x: this.pending.x + this.pending.w, y: this.pending.y + this.pending.h };
      } else if (corner === "tr") {
        anchor = { x: this.pending.x, y: this.pending.y + this.pending.h };
      } else if (corner === "bl") {
        anchor = { x: this.pending.x + this.pending.w, y: this.pending.y };
      }

      this.stopEvent(event);
      this.interaction = "resize";
      this.resizeCorner = corner;
      this.pointerId = event.pointerId;
      this.dragStart = {
        local: anchor,
        pending: { ...this.pending },
      };
      this.attachWindowPointerListeners();
    },
    onHoverMove(event) {
      if (!this.isDrawMode || this.interaction) {
        return;
      }

      if (!this.pending || !this.bounds) {
        this.hoverCursor = null;
        return;
      }

      const local = this.toLocal(event.clientX, event.clientY);

      if (!this.insideBounds(local)) {
        this.hoverCursor = null;
        return;
      }

      const corner = this.hitTestCorner(local, this.pending);

      if (corner) {
        this.hoverCursor = corner === "tl" || corner === "br" ? "nwse-resize" : "nesw-resize";
        return;
      }

      if (this.insidePending(local, this.pending)) {
        this.hoverCursor = "move";
        return;
      }

      this.hoverCursor = null;
    },
    onHoverLeave() {
      this.hoverCursor = null;
    },
    beginMove(local, event) {
      if (!this.pending) {
        return;
      }

      this.stopEvent(event);
      this.interaction = "move";
      this.resizeCorner = null;
      this.pointerId = event.pointerId;
      this.dragStart = {
        local,
        pending: { ...this.pending },
      };
      this.attachWindowPointerListeners();
    },
    toLocal(clientX, clientY) {
      if (!this.bounds || !this.$refs.root) {
        return { x: 0, y: 0 };
      }

      const rootRect = this.$refs.root.getBoundingClientRect();

      return {
        x: clientX - rootRect.left - this.bounds.left,
        y: clientY - rootRect.top - this.bounds.top,
      };
    },
    insideBounds(point) {
      return this.bounds && point.x >= 0 && point.y >= 0 && point.x <= this.bounds.width && point.y <= this.bounds.height;
    },
    clamp01(value) {
      if (value < 0) {
        return 0;
      }

      if (value >= 1) {
        return 0.999999;
      }

      return value;
    },
  },
};
</script>

<style scoped>
.p-face-markers {
  position: absolute;
  inset: 0;
  z-index: 2;
  user-select: none;
  pointer-events: none;
  touch-action: auto;
}

.p-face-markers.is-drawing {
  pointer-events: auto;
  touch-action: none;
}

.p-face-markers__svg {
  pointer-events: none;
  overflow: visible;
}

.p-face-markers__rect {
  fill: rgba(255, 255, 255, 0.06);
  stroke: rgba(255, 255, 255, 0.9);
  stroke-width: 2px;
  vector-effect: non-scaling-stroke;
  pointer-events: none;
}

.p-face-markers__rect--named {
  stroke: rgba(var(--v-theme-info), 0.98);
}

.p-face-markers__rect--draft {
  stroke: rgba(var(--v-theme-info), 1);
  stroke-dasharray: 6 4;
  fill: rgba(var(--v-theme-info), 0.14);
}

.p-face-markers__handle {
  fill: rgba(var(--v-theme-info), 1);
  stroke: rgba(0, 0, 0, 0.9);
  stroke-width: 1.5px;
  vector-effect: non-scaling-stroke;
}

.p-face-markers__label {
  font-size: 13px;
  font-weight: 600;
  fill: #fff;
  stroke: rgba(0, 0, 0, 0.85);
  stroke-width: 3px;
  paint-order: stroke fill;
  pointer-events: none;
  dominant-baseline: hanging;
}

.p-face-markers__confirm {
  display: flex;
  gap: 8px;
  pointer-events: auto;
  z-index: 3;
}

.p-face-markers__btn {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.45);
}

.p-face-markers__btn--confirm {
  background: rgba(var(--v-theme-info), 0.95);
}

.p-face-markers__btn--cancel {
  background: rgba(90, 90, 90, 0.95);
}

.p-face-markers__btn.is-disabled,
.p-face-markers__btn[disabled] {
  opacity: 0.6;
  cursor: default;
}
</style>
