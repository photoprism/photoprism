'use strict';

(function () {
  function supportsModernJs() {
    var checks = [
      { ok: function () { return typeof window.Promise === 'function'; }, reason: 'Promise' },
      { ok: function () { return typeof window.Symbol === 'function'; }, reason: 'Symbol' },
      { ok: function () { return typeof window.fetch === 'function'; }, reason: 'fetch' },
      { ok: function () { return typeof window.URL === 'function'; }, reason: 'URL' },
      { ok: function () { return typeof window.URLSearchParams === 'function'; }, reason: 'URLSearchParams' },
      { ok: function () { return typeof window.AbortController === 'function'; }, reason: 'AbortController' },
      { ok: function () { return typeof Object.assign === 'function'; }, reason: 'Object.assign' },
      { ok: function () { return typeof Array.from === 'function'; }, reason: 'Array.from' },
      { ok: function () { return typeof Array.prototype.flat === 'function'; }, reason: 'Array.prototype.flat' },
      // The supported floor (Chrome and Edge 119, Firefox 128, Safari 16.4): relative color syntax,
      // then the features the UI framework and map worker need unconditionally.
      {
        ok: function () {
          return !!(window.CSS && CSS.supports && CSS.supports('color', 'lab(from red l a b)'));
        },
        reason: 'CSS relative color syntax'
      },
      { ok: function () { return typeof window.CSSLayerBlockRule === 'function'; }, reason: 'CSS cascade layers' },
      { ok: function () { return !!(window.CSS && typeof CSS.registerProperty === 'function'); }, reason: 'CSS.registerProperty' },
      { ok: function () { return !!(window.Intl && typeof Intl.Segmenter === 'function'); }, reason: 'Intl.Segmenter' },
      { ok: function () { return typeof Object.hasOwn === 'function'; }, reason: 'Object.hasOwn' },
      {
        ok: function () {
          var script = document.createElement('script');
          return 'noModule' in script;
        },
        reason: 'script.noModule'
      }
    ];

    for (var i = 0; i < checks.length; i++) {
      if (!checks[i].ok()) {
        return { ok: false, reason: checks[i].reason };
      }
    }

    return { ok: true };
  }

  function showUnsupportedMessage(message) {
    var body = document.body;
    if (body && body.className.indexOf('unsupported-browser') === -1) {
      body.className += ' unsupported-browser';
    }

    var progress = document.getElementById('progress');
    if (progress) {
      progress.style.display = 'none';
    }

    var busy = document.getElementById('busy-overlay');
    if (busy) {
      busy.style.display = 'none';
    }

    var splashInfo = document.getElementById('splash-info');
    if (splashInfo) {
      splashInfo.innerHTML = '';
      var info = document.createElement('div');
      info.className = 'splash-warning';
      info.textContent = message;
      splashInfo.appendChild(info);
    }
  }

  var support = supportsModernJs();
  window.__PHOTOPRISM_SUPPORTS__ = support.ok;

  if (!support.ok) {
    window.__PHOTOPRISM_SUPPORTS_REASON__ = support.reason;
    showUnsupportedMessage('PhotoPrism requires a current version of Chrome, Safari, Edge, or Firefox. Please update your browser or switch to a supported device.');
  }
})();
