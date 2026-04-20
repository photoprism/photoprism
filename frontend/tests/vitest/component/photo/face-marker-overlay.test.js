import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { nextTick } from "vue";
import PFaceMarkerOverlay from "component/photo/face-marker-overlay.vue";

// Stub image bounds on the page.
const IMAGE_RECT = { left: 100, top: 50, width: 400, height: 300, right: 500, bottom: 350 };
const ROOT_RECT = { left: 0, top: 0, width: 800, height: 600, right: 800, bottom: 600 };

function createPswpStub() {
  // Use a real <img> so `instanceof HTMLImageElement` works.
  const img = document.createElement("img");
  img.getBoundingClientRect = () => IMAGE_RECT;
  const listeners = {};
  return {
    img,
    currSlide: { content: { element: img } },
    on: vi.fn((name, fn) => {
      listeners[name] = fn;
    }),
    off: vi.fn((name) => {
      delete listeners[name];
    }),
    _listeners: listeners,
  };
}

function mountOverlay(props = {}, listeners = {}) {
  const pswp = createPswpStub();
  const wrapper = mount(PFaceMarkerOverlay, {
    props: {
      markers: [],
      pswp,
      // Default to draw mode so the existing pointer-driven tests work
      // unchanged. Display-mode tests pass `mode: "display"` explicitly.
      mode: "draw",
      ...props,
      // Forward listeners as props so `wrapper.emitted()` limitations
      // don't matter — the component calls `this.$emit(name, ...)`, which
      // Vue then dispatches to the `on<Event>` prop provided here.
      onCreate: listeners.onCreate || (() => {}),
      onCancel: listeners.onCancel || (() => {}),
    },
    attachTo: document.body,
  });
  // Stub the root element bounding rect so toLocal math is predictable.
  wrapper.vm.$refs.root.getBoundingClientRect = () => ROOT_RECT;
  // Manually populate bounds (the real update path uses requestAnimationFrame).
  wrapper.vm.updateBounds();
  return { wrapper, pswp };
}

