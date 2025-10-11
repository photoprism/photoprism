/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import "core-js/stable";
import "regenerator-runtime/runtime";
import $api from "common/api";
import $notify from "common/notify";
import { $view } from "common/view";
import { $lightbox } from "common/lightbox";
import { PhotoClipboard } from "common/clipboard";
import $event from "common/event";
import $log from "common/log";
import $util from "common/util";
import * as components from "component/components";
import icons from "component/icons";
import defaults from "component/defaults";
import PhotoPrism from "app.vue";
import { createRouter, createWebHistory } from "vue-router";
import routes from "app/routes";
import { $config, $session } from "app/session";
import { Settings as Luxon } from "luxon";
import Socket from "common/websocket";
import { createApp } from "vue";
import { createVuetify } from "vuetify";
import Vue3Sanitize from "vue-3-sanitize";
import VueSanitize from "vue-sanitize-directive";
import FloatingVue from "floating-vue";
import VueLuxon from "vue-luxon";
import { passiveSupport } from "passive-events-support/src/utils";
import * as themes from "options/themes";
import Hls from "hls.js";
import { createGettext, T } from "common/gettext";
import { Locale } from "locales";
import * as offline from "@lcdp/offline-plugin/runtime";
import { aliases, mdi } from "vuetify/iconsets/mdi";
import "vuetify/styles";
import "@mdi/font/css/materialdesignicons.css";
import "css/app.css";

// see https://www.npmjs.com/package/passive-events-support
passiveSupport({ events: ["touchstart", "touchmove", "wheel", "mousewheel"] });

// Check if running on a mobile device.
const $isMobile = $util.isMobile();

window.$isMobile = $isMobile;

$config.progress(50);

