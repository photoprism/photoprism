import { describe, it, expect, beforeEach, vi } from "vitest";
import { $view, preventNavigationTouchEvent } from "common/view";
import { buildNamespace } from "common/storage";

// jsdom does not implement TouchEvent (PointerEvent is the modern path),
// but preventNavigationTouchEvent uses `ev instanceof TouchEvent` as its
// first guard. Polyfill a minimal subclass so the test can construct
// synthetic TouchEvents and exercise every branch of the handler.
if (typeof globalThis.TouchEvent === "undefined") {
  globalThis.TouchEvent = class TouchEvent extends Event {
    constructor(type, init = {}) {
      super(type, init);
    }
  };
}

// Builds a synthetic TouchEvent with the shape preventNavigationTouchEvent
// reads: `type`, `cancelable`, `touches[0].clientX/clientY`, and `target`.
// Wraps preventDefault in a vi.fn so we can count calls precisely (Event's
// own preventDefault just flips defaultPrevented; the call count is more
// expressive in assertions).
function makeTouchEvent({ type = "touchmove", cancelable = true, clientX = 0, clientY = 0, target = document.body, changedOnly = false } = {}) {
  const ev = new TouchEvent(type, { cancelable, bubbles: true });
  const touch = { clientX, clientY };
  if (changedOnly) {
    Object.defineProperty(ev, "touches", { value: [] });
    Object.defineProperty(ev, "changedTouches", { value: [touch] });
  } else {
    Object.defineProperty(ev, "touches", { value: [touch] });
    Object.defineProperty(ev, "changedTouches", { value: [touch] });
  }
  Object.defineProperty(ev, "target", { value: target });
  ev.preventDefault = vi.fn(() => {
    Object.defineProperty(ev, "defaultPrevented", { value: true, configurable: true });
  });
  return ev;
}

