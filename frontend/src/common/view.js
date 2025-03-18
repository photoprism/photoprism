import { toRaw } from "vue";
import $notify from "common/notify";

const TouchStartEvent = "touchstart";
const TouchMoveEvent = "touchmove";

// True if debug and/or trace logs should be recorded.
const debug = window.__CONFIG__?.debug;
const trace = window.__CONFIG__?.trace;

// Returns the <html> element.
export function getHtmlElement() {
  return document.documentElement;
}

// Initializes the <html> element by removing the "class" attribute.
export function initHtmlElement() {
  const htmlElement = document.documentElement;

  if (htmlElement && htmlElement.hasAttribute("class")) {
    if (debug) {
      console.log(`html: removed class="${htmlElement.getAttribute("class")}"`);
    }

    // Remove the class="loading" attribute from <html> when the application has loaded.
    htmlElement.removeAttribute("class");
    htmlElement.setAttribute("style", "");

    // If requested, hide the scrollbar permanently by adding class="hide-scrollbar" to <html>.
    if (document.body.classList.contains("hide-scrollbar")) {
      htmlElement.setAttribute("class", "hide-scrollbar");

      if (debug) {
        console.log('html: added class="hide-scrollbar" to permanently hide the scrollbar');
      }
    }
  }
}
// Set a :root style variable, or removes it if the value is empty.
export function setHtmlStyle(key, value) {
  if (!key) {
    return false;
  }

  const htmlElement = getHtmlElement();

  if (!htmlElement) {
    return false;
  } else if (value) {
    htmlElement.style.setProperty(key, value);
  } else {
    htmlElement.style.removeProperty(key);
  }

  return true;
}

// Returns the <body> element.
export function getBodyElement() {
  return document.body;
}

// Returns the width of the vertical window scrollbar.
export function getScrollbarWidth() {
  const body = getBodyElement();

  if (!body || !window.innerWidth) {
    return 0;
  }

  return window.innerWidth - body.offsetWidth;
}

// Checks if the element is a button.
export function isInputElement(el) {
  if (!el) {
    return false;
  }

  return el instanceof HTMLButtonElement;
}

// Checks if the element is an image, video, or canvas.
export function isMediaElement(el) {
  if (!el) {
    return false;
  }

  return el instanceof HTMLImageElement || el instanceof HTMLVideoElement || el instanceof HTMLCanvasElement;
}

// Component refs supported for automatic focus element detection.
const focusRefs = ["form", "content", "container", "dialog", "page"];

// Returns the most likely focus element for the given component, or null if none exists.
export function findFocusElement(c) {
  if (!c) {
    return null;
  }

  let el, ref;

  if (c.$refs && c.$refs instanceof Object) {
    focusRefs.forEach((r) => {
      if (c.$refs[r] && c.$refs[r] instanceof Object) {
        if (c.$refs[r].$el instanceof HTMLElement && c.$refs[r].$el.getAttribute("tabindex") !== null) {
          ref = c.$refs[r].$el;
        } else if (c.$refs[r] instanceof HTMLElement && c.$refs[r].getAttribute("tabindex") !== null) {
          ref = c.$refs[r];
        }
      }
    });
  }

  if (!ref || !(ref instanceof Object) || typeof ref.getAttribute !== "function") {
    ref = null;
  } else if (ref.getAttribute("tabindex") === null) {
    ref = null;
  }

  if (!ref && c.$el && c.$el instanceof Object) {
    if (c.$el instanceof HTMLElement) {
      ref = c.$el;
    } else if (c.$el.parentElement && c.$el.parentElement instanceof HTMLElement) {
      ref = c.$el.parentElement;
    }
  }

  if (ref) {
    if (ref.$el && ref.$el instanceof HTMLElement) {
      ref = ref.$el;
    }

    if (ref instanceof HTMLElement) {
      if (ref.getAttribute("tabindex") !== null) {
        return ref;
      }

      try {
        el = ref.querySelector('input[tabindex="1"]');
        if (el && el instanceof HTMLElement) {
          return el;
        }
      } catch (_) {
        // Ignore.
      }
    }
  }

  if (c.$refs?.dialog) {
    return document.querySelector(".v-overlay-container .v-overlay__content");
  }

  return null;
}

