/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

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
import "common/navigation";
import Api from "common/api";
import Notify from "common/notify";
import Scrollbar from "common/scrollbar";
import Clipboard from "common/clipboard";
import { installComponents } from "component/components";
import { installDialogs } from "dialog/dialogs";
import customIcons from "component/icons";
import Event from "pubsub-js";
import { createGettext } from "vue3-gettext";
import Log from "common/log";
import PhotoPrism from "app.vue";
import { createRouter, createWebHistory } from "vue-router";
import routes from "app/routes";
import { config, session } from "app/session";
import { Settings } from "luxon";
import Socket from "common/websocket";
import Viewer from "common/viewer";
import { createApp } from "vue";
import { createVuetify } from "vuetify";
import VueLuxon from "vue-luxon";
import * as themes from "options/themes";
// import VueFilters from "vue2-filters";
// import VueFullscreen from "vue-fullscreen";
import VueInfiniteScroll from "vue-infinite-scroll";
import Hls from "hls.js";
import "common/maptiler-lang";
import { T, Mount } from "common/vm";
import * as offline from "@lcdp/offline-plugin/runtime";
import { aliases, mdi } from "vuetify/iconsets/mdi";
import "vuetify/styles";
import "@mdi/font/css/materialdesignicons.css";

import { passiveSupport } from "passive-events-support/src/utils";
passiveSupport({ events: ["touchstart", "touchmove", "wheel", "mousewheel"] });

config.progress(50);

