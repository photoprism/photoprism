// Utility class to detect, request, and exit fullscreen mode.
class Fullscreen {
  // Returns true if the browser supports the fullscreen API.
  isSupported() {
    // see https://developer.mozilla.org/en-US/docs/Web/API/Document/fullscreenEnabled
    return !!document.fullscreenEnabled;
  }

  // Returns true if fullscreen mode is enabled.
  isEnabled() {
    // see https://developer.mozilla.org/en-US/docs/Web/API/Document/fullscreenElement
    return !!document.fullscreenElement || !!document.mozFullScreenElement;
  }

  // Toggles fullscreen mode and returns a Promise.
  toggle() {
    if (this.isEnabled()) {
      return this.exit();
    } else {
      return this.request();
    }
  }

  // Exits fullscreen mode if enabled and returns a Promise.
  exit() {
    if (this.isEnabled()) {
      // see https://developer.mozilla.org/en-US/docs/Web/API/Document/exitFullscreen
      return document.exitFullscreen();
    }

    return Promise.resolve();
  }

  // Requests to enter fullscreen mode if not already enabled and returns a Promise.
  request() {
    if (!this.isEnabled()) {
      // see https://developer.mozilla.org/en-US/docs/Web/API/Element/requestFullscreen
      return document.documentElement.requestFullscreen({ navigationUI: "hide" });
    }

    return Promise.resolve();
  }
}

const $fullscreen = new Fullscreen();

export default $fullscreen;