// Gives focus to the specified HTML element, or the first element that matches the specified selector string.
export function setFocus(el, selector, scroll) {
  if (!el) {
    return false;
  }

  let options = { preventScroll: !scroll };

  if (typeof el === "string") {
    el = document.querySelector(el);
  } else if (el instanceof Object) {
    if (!selector && typeof el.focus === "function") {
      try {
        el.focus(options);
        return true;
      } catch (err) {
        console.log(`failed to call el.focus(): ${err}`, el);
      }
    }

    if (el.$el && el.$el instanceof HTMLElement) {
      el = el.$el;
    }
  }

  if (el && el instanceof HTMLElement) {
    if (selector && typeof selector === "string") {
      el = el.querySelector(selector);

      if (!el || !(el instanceof HTMLElement)) {
        return false;
      }
    }

    if (trace) {
      console.log("giving focus to this element:", el);
    }

    try {
      el.focus(options);
      return true;
    } catch (err) {
      console.log(`failed to give focus to element: ${err}`, el);
    }
  } else if (trace) {
    console.log("invalid focus element:", el);
  }

  return false;
}

// Prevents the default navigation touch gestures.
export function preventNavigationTouchEvent(ev) {
  if (ev instanceof TouchEvent && ev.cancelable) {
    // console.log(`${ev.type} @ ${ev.touches[0].clientX.toString()} x ${ev.touches[0].clientY.toString()}`, ev.target);
    if (ev.type === TouchStartEvent && (isMediaElement(ev.target) || ev.touches[0].clientX <= 30)) {
      if (window.innerHeight - ev.touches[0].clientY > 128 || ev.touches[0].clientX <= 30) {
        ev.preventDefault();
        // console.log(`prevented ${ev.type} @ ${ev.touches[0].clientX.toString()} x ${ev.touches[0].clientY.toString()}`);
      }
    } else if (ev.type === TouchMoveEvent && !isInputElement(ev.target)) {
      ev.preventDefault();
      // console.log(`prevented ${ev.type} @ ${ev.touches[0].clientX.toString()} x ${ev.touches[0].clientY.toString()}`);
    }
  }
}

// Returns a random string that can be used as an identifier.
export function generateRandomId() {
  return Date.now().toString(36) + Math.random().toString(36).substring(2, 18);
}

// View keeps track of the visible components and dialogs,
// and updates the window and <html> body as needed.
export class View {
  // Initializes the instance properties with the default values.
  constructor() {
    this.uid = 0;
    this.scopes = [];
    this.hideScrollbar = false;
    this.preventNavigation = false;

    if (trace) {
      document.addEventListener("focusin", (ev) => {
        console.log("%cdocument.focusin", "color: #B2EBF2;", ev.target);
      });
    }
  }

  // Changes the view context to the specified component,
  // and updates the window and <html> body as needed.
  enter(c, focusElement, focusSelector) {
    if (!c) {
      return false;
    }

    if (this.isRoot()) {
      initHtmlElement();
    }

    if (c !== this.current()) {
      this.scopes.push(c);
    }

    this.apply(c, focusElement, focusSelector);

    return this.scopes.length;
  }

  // Returns to the parent view context of the specified component,
  // and updates the window and <html> body as needed.
  leave(c) {
    if (!c || this.scopes.length === 0) {
      return false;
    }

    const index = this.scopes.findLastIndex((s) => s === c);

    if (index > 0) {
      this.scopes.splice(index, 1);
    } else if (index < 0) {
      return;
    }

    if (this.scopes.length) {
      this.apply(this.scopes[this.scopes.length - 1]);
    }

    return this.scopes.length;
  }