config.update().finally(() => {
  // Initialize libs and framework.
  config.progress(66);
  const viewer = new Viewer();
  const isPublic = config.isPublic();
  const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || (navigator.maxTouchPoints && navigator.maxTouchPoints > 2);

  let app = createApp(PhotoPrism);
  // Initialize language and detect alignment.
  app.config.globalProperties.$language = config.getLanguage();
  Settings.defaultLocale = app.config.globalProperties.$language.substring(0, 2);
  // Detect right-to-left languages such as Arabic and Hebrew
  const rtl = config.rtl();

  // HTTP Live Streaming (video support).
  window.Hls = Hls;

  // Assign helpers to VueJS prototype.
  app.config.globalProperties.$event = Event;
  app.config.globalProperties.$notify = Notify;
  app.config.globalProperties.$scrollbar = Scrollbar;
  app.config.globalProperties.$viewer = viewer;
  app.config.globalProperties.$session = session;
  app.config.globalProperties.$api = Api;
  app.config.globalProperties.$log = Log;
  app.config.globalProperties.$socket = Socket;
  app.config.globalProperties.$config = config;
  app.config.globalProperties.$clipboard = Clipboard;
  app.config.globalProperties.$isMobile = isMobile;
  app.config.globalProperties.$rtl = rtl;
  app.config.globalProperties.$sponsorFeatures = () => {
    return config.load().finally(() => {
      if (config.values.sponsor) {
        return Promise.resolve();
      } else {
        return Promise.reject();
      }
    });
  };

  // Create Vuetify instance.
  const vuetify = createVuetify({
    rtl,
    defaults: {
      VBtn: {
        flat: true,
        variant: "flat",
        ripple: true,
      },
      VSwitch: {
        flat: true,
        density: "compact",
        baseColor: "surface",
        hideDetails: true,
        ripple: false,
      },
      VRating: {
        density: "compact",
        color: "on-surface",
        activeColor: "surface-variant",
        ripple: false,
      },
      VCheckbox: {
        density: "compact",
        color: "surface-variant",
        hideDetails: "auto",
        ripple: false,
      },
      VTextField: {
        flat: true,
        variant: "solo-filled",
        color: "surface-variant",
        hideDetails: "auto",
      },
      VTextarea: {
        flat: true,
        variant: "solo-filled",
        color: "surface-variant",
        hideDetails: "auto",
      },
      VAutocomplete: {
        flat: true,
        variant: "solo-filled",
        color: "surface-variant",
        itemTitle: "text",
        itemValue: "value",
        hideDetails: "auto",
        hideNoData: true,
      },
      VCombobox: {
        flat: true,
        variant: "solo-filled",
        color: "surface-variant",
        itemTitle: "text",
        itemValue: "value",
        hideDetails: "auto",
      },
      VSelect: {
        flat: true,
        variant: "solo-filled",
        color: "surface-variant",
        itemTitle: "text",
        itemValue: "value",
        hideDetails: "auto",
      },
      VCard: {
        density: "compact",
        color: "background",
        flat: true,
        ripple: false,
      },
      VTab: {
        color: "on-surface",
        baseColor: "on-surface-variant",
        ripple: false,
      },
      VTabs: {
        grow: true,
        elevation: 0,
        color: "on-surface",
        bgColor: "secondary",
        baseColor: "secondary",
        sliderColor: "surface-variant",
      },
      VTable: {
        density: "comfortable",
      },
      VListItem: {
        ripple: false,
      },
      VDataTable: {
        color: "background",
      },
      VExpansionPanel: {
        tile: true,
        ripple: false,
      },
      VExpansionPanels: {
        flat: true,
        tile: true,
        static: true,
        variant: "accordion",
        bgColor: "card",
        ripple: false,
      },
      VProgressLinear: {
        height: 10,
        rounded: true,
        color: "surface-variant",
      },
    },
    icons: {
      defaultSet: "mdi",
      aliases,
      sets: {
        mdi,
        ...customIcons,
      },
    },
    theme: {
      defaultTheme: config.themeName,
      themes: themes.All(),
    },
  });
  app.use(vuetify);

  // Register other VueJS plugins.
  const gettext = createGettext({
    translations: config.translations,
    silent: true, // !config.values.debug,
    defaultLanguage: app.config.globalProperties.$language,
    // autoAddKeyAttributes: true,
  });
  app.use(gettext);

  // TODO: check it
  // debugger;
  // app.use(VueLuxon);
  app.config.globalProperties.$luxon = VueLuxon;
  app.config.globalProperties.$fullscreen = VueInfiniteScroll;
  app.use(VueInfiniteScroll);
  // app.use(VueFullscreen);
  // app.use(VueFilters);
  // app.use(Components);
  installComponents(app);
  // app.use(Dialogs);
  installDialogs(app);

  // make scroll-pos-restore compatible with bfcache
  // this is required to make scroll-pos-restore work on iOS in PWA-mode
  window.addEventListener("pagehide", (event) => {
    if (event.persisted) {
      localStorage.setItem("lastScrollPosBeforePageHide", JSON.stringify({ x: window.scrollX, y: window.scrollY }));
    }
  });
  window.addEventListener("pageshow", (event) => {
    if (event.persisted) {
      const lastSavedScrollPos = localStorage.getItem("lastScrollPosBeforePageHide");
      if (lastSavedScrollPos !== undefined && lastSavedScrollPos !== null && lastSavedScrollPos !== "") {
        window.positionToRestore = JSON.parse(localStorage.getItem("lastScrollPosBeforePageHide"));
        // wait for other things that set the scroll-pos anywhere in the app to fire
        setTimeout(() => {
          if (window.positionToRestore !== undefined) {
            window.scrollTo(window.positionToRestore.x, window.positionToRestore.y);
          }
        }, 50);

        // let's give the scrollBehaviour-function some time to use the restored
        // position instead of resetting the scroll-pos to 0,0
        setTimeout(() => {
          window.positionToRestore = undefined;
        }, 250);
      }
    }

    localStorage.removeItem("lastScrollPosBeforePageHide");
  });

  // Configure client-side routing.
  const router = createRouter({
    history: createWebHistory(config.baseUri + "/library/"),
    routes: routes,
    scrollBehavior(to, from, savedPosition) {
      let prevScrollPos = savedPosition;

      if (window.positionToRestore !== undefined) {
        prevScrollPos = window.positionToRestore;
      }
      window.positionToRestore = undefined;

      if (prevScrollPos) {
        return new Promise((resolve) => {
          Notify.ajaxWait().then(() => {
            setTimeout(() => {
              resolve(prevScrollPos);
            }, 200);
          });
        });
      } else {
        return { left: 0, top: 0 };
      }
    },
  });
  app.use(router);

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
    const t = to.meta["title"] ? to.meta["title"] : "";

    if (t !== "" && config.values.siteTitle !== t && config.values.name !== t) {
      config.page.title = T(t);

      if (config.page.title.startsWith(config.values.siteTitle)) {
        window.document.title = config.page.title;
      } else if (config.page.title === "") {
        window.document.title = config.values.siteTitle;
      } else {
        window.document.title = config.page.title + " – " + config.values.siteTitle;
      }
    } else {
      config.page.title = config.values.name;

      if (config.values.siteCaption === "") {
        window.document.title = config.values.siteTitle;
      } else {
        window.document.title = config.values.siteCaption;
      }
    }
  });

  if (isMobile) {
    document.body.classList.add("mobile");
  } else {
    // Pull client config every 10 minutes in case push fails (except on mobile to save battery).
    setInterval(() => config.update(), 600000);
  }

  // Start application.
  Mount(app, router);
  if (config.baseUri === "") {
    offline.install();
  }
});