$config.update().finally(() => {
  // Initialize libs and framework.
  $config.progress(66);

  // Check if running in public mode.
  const $isPublic = $config.isPublic();

  let app = createApp(PhotoPrism);

  // Initialize language and detect its alignment.
  app.config.globalProperties.$language = $config.getLanguageLocale();
  Luxon.defaultLocale = $config.getLanguageCode();

  // Detect right-to-left languages such as Arabic and Hebrew
  const $isRtl = $config.isRtl();

  // HTTP Live Streaming (video support).
  window.Hls = Hls;

  // Assign helpers to VueJS prototype.
  app.config.globalProperties.$isRtl = $isRtl;
  app.config.globalProperties.$isMobile = $isMobile;
  app.config.globalProperties.$event = $event;
  app.config.globalProperties.$notify = $notify;
  app.config.globalProperties.$view = $view;
  app.config.globalProperties.$lightbox = $lightbox;
  app.config.globalProperties.$session = $session;
  app.config.globalProperties.$api = $api;
  app.config.globalProperties.$log = $log;
  app.config.globalProperties.$socket = Socket;
  app.config.globalProperties.$config = $config;
  app.config.globalProperties.$clipboard = PhotoClipboard;
  app.config.globalProperties.$util = $util;
  app.config.globalProperties.$sponsorFeatures = () => {
    return $config.load().finally(() => {
      if ($config.values.sponsor) {
        return Promise.resolve();
      } else {
        return Promise.reject();
      }
    });
  };

  // Create Vue 3 Gettext instance.
  const gettext = createGettext($config);

  // Create Vuetify 3 instance.
  const vuetify = createVuetify({
    defaults,
    icons: {
      defaultSet: "mdi",
      aliases,
      sets: {
        mdi,
        ...icons,
      },
    },
    theme: {
      defaultTheme: $config.themeName,
      themes: themes.All(),
      variations: themes.variations,
    },
    locale: Locale(),
  });

  // Use Vuetify 3.
  app.use(vuetify);

  // Use Vue 3 Gettext.
  app.use(gettext);

  // Use HTML sanitizer with v-sanitize directive.
  app.use(Vue3Sanitize, {
    allowedTags: ["b", "strong", "span"],
    allowedAttributes: { b: ["dir"], strong: ["dir"], span: ["dir"] },
  });
  app.use(VueSanitize);

  // FloatingVue is a library to easily add tooltips to the UI:
  // https://floating-vue.starpad.dev/guide/installation
  FloatingVue.options.themes.tooltip.placement = "top";
  app.use(FloatingVue);

  // TODO: check it
  // debugger;
  // app.use(VueLuxon);
  app.config.globalProperties.$luxon = VueLuxon;
  components.install(app);

  // Persist scroll position when the browser parks the page in the back/forward cache (e.g. iOS PWA).
  window.addEventListener("pagehide", (ev) => {
    if (ev.persisted) {
      $view.saveWindowScrollPos();
    }
  });

  // When the page is resumed, either restore the stored scroll position (bfcache) or clear stale data.
  window.addEventListener("pageshow", (ev) => {
    if (ev.persisted) {
      $view.restoreWindowScrollPos();
    } else {
      $view.clearWindowScrollPos();
    }
  });

  // Configure the Vue Router; base path mirrors the library prefix so deep links behave correctly.
  const router = createRouter({
    history: createWebHistory($config.baseUri + "/library/"),
    routes: routes,
    // Apply the last saved scroll position when navigating within the SPA.
    scrollBehavior(to, from, savedPosition) {
      const prevScrollPos = $view.getWindowScrollPos(savedPosition);

      if (prevScrollPos) {
        // Wait until Vue has painted and all pending AJAX batches finish before restoring the offset.
        return new Promise((resolve) => {
          requestAnimationFrame(() => {
            $notify.ajaxWait().then(() => {
              setTimeout(() => {
                resolve(prevScrollPos);
              }, 200);
            });
          });
        });
      } else {
        return { left: 0, top: 0 };
      }
    },
  });

  const currentHistoryState = () => {
    if (router.options?.history?.state) {
      return router.options.history.state;
    }

    if (typeof window !== "undefined" && typeof window.history !== "undefined") {
      return window.history.state;
    }

    return undefined;
  };

  router.beforeEach((to, from, next) => {
    $view.prepareNavigation(currentHistoryState());
    next();
  });

  router.afterEach(() => {
    $view.commitNavigation(currentHistoryState());
    requestAnimationFrame(() => {
      $view.resetNavigationDirection();
    });
  });

  // Add global guards to block navigation when dialogs are open and enforce auth/settings rules.
  router.beforeEach((to) => {
    if ($view.preventNavigation) {
      // Disable navigation when a fullscreen dialog or lightbox is open.
      return false;
    } else if (to.matched.some((record) => record.meta.settings) && $config.values.disable.settings) {
      return { name: "home" };
    } else if (to.matched.some((record) => record.meta.admin)) {
      if ($isPublic || $session.isAdmin()) {
        return true;
      } else {
        $session.setLoginRedirectUrl(to.href);
        return { name: "login" };
      }
    } else if (to.matched.some((record) => record.meta.requiresAuth)) {
      if ($isPublic || $session.isUser()) {
        return true;
      } else {
        $session.setLoginRedirectUrl(to.href);
        return { name: "login" };
      }
    } else {
      return true;
    }
  });

  // Update the document title after every route change to reflect the active view.
  router.afterEach((to) => {
    const t = to.meta["title"] ? to.meta["title"] : "";

    if (t !== "" && $config.values.siteTitle !== t && $config.values.name !== t) {
      $config.page.title = T(t);

      if ($config.page.title.startsWith($config.values.siteTitle)) {
        window.document.title = $config.page.title;
      } else if ($config.page.title === "") {
        window.document.title = $config.values.siteTitle;
      } else {
        window.document.title = $config.page.title + " – " + $config.values.siteTitle;
      }
    } else {
      $config.page.title = $config.values.name;

      if ($config.values.siteCaption === "") {
        window.document.title = $config.values.siteTitle;
      } else {
        window.document.title = $config.values.siteCaption;
      }
    }
  });

  // Use router.
  app.use(router);
  window.$router = router;

  if ($isMobile) {
    // Add "mobile" class to body if running on a mobile device.
    document.body.classList.add("mobile");
  } else {
    // Pull client config every 10 minutes in case push fails (except on mobile to save battery).
    setInterval(() => $config.update(), 600000);
  }

  // Mount to #app.
  app.mount("#app");

  // Allows the application to be installed as a PWA.
  if ($config.baseUri === "") {
    offline.install();
  }
});