  // Updates the window and the <html> body elements based on the specified component.
  apply(c, focusElement, focusSelector) {
    if (!c || typeof c !== "object" || !Number.isInteger(c?.$?.uid) || !c.$el) {
      console.log(`view: invalid component (#${this.uid.toString()})`, c);
      return false;
    }

    // Get the component's name and numeric ID.
    const name = c?.$options?.name ? c.$options.name : "";
    const uid = c.$.uid;

    if (!name) {
      console.log(`view: component needs a name (#${uid})`, c);
      return false;
    }

    // When debug mode is enabled, write logs to a collapsed group in the browser console:
    // https://developer.mozilla.org/en-US/docs/Web/API/console/groupCollapsed_static
    if (debug) {
      const scope = this.scopes.map((s) => `${s?.$options?.name} #${s?.$?.uid.toString()}`).join(" › ");
      // To make them easy to recognize, the collapsed view logs are displayed
      // in the browser console with bold white text on a purple background.
      console.groupCollapsed(
        `%c${scope}`,
        "background: #502A85; color: white; padding: 3px 5px; border-radius: 8px; font-weight: bold;"
      );
      console.log("data:", toRaw(c?.$data));
    }

    // Automatically focus the active component if its element tabindex attribute is set to "1":
    // https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/tabindex
    if (focusElement) {
      setFocus(focusElement, focusSelector, false);
    } else {
      setFocus(findFocusElement(c), false, false);
    }

    // Return, as it should not be necessary to apply the same state twice.
    if (this.uid === uid) {
      if (debug) {
        console.groupEnd();
      }
      return;
    }

    let hideScrollbar = this.layers() > 2 ? this.hideScrollbar : false;
    let disableScrolling = false;
    let disableNavigationGestures = false;
    let preventNavigation = uid > 0 && !name.startsWith("PPage");

    switch (name) {
      case "PPagePlaces":
        hideScrollbar = true;
        break;
      case "PServiceUpload":
        preventNavigation = false;
        break;
      case "PPageLogin":
        hideScrollbar = true;
        preventNavigation = true;
        break;
      case "PPhotoEditDialog":
        hideScrollbar = window.innerWidth < 960;
        disableScrolling = true;
        preventNavigation = true;
        break;
      case "PPhotoUploadDialog":
        hideScrollbar = window.innerWidth < 1280;
        disableScrolling = true;
        preventNavigation = true;
        break;
      case "PLightbox":
        hideScrollbar = true;
        disableScrolling = true;
        disableNavigationGestures = true;
        preventNavigation = true;
        break;
    }

    this.hideScrollbar = hideScrollbar;
    this.preventNavigation = preventNavigation;

    const htmlEl = getHtmlElement();

    if (!htmlEl) {
      if (debug) {
        console.log(`html: failed to get element (#${this.uid.toString()})`, c);
        console.groupEnd();
      }
      return false;
    }

    const bodyEl = getBodyElement();

    if (!bodyEl) {
      if (debug) {
        console.log(`body: failed to get element (#${this.uid.toString()})`, c);
        console.groupEnd();
      }
      return false;
    }

    if (hideScrollbar) {
      if (!bodyEl.classList.contains("hide-scrollbar")) {
        bodyEl.classList.add("hide-scrollbar");
        setHtmlStyle("scrollbar-width", "none");
        setHtmlStyle("overflow-y", "hidden");

        if (debug) {
          console.log(`html: added style="scrollbar-width: none; overflow-y: hidden;"`);
        }
      }
    } else if (bodyEl.classList.contains("hide-scrollbar")) {
      bodyEl.classList.remove("hide-scrollbar");
      setHtmlStyle("scrollbar-width");
      setHtmlStyle("overflow-y");

      if (debug) {
        console.log(`html: removed style="scrollbar-width: none; overflow-y: hidden;"`);
      }
    }

    if (disableScrolling) {
      if (!bodyEl.classList.contains("disable-scrolling")) {
        bodyEl.classList.add("disable-scrolling");
        if (debug) {
          console.log(`body: added class="disable-scrolling"`);
        }
      }
    } else if (bodyEl.classList.contains("disable-scrolling")) {
      bodyEl.classList.remove("disable-scrolling");
      if (debug) {
        console.log(`body: removed class="disable-scrolling"`);
      }
    }

    if (disableNavigationGestures) {
      if (!bodyEl.classList.contains("disable-navigation-gestures")) {
        bodyEl.classList.add("disable-navigation-gestures");
        window.addEventListener(TouchStartEvent, preventNavigationTouchEvent, { passive: false });
        window.addEventListener(TouchMoveEvent, preventNavigationTouchEvent, { passive: false });
        if (debug) {
          console.log(`view: disabled touch navigation gestures`);
        }
      }
    } else if (bodyEl.classList.contains("disable-navigation-gestures")) {
      bodyEl.classList.remove("disable-navigation-gestures");
      window.removeEventListener(TouchStartEvent, preventNavigationTouchEvent, false);
      window.removeEventListener(TouchMoveEvent, preventNavigationTouchEvent, false);
      if (debug) {
        console.log(`view: re-enabled touch navigation gestures`);
      }
    }

    if (debug) {
      console.groupEnd();
    }
    return true;
  }

