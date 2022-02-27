/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import "core-js/stable";
import "regenerator-runtime/runtime";
import Api from "common/api";
import Notify from "common/notify";
import Scrollbar from "common/scrollbar";
import Clipboard from "common/clipboard";
import Components from "share/components";
import icons from "component/icons";
import Dialogs from "dialog/dialogs";
import Event from "pubsub-js";
import GetTextPlugin from "vue-gettext";
import Log from "common/log";
import PhotoPrism from "share.vue";
import Router from "vue-router";
import Routes from "share/routes";
import { config, session } from "app/session";
import { Settings } from "luxon";
import Socket from "common/websocket";
import Viewer from "common/viewer";
import Vue from "vue";
import Vuetify from "vuetify";
import VueLuxon from "vue-luxon";
import VueFilters from "vue2-filters";
import VueFullscreen from "vue-fullscreen";
import VueInfiniteScroll from "vue-infinite-scroll";
import Hls from "hls.js";
import { $gettext, Mount } from "common/vm";
import * as options from "./options/options";

// Initialize helpers
const viewer = new Viewer();
const isPublic = config.get("public");
const isMobile =
  /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) ||
  (navigator.maxTouchPoints && navigator.maxTouchPoints > 2);

// Initialize language and detect alignment
Vue.config.language = config.values.settings.ui.language;
Settings.defaultLocale = Vue.config.language.substring(0, 2);
const languages = options.Languages();
const rtl = languages.some((lang) => lang.value === Vue.config.language && lang.rtl);

// Get initial theme colors from config
const theme = config.theme.colors;

// HTTP Live Streaming (video support)
window.Hls = Hls;

// Assign helpers to VueJS prototype
Vue.prototype.$event = Event;
Vue.prototype.$notify = Notify;
Vue.prototype.$scrollbar = Scrollbar;
Vue.prototype.$viewer = viewer;
Vue.prototype.$session = session;
Vue.prototype.$api = Api;
Vue.prototype.$log = Log;
Vue.prototype.$socket = Socket;
Vue.prototype.$config = config;
Vue.prototype.$clipboard = Clipboard;
Vue.prototype.$isMobile = isMobile;
Vue.prototype.$rtl = rtl;

// Register Vuetify
Vue.use(Vuetify, { rtl, icons, theme });

// Register other VueJS plugins
Vue.use(GetTextPlugin, {
  translations: config.translations,
  silent: true, // !config.values.debug,
  defaultLanguage: Vue.config.language,
  autoAddKeyAttributes: true,
});

Vue.use(VueLuxon);
Vue.use(VueInfiniteScroll);
Vue.use(VueFullscreen);
Vue.use(VueFilters);
Vue.use(Components);
Vue.use(Dialogs);
Vue.use(Router);

// Configure client-side routing
const router = new Router({
  routes: Routes,
  mode: "history",
  base: config.baseUri + "/",
  saveScrollPosition: true,
  scrollBehavior: (to, from, savedPosition) => {
    if (savedPosition) {
      return new Promise((resolve) => {
        Notify.ajaxWait().then(() => {
          setTimeout(() => {
            resolve(savedPosition);
          }, 200);
        });
      });
    } else {
      return { x: 0, y: 0 };
    }
  },
});

router.beforeEach((to, from, next) => {
  if (document.querySelector(".v-dialog--active.v-dialog--fullscreen")) {
    // Disable back button in full-screen viewers and editors.
    next(false);
  } else if (to.matched.some((record) => record.meta.settings) && config.values.disable.settings) {
    next({ name: "home" });
  } else if (to.matched.some((record) => record.meta.admin)) {
    if (isPublic || session.isAdmin()) {
      next();
    } else {
      next({
        name: "login",
        params: { nextUrl: to.fullPath },
      });
    }
  } else if (to.matched.some((record) => record.meta.auth)) {
    if (isPublic || session.isUser()) {
      next();
    } else {
      next({
        name: "login",
        params: { nextUrl: to.fullPath },
      });
    }
  } else {
    next();
  }
});

router.afterEach((to) => {
  if (to.meta.title && config.values.siteTitle !== to.meta.title) {
    config.page.title = $gettext(to.meta.title);
    window.document.title = config.page.title;
  } else {
    config.page.title = config.values.siteTitle;
    window.document.title = config.values.siteTitle;
  }
});

if (isMobile) {
  document.body.classList.add("mobile");
} else {
  // Pull client config every 10 minutes in case push fails (except on mobile to save battery).
  setInterval(() => config.update(), 600000);
}

// Set body class for chrome-only optimizations.
if (navigator.appVersion.indexOf("Chrome/") !== -1) {
  document.body.classList.add("chrome");
}

// Start application.
Mount(Vue, PhotoPrism, router);