describe("PFaceMarkerOverlay", () => {
  beforeEach(() => {
    // Ensure rAF runs synchronously for predictable tests.
    vi.stubGlobal("requestAnimationFrame", (cb) => {
      cb();
      return 1;
    });
    vi.stubGlobal("cancelAnimationFrame", () => {});
  });

  it("mounts with image bounds derived from the PhotoSwipe slide", () => {
    const { wrapper } = mountOverlay();
    expect(wrapper.vm.bounds).toEqual({
      left: IMAGE_RECT.left - ROOT_RECT.left,
      top: IMAGE_RECT.top - ROOT_RECT.top,
      width: IMAGE_RECT.width,
      height: IMAGE_RECT.height,
    });
  });

  it("attaches and detaches PhotoSwipe event listeners", () => {
    const { wrapper, pswp } = mountOverlay();
    expect(pswp.on).toHaveBeenCalledWith("zoomPanUpdate", expect.any(Function));
    expect(pswp.on).toHaveBeenCalledWith("change", expect.any(Function));
    expect(pswp.on).toHaveBeenCalledWith("resize", expect.any(Function));
    wrapper.unmount();
    expect(pswp.off).toHaveBeenCalledWith("zoomPanUpdate", expect.any(Function));
    expect(pswp.off).toHaveBeenCalledWith("change", expect.any(Function));
    expect(pswp.off).toHaveBeenCalledWith("resize", expect.any(Function));
  });

  it("renders a square rect for each existing marker at the correct pixel size", async () => {
    const markers = [
      { UID: "m1", Name: "Jane", X: 0.25, Y: 0.1, W: 0.2, H: 0.2 },
      { UID: "m2", Name: "", X: 0.6, Y: 0.5, W: 0.1, H: 0.1 },
    ];
    const { wrapper } = mountOverlay({ markers });
    await nextTick();
    await flushPromises();
    // SVG elements don't use class selectors reliably in Vue Test Utils,
    // so look them up through the DOM.
    const rects = wrapper.element.querySelectorAll("rect");
    // 2 marker rects; the draft rect only renders during drawing.
    expect(rects.length).toBe(2);
    expect(rects[0].getAttribute("x")).toBe(String(0.25 * IMAGE_RECT.width));
    expect(rects[0].getAttribute("y")).toBe(String(0.1 * IMAGE_RECT.height));
    expect(rects[0].getAttribute("width")).toBe(String(0.2 * IMAGE_RECT.width));
    expect(rects[0].getAttribute("height")).toBe(String(0.2 * IMAGE_RECT.height));
  });

  it("renders a <title> element for named markers (hover tooltip)", async () => {
    const markers = [
      { UID: "m1", Name: "Jane Doe", X: 0.1, Y: 0.1, W: 0.2, H: 0.2 },
      { UID: "m2", Name: "", X: 0.5, Y: 0.5, W: 0.1, H: 0.1 },
    ];
    const { wrapper } = mountOverlay({ markers });
    await nextTick();
    await flushPromises();
    const rects = wrapper.element.querySelectorAll("rect");
    const namedTitle = rects[0].querySelector("title");
    const unnamedTitle = rects[1].querySelector("title");
    expect(namedTitle).not.toBeNull();
    expect(namedTitle.textContent).toBe("Jane Doe");
    expect(unnamedTitle).toBeNull();
  });

  it("highlights named markers with the 'named' modifier", async () => {
    const markers = [
      { UID: "m1", Name: "Jane", X: 0.1, Y: 0.1, W: 0.2, H: 0.2 },
      { UID: "m2", Name: "", X: 0.5, Y: 0.5, W: 0.1, H: 0.1 },
    ];
    const { wrapper } = mountOverlay({ markers });
    await nextTick();
    await flushPromises();
    const rects = wrapper.element.querySelectorAll("rect");
    expect(rects[0].classList.contains("p-face-markers__rect--named")).toBe(true);
    expect(rects[1].classList.contains("p-face-markers__rect--named")).toBe(false);
  });

  it("commits a draft as pending on pointer up without emitting yet", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    // Start the drag at (150, 100) viewport → image-local (50, 50).
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 1,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBe("draw");

    // Drag to (280, 180) viewport → image-local (180, 130). dx=130 dy=80 → side 130.
    wrapper.vm.onPointerMove({ pointerId: 1, clientX: 280, clientY: 180 });
    expect(wrapper.vm.draft).toMatchObject({ x: 50, y: 50, w: 130, h: 130 });

    wrapper.vm.onPointerUp({ pointerId: 1 });
    // Pointer up stores the draft as "pending"; no create event is emitted
    // until the user clicks the overlay confirm button.
    expect(onCreate).not.toHaveBeenCalled();
    expect(wrapper.vm.interaction).toBeNull();
    expect(wrapper.vm.draft).toBeNull();
    expect(wrapper.vm.pending).toMatchObject({ x: 50, y: 50, w: 130, h: 130 });
  });

  it("emits a normalized square marker only after explicit confirmation", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 10,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 10, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 10 });

    wrapper.vm.onConfirmPending();
    expect(onCreate).toHaveBeenCalledTimes(1);
    const area = onCreate.mock.calls[0][0];
    expect(area.X).toBeCloseTo(50 / IMAGE_RECT.width, 6);
    expect(area.Y).toBeCloseTo(50 / IMAGE_RECT.height, 6);
    expect(area.W).toBeCloseTo(130 / IMAGE_RECT.width, 6);
    expect(area.H).toBeCloseTo(130 / IMAGE_RECT.height, 6);
    // The parent (lightbox) is responsible for clearing the pending draft
    // via clearPending() after a successful save. The overlay itself must
    // keep it visible so a failed save can be retried.
    expect(wrapper.vm.pending).not.toBeNull();
  });

  it("clears the pending draft when the parent calls clearPending (success path)", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 40,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 40, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 40 });
    wrapper.vm.onConfirmPending();
    expect(wrapper.vm.pending).not.toBeNull();

    wrapper.vm.clearPending();
    expect(wrapper.vm.pending).toBeNull();
  });

  it("keeps the pending draft visible after confirm so the user can retry on save failure", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 41,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 41, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 41 });
    wrapper.vm.onConfirmPending();

    // Simulate the parent receiving a save failure: it never calls
    // clearPending(), so the pending rect must remain on screen.
    expect(onCreate).toHaveBeenCalledTimes(1);
    expect(wrapper.vm.pending).not.toBeNull();
    expect(wrapper.vm.activeDraft).toBe(wrapper.vm.pending);
  });

  it("discards the pending draft without emitting on cancel", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 11,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 11, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 11 });
    expect(wrapper.vm.pending).not.toBeNull();

    wrapper.vm.onCancelPending();
    expect(wrapper.vm.pending).toBeNull();
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("replaces an unconfirmed pending draft when the user starts a new drag outside it", () => {
    const { wrapper } = mountOverlay();
    // Draw a pending rect at image-local (50, 50, 130, 130).
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 12,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 12, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 12 });
    expect(wrapper.vm.pending).not.toBeNull();

    // Pointerdown far outside the pending body/corners (local 300, 270),
    // which must start a fresh draw and clear the previous pending.
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 13,
      clientX: 400,
      clientY: 320,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.pending).toBeNull();
    expect(wrapper.vm.interaction).toBe("draw");
  });

  it("discards a pending draft on Escape without exiting the mode", () => {
    const onCancel = vi.fn();
    const { wrapper } = mountOverlay({}, { onCancel });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 14,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 14, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 14 });
    expect(wrapper.vm.pending).not.toBeNull();

    wrapper.vm.onKeyDown({ key: "Escape" });
    expect(wrapper.vm.pending).toBeNull();
    expect(onCancel).not.toHaveBeenCalled();
  });

  it("enforces squareness during drag (equal screen-pixel width and height)", () => {
    const { wrapper } = mountOverlay();
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 2,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    // dx=50, dy=20 → side=50
    wrapper.vm.onPointerMove({ pointerId: 2, clientX: 250, clientY: 170 });
    expect(wrapper.vm.draft.w).toBe(wrapper.vm.draft.h);
  });

  it("ignores drags smaller than the minimum draw size", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 3,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    // Tiny move.
    wrapper.vm.onPointerMove({ pointerId: 3, clientX: 205, clientY: 155 });
    wrapper.vm.onPointerUp({ pointerId: 3 });
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("ignores pointer downs outside the image bounds", () => {
    const { wrapper } = mountOverlay();
    // clientX=10 → local.x = 10 - 0 - 100 = -90 (outside).
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 4,
      clientX: 10,
      clientY: 10,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBeNull();
  });

  it("ignores non-left mouse buttons", () => {
    const { wrapper } = mountOverlay();
    wrapper.vm.onPointerDown({
      button: 2,
      pointerId: 5,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBeNull();
  });

  it("cancels an in-progress drag on Escape", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({}, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 6,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 6, clientX: 300, clientY: 250 });
    expect(wrapper.vm.interaction).toBe("draw");

    wrapper.vm.onKeyDown({ key: "Escape" });
    expect(wrapper.vm.interaction).toBeNull();
    expect(wrapper.vm.draft).toBeNull();
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("emits cancel on Escape when not drafting", () => {
    const onCancel = vi.fn();
    const { wrapper } = mountOverlay({}, { onCancel });
    wrapper.vm.onKeyDown({ key: "Escape" });
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("clamps the drawn square inside the image bounds", () => {
    const { wrapper } = mountOverlay();
    // Start near bottom-right (client 480, 340 → local 380, 290).
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 7,
      clientX: 480,
      clientY: 340,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    // Drag far past bottom-right.
    wrapper.vm.onPointerMove({ pointerId: 7, clientX: 700, clientY: 500 });
    const draft = wrapper.vm.draft;
    expect(draft.x + draft.w).toBeLessThanOrEqual(IMAGE_RECT.width + 0.0001);
    expect(draft.y + draft.h).toBeLessThanOrEqual(IMAGE_RECT.height + 0.0001);
    expect(draft.w).toBe(draft.h);
  });

  // Display mode — the overlay renders existing markers but does not
  // capture the pointer, so PhotoSwipe pan/zoom keeps working.
  it("does not start drafting on pointer down in display mode", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({ mode: "display" }, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 1,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBeNull();
    expect(wrapper.vm.draft).toBeNull();
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("renders existing markers in display mode", async () => {
    const markers = [{ UID: "m1", Name: "Jane", X: 0.1, Y: 0.1, W: 0.2, H: 0.2 }];
    const { wrapper } = mountOverlay({ mode: "display", markers });
    await nextTick();
    await flushPromises();
    const rects = wrapper.element.querySelectorAll("rect");
    expect(rects.length).toBe(1);
  });

  it("renders a visible name label below each named marker in display mode", async () => {
    const markers = [
      { UID: "m1", Name: "Jane", X: 0.1, Y: 0.1, W: 0.2, H: 0.2 },
      { UID: "m2", Name: "", X: 0.5, Y: 0.5, W: 0.1, H: 0.1 },
    ];
    const { wrapper } = mountOverlay({ mode: "display", markers });
    await nextTick();
    await flushPromises();
    const labels = wrapper.element.querySelectorAll("text.p-face-markers__label");
    expect(labels.length).toBe(1);
    expect(labels[0].textContent).toBe("Jane");
    // Label is positioned below the rect (y > rect's y + height).
    const labelY = parseFloat(labels[0].getAttribute("y"));
    const rectBottom = (0.1 + 0.2) * IMAGE_RECT.height;
    expect(labelY).toBeGreaterThan(rectBottom);
  });

  it("does not render visible name labels while in draw mode", async () => {
    const markers = [{ UID: "m1", Name: "Jane", X: 0.1, Y: 0.1, W: 0.2, H: 0.2 }];
    const { wrapper } = mountOverlay({ mode: "draw", markers });
    await nextTick();
    await flushPromises();
    const labels = wrapper.element.querySelectorAll("text.p-face-markers__label");
    expect(labels.length).toBe(0);
  });

  it("discards an active draft when the mode changes from draw to display", async () => {
    const { wrapper } = mountOverlay();
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 21,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 21, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 21 });
    expect(wrapper.vm.pending).not.toBeNull();

    await wrapper.setProps({ mode: "display" });
    expect(wrapper.vm.pending).toBeNull();
    expect(wrapper.vm.draft).toBeNull();
    expect(wrapper.vm.interaction).toBeNull();
  });

  it("does not emit create from confirm when busy is true", () => {
    const onCreate = vi.fn();
    const { wrapper } = mountOverlay({ busy: true }, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 31,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 31, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 31 });
    wrapper.vm.onConfirmPending();
    expect(onCreate).not.toHaveBeenCalled();
    // Pending stays put so the user can retry once the request settles.
    expect(wrapper.vm.pending).not.toBeNull();
  });

  // ------------------------------------------------------------------
  // Pending reposition (body drag) and corner resize.
  // The feature brief Add UC Main Scenario Step 4 requires: "the user
  // repositions the marker by dragging its body and resizes it by
  // dragging its corners; the marker MUST remain square at all times".
  // ------------------------------------------------------------------

  // Small helper to drop a pending rect at a known image-local position.
  // Returns `{ wrapper, pending }` where pending is { x, y, w, h }.
  function drawPending(extraProps = {}) {
    const { wrapper } = mountOverlay(extraProps);
    // client (150, 100) → local (50, 50) start; client (280, 180) → local (180, 130), dx=130 dy=80 → side 130.
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 100,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 100, clientX: 280, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 100 });
    return { wrapper, pending: { ...wrapper.vm.pending } };
  }

  it("repositions the pending rect by dragging its body (size unchanged)", () => {
    const { wrapper, pending } = drawPending();
    // Start a move from inside the pending body (local ~100, 100 → client 200, 150).
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 101,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBe("move");
    // Drag +30 in both axes → client (230, 180).
    wrapper.vm.onPointerMove({ pointerId: 101, clientX: 230, clientY: 180 });
    wrapper.vm.onPointerUp({ pointerId: 101 });

    expect(wrapper.vm.pending.x).toBeCloseTo(pending.x + 30, 6);
    expect(wrapper.vm.pending.y).toBeCloseTo(pending.y + 30, 6);
    expect(wrapper.vm.pending.w).toBe(pending.w);
    expect(wrapper.vm.pending.h).toBe(pending.h);
    expect(wrapper.vm.interaction).toBeNull();
  });

  it("clamps a body-drag inside the image bounds and preserves size", () => {
    const { wrapper, pending } = drawPending();
    // Grab the body and drag far past the bottom-right.
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 102,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 102, clientX: 900, clientY: 900 });
    wrapper.vm.onPointerUp({ pointerId: 102 });

    // Width/height unchanged, rect fully inside image bounds.
    expect(wrapper.vm.pending.w).toBe(pending.w);
    expect(wrapper.vm.pending.h).toBe(pending.h);
    expect(wrapper.vm.pending.x + wrapper.vm.pending.w).toBeLessThanOrEqual(IMAGE_RECT.width + 0.0001);
    expect(wrapper.vm.pending.y + wrapper.vm.pending.h).toBeLessThanOrEqual(IMAGE_RECT.height + 0.0001);
    expect(wrapper.vm.pending.x).toBe(IMAGE_RECT.width - pending.w);
    expect(wrapper.vm.pending.y).toBe(IMAGE_RECT.height - pending.h);
  });

  it("resizes from the br corner with the tl corner as a fixed anchor and stays square", () => {
    const { wrapper, pending } = drawPending();
    // br corner is at image-local (180, 180) → client (280, 230).
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 103,
      clientX: 280,
      clientY: 230,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBe("resize");
    expect(wrapper.vm.resizeCorner).toBe("br");
    // Drag br outward to client (350, 300) → local (250, 250). Anchor tl is at (50, 50),
    // so dx = 200, dy = 200 → side = 200 → new rect (50, 50, 200, 200).
    wrapper.vm.onPointerMove({ pointerId: 103, clientX: 350, clientY: 300 });
    wrapper.vm.onPointerUp({ pointerId: 103 });

    expect(wrapper.vm.pending.x).toBeCloseTo(pending.x, 6);
    expect(wrapper.vm.pending.y).toBeCloseTo(pending.y, 6);
    expect(wrapper.vm.pending.w).toBe(wrapper.vm.pending.h);
    expect(wrapper.vm.pending.w).toBeCloseTo(200, 6);
  });

  it("resizes from the tl corner with the br corner as a fixed anchor", () => {
    const { wrapper, pending } = drawPending();
    // tl corner is at image-local (50, 50) → client (150, 100). Drag inward to shrink.
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 104,
      clientX: 150,
      clientY: 100,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBe("resize");
    expect(wrapper.vm.resizeCorner).toBe("tl");
    // Move tl to local (120, 120) → client (220, 170). Anchor is br at local (180, 180).
    // dx = 120 - 180 = -60, dy = 120 - 180 = -60 → side 60.
    // Result: x = 180 - 60 = 120, y = 180 - 60 = 120, w = h = 60.
    wrapper.vm.onPointerMove({ pointerId: 104, clientX: 220, clientY: 170 });
    wrapper.vm.onPointerUp({ pointerId: 104 });

    // br corner stays fixed.
    expect(wrapper.vm.pending.x + wrapper.vm.pending.w).toBeCloseTo(pending.x + pending.w, 6);
    expect(wrapper.vm.pending.y + wrapper.vm.pending.h).toBeCloseTo(pending.y + pending.h, 6);
    expect(wrapper.vm.pending.w).toBe(wrapper.vm.pending.h);
    expect(wrapper.vm.pending.w).toBeCloseTo(60, 6);
  });

  it("prioritizes a corner hit over body move when the pointer lands near a corner", () => {
    const { wrapper } = drawPending();
    // tr corner is at local (180, 50) → client (280, 100). Click 4 px inside the body
    // (client 277, 103 → local 177, 53) — still within the 14 px hit radius of tr.
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 105,
      clientX: 277,
      clientY: 103,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    expect(wrapper.vm.interaction).toBe("resize");
    expect(wrapper.vm.resizeCorner).toBe("tr");
  });

  it("reverts pending to the pre-interaction snapshot when Escape is pressed during a resize", () => {
    const { wrapper, pending } = drawPending();
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 106,
      clientX: 280,
      clientY: 230, // br corner
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 106, clientX: 360, clientY: 260 });
    // Pending was mutated mid-resize.
    expect(wrapper.vm.pending.w).not.toBe(pending.w);

    wrapper.vm.onKeyDown({ key: "Escape" });

    expect(wrapper.vm.pending).toEqual(pending);
    expect(wrapper.vm.interaction).toBeNull();
    expect(wrapper.vm.resizeCorner).toBeNull();
  });

  it("reverts pending when Escape is pressed during a body move", () => {
    const { wrapper, pending } = drawPending();
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 107,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 107, clientX: 260, clientY: 200 });
    expect(wrapper.vm.pending.x).not.toBe(pending.x);

    wrapper.vm.onKeyDown({ key: "Escape" });

    expect(wrapper.vm.pending).toEqual(pending);
    expect(wrapper.vm.interaction).toBeNull();
  });

  it("hides the confirm buttons during an active manipulation and shows them after", async () => {
    const { wrapper } = drawPending();
    await nextTick();
    // Before the drag, confirm group is rendered.
    expect(wrapper.element.querySelector(".p-face-markers__confirm")).not.toBeNull();

    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 108,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    await nextTick();
    // During the move, the confirm group is hidden.
    expect(wrapper.element.querySelector(".p-face-markers__confirm")).toBeNull();

    wrapper.vm.onPointerMove({ pointerId: 108, clientX: 220, clientY: 170 });
    wrapper.vm.onPointerUp({ pointerId: 108 });
    await nextTick();
    // After release the confirm group is back.
    expect(wrapper.element.querySelector(".p-face-markers__confirm")).not.toBeNull();
  });

  it("enforces the minimum square side during a corner resize", () => {
    const { wrapper, pending } = drawPending();
    // Grab the br corner and collapse it back toward the tl anchor past zero.
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 109,
      clientX: 280, // br corner
      clientY: 230,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 109, clientX: 140, clientY: 90 }); // local (40, 40) — past anchor
    wrapper.vm.onPointerUp({ pointerId: 109 });

    // The square must not collapse below MIN_DRAW_SIZE (16). Anchor was tl at (50, 50).
    expect(wrapper.vm.pending.w).toBe(wrapper.vm.pending.h);
    expect(wrapper.vm.pending.w).toBeGreaterThanOrEqual(16);
    // The anchored tl corner should still sit at (50, 50).
    expect(pending.x).toBe(50);
    expect(pending.y).toBe(50);
  });

  // Role gating — a full pointer cycle in display mode must never emit
  // `create`, and the draft confirm/cancel buttons must not be mounted.
  // The complementary display-mode rendering and single-event tests
  // live above (lines 404 and 420).
  it("is inert for the full pointer cycle in display mode", async () => {
    const onCreate = vi.fn();
    const markers = [{ UID: "m1", Name: "Jane", X: 0.1, Y: 0.1, W: 0.2, H: 0.2 }];
    const { wrapper } = mountOverlay({ mode: "display", markers }, { onCreate });
    wrapper.vm.onPointerDown({
      button: 0,
      pointerId: 2,
      clientX: 200,
      clientY: 150,
      stopPropagation: () => {},
      preventDefault: () => {},
    });
    wrapper.vm.onPointerMove({ pointerId: 2, clientX: 360, clientY: 260 });
    wrapper.vm.onPointerUp({ pointerId: 2 });
    expect(wrapper.vm.pending).toBeNull();
    expect(wrapper.vm.draft).toBeNull();
    expect(wrapper.vm.interaction).toBeNull();
    expect(onCreate).not.toHaveBeenCalled();
    await nextTick();
    await flushPromises();
    expect(wrapper.element.querySelector("button.p-face-markers__btn--confirm")).toBeNull();
    expect(wrapper.element.querySelector("button.p-face-markers__btn--cancel")).toBeNull();
  });
});