  // Returns the current number of view layers.
  layers() {
    return this.scopes?.length ? this.scopes.length : 0;
  }

  // Returns the currently active view component or null if none exists.
  current() {
    if (this.scopes.length) {
      return this.scopes[this.scopes.length - 1];
    } else {
      return null;
    }
  }

  // Returns the parent view of the currently active view or null if none exists.
  parent() {
    if (this.scopes.length > 1) {
      return this.scopes[this.scopes.length - 2];
    } else {
      return null;
    }
  }

  // Returns the name of the parent view component or an empty string if none exists.
  parentName() {
    const c = this.parent();

    if (!c) {
      return "";
    }

    return c?.$options?.name ? c.$options.name : "";
  }

  // Returns the currently active view data or an empty reactive object otherwise.
  data() {
    const c = this.current();

    if (c && c.$data) {
      return c.$data;
    } else {
      return {};
    }
  }

  // Gives focus to the specified HTML element, or the first element that matches the specified selector string.
  focus(el, selector, scroll) {
    return setFocus(el, selector, scroll);
  }

  // Navigates to the specified URL, optionally with a delay set in milliseconds and a blocked user interface.
  redirect(url, delay, blockUI) {
    // Return if no URL was passed.
    if (!url) {
      console.warn(`cannot redirect because no URL was specified`);
      return;
    }

    // Verify that the target URL is different from the current location.
    const link = document.createElement("a");
    link.href = url;
    if (window.location.href === link.toString()) {
      console.warn(`cannot redirect to ${url} because it is the current location`);
      return;
    }

    // Block the user interface, if requested.
    if (blockUI) {
      $notify.blockUI();
    }

    // Make sure navigation is allowed.
    this.preventNavigation = false;

    // Navigate to the URL, optionally with the specified delay in milliseconds.
    if (delay) {
      if (trace) {
        console.log(`%credirect to "${url}" (${delay}ms delay)`, "color: #F06292");
      }

      setTimeout(() => {
        window.location = url;
      }, delay);
    } else {
      if (trace) {
        console.log(`%credirect to "${url}"`, "color: #F06292");
      }

      window.location = url;
    }
  }

  // Returns true if the specified view component is currently inactive, e.g. hidden in the background.
  isHidden(c) {
    return !this.isActive(c);
  }

  // Returns true if the specified view component is currently active, e.g. visible in the foreground.
  isActive(c) {
    if (!c || this.isApp()) {
      return true;
    }

    const context = this.scopes[this.scopes.length - 1];

    if (typeof c === "object") {
      return c === context;
    } else if (typeof c === "string") {
      return context?.$options?.name === c;
    } else if (typeof c === "number") {
      return context?.$?.uid === c;
    }

    return false;
  }

  // Returns true if no view is currently active.
  isRoot() {
    return !this.scopes.length;
  }

  // Returns true if no view or the main view of the app is currently active.
  isApp() {
    if (this.isRoot()) {
      return true;
    }

    const c = this.scopes[this.scopes.length - 1];

    return c?.$options?.name === "App" || c?.$?.uid === 0;
  }
}

// $view is the default View instance.
export const $view = new View();
