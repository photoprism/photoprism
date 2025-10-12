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

  function loadBundle(jsUri) {
    var script = document.createElement('script');
    script.src = jsUri;
    script.defer = true;
    document.body.appendChild(script);
  }

  var config = window.__CONFIG__ || {};
  var jsUri = config.jsUri || config.JsUri;

  var support = supportsModernJs();
  window.__PHOTOPRISM_SUPPORTS__ = support.ok;

  if (!support.ok) {
    window.__PHOTOPRISM_SUPPORTS_REASON__ = support.reason;
    showUnsupportedMessage('PhotoPrism requires Safari 13 (iOS 13) or a current version of Chrome, Edge, or Firefox. Please update your browser or switch to a supported device.');
    return;
  }

  if (!jsUri) {
    console.warn('PhotoPrism: Unable to find bundle URL (jsUri) in configuration.');
    return;
  }

  loadBundle(jsUri);
})();