describe("common/view", () => {
  it("should return parent", () => {
    expect($view.getParent()).toBe(null);
  });
  it("should return parent name", () => {
    expect($view.getParentName()).toBe("");
  });
  it("should return data", () => {
    expect($view.getData()).toEqual({});
  });
  it("should return number of layers", () => {
    expect($view.len()).toBe(0);
  });
  it("should return if root view is active", () => {
    expect($view.isRoot()).toBe(true);
  });
  it("should return if view is app", () => {
    expect($view.isApp()).toBe(true);
  });

  describe("window scroll position helpers", () => {
    const storageKey = "window.scroll.pos";
    const namespacedStorageKey = buildNamespace(window.__CONFIG__?.storageNamespace) + storageKey;

    beforeEach(() => {
      localStorage.clear();
      delete window.positionToRestore;
    });

    it("saves and restores an explicit scroll position", () => {
      const pos = { left: 123, top: 456 };

      $view.saveWindowScrollPos(pos);

      expect(window.positionToRestore).toEqual(pos);
      expect(localStorage.getItem(namespacedStorageKey)).toEqual(JSON.stringify(pos));

      const restored = $view.getWindowScrollPos();
      expect(restored).toEqual(pos);
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });

    it("prefers in-memory value over localStorage", () => {
      const memoryPos = { left: 10, top: 20 };
      const storedPos = { left: 30, top: 40 };

      window.positionToRestore = memoryPos;
      localStorage.setItem(namespacedStorageKey, JSON.stringify(storedPos));

      const restored = $view.getWindowScrollPos();

      expect(restored).toEqual(memoryPos);
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });

    it("falls back to stored value when memory value is invalid", () => {
      window.positionToRestore = { left: Number.NaN, top: 1 };
      const storedPos = { left: 77, top: 88 };
      localStorage.setItem(namespacedStorageKey, JSON.stringify(storedPos));

      const restored = $view.getWindowScrollPos();

      expect(restored).toEqual(storedPos);
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });

    it("clears invalid stored data", () => {
      localStorage.setItem(namespacedStorageKey, "{invalid json");

      const restored = $view.getWindowScrollPos();

      expect(restored).toBeUndefined();
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });
  });

  // preventNavigationTouchEvent is the navigation-gesture suppressor wired
  // to `window` while the lightbox view is active. Its purpose is non-
  // obvious (it intercepts iOS swipe-back / pull-to-refresh / horizontal-
  // swipe-to-navigate so users can't accidentally close PhotoPrism while
  // using PhotoSwipe's pan/swipe gestures), so regressions are silent
  // until somebody reports broken scroll on iPad or a broken Back button.
  // The contract these tests pin:
  //
  //   1. Only TouchEvents that can be canceled are inspected at all.
  //   2. Only touchstart / touchmove are inspected.
  //   3. Touches inside the 30px L / R / T edge bands trigger preventDefault.
  //   4. Inner-area touches are left alone (sidebar / dialog / menu scroll).
  //   5. Touches on interactive widgets (button, input, textarea, select,
  //      anchored link, role=button) are exempt even inside the edge bands
  //      — so the overlay's top-left Back button and the sidebar's right-
  //      edge action buttons stay tap-reliable on iPad. The exemption
  //      walks up via `target.closest(...)` so taps that land on an inner
  //      <svg> / <i> glyph are exempt too.
  describe("preventNavigationTouchEvent", () => {
    const innerWidth = 800; // jsdom default
    const innerHeight = 600;

    beforeEach(() => {
      // Pin the viewport so edge-band math is deterministic.
      Object.defineProperty(window, "innerWidth", { configurable: true, value: innerWidth });
      Object.defineProperty(window, "innerHeight", { configurable: true, value: innerHeight });
    });

    it("is a no-op for non-TouchEvent inputs", () => {
      const ev = new Event("touchmove", { cancelable: true });
      ev.preventDefault = vi.fn();
      preventNavigationTouchEvent(ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });

    it("is a no-op when the event is not cancelable", () => {
      const ev = makeTouchEvent({ cancelable: false, clientX: 10, clientY: 200 });
      preventNavigationTouchEvent(ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });

    it("is a no-op for event types other than touchstart / touchmove", () => {
      const touchend = makeTouchEvent({ type: "touchend", clientX: 10, clientY: 200, changedOnly: true });
      const touchcancel = makeTouchEvent({ type: "touchcancel", clientX: 10, clientY: 200, changedOnly: true });
      preventNavigationTouchEvent(touchend);
      preventNavigationTouchEvent(touchcancel);
      expect(touchend.preventDefault).not.toHaveBeenCalled();
      expect(touchcancel.preventDefault).not.toHaveBeenCalled();
    });

    it("is a no-op when no touch point is available", () => {
      const ev = new TouchEvent("touchmove", { cancelable: true });
      Object.defineProperty(ev, "touches", { value: [] });
      Object.defineProperty(ev, "changedTouches", { value: [] });
      Object.defineProperty(ev, "target", { value: document.body });
      ev.preventDefault = vi.fn();
      preventNavigationTouchEvent(ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });

    it("leaves inner-area touches alone (sidebar / dialog / menu scroll)", () => {
      const ev = makeTouchEvent({ clientX: 200, clientY: 200 });
      preventNavigationTouchEvent(ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });

    describe("edge bands trigger preventDefault on bare canvas", () => {
      it("left edge", () => {
        const ev = makeTouchEvent({ clientX: 10, clientY: 200 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });

      it("right edge", () => {
        const ev = makeTouchEvent({ clientX: innerWidth - 10, clientY: 200 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });

      it("top edge", () => {
        const ev = makeTouchEvent({ clientX: 200, clientY: 10 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });

      it("touchstart and touchmove behave the same", () => {
        const start = makeTouchEvent({ type: "touchstart", clientX: 10, clientY: 200 });
        const move = makeTouchEvent({ type: "touchmove", clientX: 10, clientY: 200 });
        preventNavigationTouchEvent(start);
        preventNavigationTouchEvent(move);
        expect(start.preventDefault).toHaveBeenCalledTimes(1);
        expect(move.preventDefault).toHaveBeenCalledTimes(1);
      });

      // The bottom edge is intentionally NOT in the suppression band — iOS
      // uses the bottom region for system gestures (Home indicator) which
      // we don't try to block, and pull-to-refresh originates from the
      // top, not the bottom.
      it("bottom edge is NOT in the suppression band", () => {
        const ev = makeTouchEvent({ clientX: 200, clientY: innerHeight - 10 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });
    });

    describe("interactive targets at edges are exempted", () => {
      let host;
      beforeEach(() => {
        host = document.createElement("div");
        document.body.appendChild(host);
      });

      it("exempts <button> direct target", () => {
        const btn = document.createElement("button");
        host.appendChild(btn);
        const ev = makeTouchEvent({ clientX: 10, clientY: 200, target: btn });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });

      it("exempts nested target inside <button> (e.g. <svg> glyph)", () => {
        const btn = document.createElement("button");
        const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
        btn.appendChild(svg);
        host.appendChild(btn);
        const ev = makeTouchEvent({ clientX: 10, clientY: 200, target: svg });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });

      it("exempts <input>, <textarea>, <select>", () => {
        for (const tag of ["input", "textarea", "select"]) {
          const el = document.createElement(tag);
          host.appendChild(el);
          const ev = makeTouchEvent({ clientX: innerWidth - 10, clientY: 200, target: el });
          preventNavigationTouchEvent(ev);
          expect(ev.preventDefault).not.toHaveBeenCalled();
        }
      });

      it("exempts anchored link (<a href>)", () => {
        const a = document.createElement("a");
        a.setAttribute("href", "#x");
        host.appendChild(a);
        const ev = makeTouchEvent({ clientX: 10, clientY: 10, target: a });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });

      it("does NOT exempt bare <a> without an href (not navigable)", () => {
        const a = document.createElement("a");
        host.appendChild(a);
        const ev = makeTouchEvent({ clientX: 10, clientY: 200, target: a });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });

      it("exempts [role=button]", () => {
        const div = document.createElement("div");
        div.setAttribute("role", "button");
        host.appendChild(div);
        const ev = makeTouchEvent({ clientX: 10, clientY: 200, target: div });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });

      it("does NOT exempt non-interactive elements at edges", () => {
        const div = document.createElement("div");
        host.appendChild(div);
        const ev = makeTouchEvent({ clientX: 10, clientY: 200, target: div });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });
    });

    describe("edge-band boundaries", () => {
      // The bands are inclusive on the canvas side: clientX <= 30 is in
      // the left band, clientX >= innerWidth - 30 is in the right band.
      // Anything past 30 / before innerWidth-30 is inner area.
      it("clientX = 30 is band-inclusive (preventDefault fires)", () => {
        const ev = makeTouchEvent({ clientX: 30, clientY: 200 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });

      it("clientX = 31 is in the inner area (preventDefault does NOT fire)", () => {
        const ev = makeTouchEvent({ clientX: 31, clientY: 200 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });

      it("clientY = 30 is in the top band (preventDefault fires)", () => {
        const ev = makeTouchEvent({ clientX: 200, clientY: 30 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).toHaveBeenCalledTimes(1);
      });

      it("clientY = 31 is in the inner area (preventDefault does NOT fire)", () => {
        const ev = makeTouchEvent({ clientX: 200, clientY: 31 });
        preventNavigationTouchEvent(ev);
        expect(ev.preventDefault).not.toHaveBeenCalled();
      });
    });
  });
});
