/*

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

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

import Photos from "page/photos.vue";
import Albums from "page/albums.vue";
import AlbumPhotos from "page/album/photos.vue";
import Places from "page/places.vue";
import Browse from "page/library/browse.vue";
import Errors from "page/library/errors.vue";
import Labels from "page/labels.vue";
import People from "page/people.vue";
import Library from "page/library.vue";
import Settings from "page/settings.vue";
import Admin from "page/admin.vue";
import Login from "page/auth/login.vue";
import Discover from "page/discover.vue";
import About from "page/about/about.vue";
import Feedback from "page/about/feedback.vue";
import License from "page/about/license.vue";
import Help from "page/help.vue";
import Connect from "page/connect.vue";
import { $gettext } from "common/vm";
import { config, session } from "./session";

const c = window.__CONFIG__;
const siteTitle = c.siteTitle ? c.siteTitle : c.name;

export default [
  {
    name: "home",
    path: "/",
    redirect: "/browse",
  },
  {
    name: "about",
    path: "/about",
    component: About,
    meta: { title: $gettext("About"), auth: false },
  },
  {
    name: "license",
    path: "/license",
    component: License,
    meta: { title: $gettext("License"), auth: false },
  },
  {
    name: "feedback",
    path: "/feedback",
    component: Feedback,
    meta: { title: $gettext("Help & Support"), auth: true },
  },
  {
    name: "help",
    path: "/help*",
    component: Help,
    meta: { title: $gettext("Help & Support"), auth: false },
  },
  {
    name: "login",
    path: "/login",
    component: Login,
    meta: { title: siteTitle, auth: false, hideNav: true },
    beforeEnter: (to, from, next) => {
      if (session.loginRequired()) {
        next();
      } else if (config.deny("photos", "search")) {
        next({ name: "albums" });
      } else {
        next({ name: "browse" });
      }
    },
  },
  {
    name: "admin",
    path: "/admin/*",
    component: Admin,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      admin: true,
      settings: true,
      background: "application-light",
    },
  },
  {
    name: "upgrade",
    path: "/upgrade",
    component: Connect,
    meta: {
      title: siteTitle,
      auth: true,
      admin: true,
      settings: true,
    },
  },
  {
    name: "connect",
    path: "/upgrade/:token",
    component: Connect,
    meta: {
      title: siteTitle,
      auth: true,
      admin: true,
      settings: true,
    },
  },
  {
    name: "browse",
    path: "/browse",
    component: Photos,
    meta: { title: siteTitle, icon: true, auth: true },
    beforeEnter: (to, from, next) => {
      if (session.loginRequired()) {
        next({ name: "login" });
      } else if (config.deny("photos", "search")) {
        next({ name: "albums" });
      } else {
        next();
      }
    },
  },
  {
    name: "all",
    path: "/all",
    component: Photos,
    meta: { title: siteTitle, auth: true },
    props: { staticFilter: { quality: "0" } },
  },
  {
    name: "photos",
    path: "/photos",
    component: Photos,
    meta: { title: $gettext("Photos"), auth: true },
    props: { staticFilter: { photo: "true" } },
  },
  {
    name: "moments",
    path: "/moments",
    component: Albums,
    meta: { title: $gettext("Moments"), auth: true },
    props: { view: "moment", defaultOrder: "newest", staticFilter: { type: "moment" } },
  },
  {
    name: "moment",
    path: "/moments/:album/:slug",
    component: AlbumPhotos,
    meta: { collName: "Moments", collRoute: "moments", auth: true },
  },
  {
    name: "albums",
    path: "/albums",
    component: Albums,
    meta: { title: $gettext("Albums"), auth: true },
    props: { view: "album", defaultOrder: "favorites", staticFilter: { type: "album" } },
  },
  {
    name: "album",
    path: "/albums/:album/:slug",
    component: AlbumPhotos,
    meta: { collName: "Albums", collRoute: "albums", auth: true },
  },
  {
    name: "calendar",
    path: "/calendar",
    component: Albums,
    meta: { title: $gettext("Calendar"), auth: true },
    props: { view: "month", defaultOrder: "newest", staticFilter: { type: "month" } },
  },
  {
    name: "month",
    path: "/calendar/:album/:slug",
    component: AlbumPhotos,
    meta: { collName: "Calendar", collRoute: "calendar", auth: true },
  },
  {
    name: "folders",
    path: "/folders",
    component: Albums,
    meta: { title: $gettext("Folders"), auth: true },
    props: { view: "folder", defaultOrder: "name", staticFilter: { type: "folder" } },
  },
  {
    name: "folder",
    path: "/folders/:album/:slug",
    component: AlbumPhotos,
    meta: { collName: "Folders", collRoute: "folders", auth: true },
  },
  {
    name: "unsorted",
    path: "/unsorted",
    component: Photos,
    meta: { title: $gettext("Unsorted"), auth: true },
    props: { staticFilter: { unsorted: "true" } },
  },
  {
    name: "favorites",
    path: "/favorites",
    component: Photos,
    meta: { title: $gettext("Favorites"), auth: true },
    props: { staticFilter: { favorite: "true" } },
  },
  {
    name: "live",
    path: "/live",
    component: Photos,
    meta: { title: $gettext("Live"), auth: true },
    props: { staticFilter: { live: "true" } },
  },
  {
    name: "videos",
    path: "/videos",
    component: Photos,
    meta: { title: $gettext("Videos"), auth: true },
    props: { staticFilter: { video: "true" } },
  },
  {
    name: "review",
    path: "/review",
    component: Photos,
    meta: { title: $gettext("Review"), auth: true },
    props: { staticFilter: { review: "true" } },
  },
  {
    name: "private",
    path: "/private",
    component: Photos,
    meta: { title: $gettext("Private"), auth: true },
    props: { staticFilter: { private: "true" } },
  },
  {
    name: "archive",
    path: "/archive",
    component: Photos,
    meta: { title: $gettext("Archive"), auth: true },
    props: { staticFilter: { archived: "true" } },
  },
  {
    name: "places",
    path: "/places",
    component: Places,
    meta: { title: $gettext("Places"), auth: true },
  },
  {
    name: "places_view",
    path: "/places/view/:s",
    component: Places,
    meta: { title: $gettext("Places"), auth: true },
  },
  {
    name: "places_browse",
    path: "/places/browse",
    component: Photos,
    meta: { title: $gettext("Places"), auth: true },
    beforeEnter: (to, from, next) => {
      if (session.loginRequired()) {
        next({ name: "login" });
      } else if (config.deny("photos", "search")) {
        next({ name: "albums" });
      } else {
        next();
      }
    },
  },
  {
    name: "states",
    path: "/states",
    component: Albums,
    meta: { title: $gettext("Places"), auth: true },
    props: { view: "state", defaultOrder: "place", staticFilter: { type: "state" } },
  },
  {
    name: "state",
    path: "/states/:album/:slug",
    component: AlbumPhotos,
    meta: { collName: "Places", collRoute: "states", auth: true },
  },
  {
    name: "files",
    path: "/index/files*",
    component: Browse,
    meta: { title: $gettext("File Browser"), auth: true },
  },
  {
    name: "hidden",
    path: "/hidden",
    component: Photos,
    meta: { title: $gettext("Hidden Files"), auth: true },
    props: { staticFilter: { hidden: "true" } },
  },
  {
    name: "errors",
    path: "/errors",
    component: Errors,
    meta: { title: $gettext("Errors"), auth: true },
  },
  {
    name: "labels",
    path: "/labels",
    component: Labels,
    meta: { title: $gettext("Labels"), auth: true },
  },
  {
    name: "people",
    path: "/people",
    component: People,
    meta: { title: $gettext("People"), auth: true, background: "application-light" },
    beforeEnter: (to, from, next) => {
      if (!config || !from || !from.name || from.name.startsWith("people")) {
        next();
      } else {
        config.load().finally(() => {
          // Open new faces tab when there are no people.
          if (config.values.count.people === 0) {
            if (config.allow("people", "manage")) {
              next({ name: "people_faces" });
            } else {
              next({ name: "albums" });
            }
          } else {
            next();
          }
        });
      }
    },
  },
  {
    name: "people_faces",
    path: "/people/new",
    component: People,
    meta: { title: $gettext("People"), auth: true, background: "application-light" },
  },
  {
    name: "library_index",
    path: "/index",
    component: Library,
    meta: { title: $gettext("Library"), auth: true, background: "application-light" },
    props: { tab: "library_index" },
  },
  {
    name: "library_import",
    path: "/import",
    component: Library,
    meta: { title: $gettext("Library"), auth: true, background: "application-light" },
    props: { tab: "library_import" },
  },
  {
    name: "library_logs",
    path: "/logs",
    component: Library,
    meta: { title: $gettext("Library"), auth: true, background: "application-light" },
    props: { tab: "library_logs" },
  },
  {
    name: "settings",
    path: "/settings",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-general" },
  },
  {
    name: "settings_media",
    path: "/settings/media",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      admin: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-media" },
  },
  {
    name: "settings_advanced",
    path: "/settings/advanced",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      admin: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-advanced" },
  },
  {
    name: "settings_services",
    path: "/settings/services",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-services" },
  },
  {
    name: "settings_account",
    path: "/settings/account",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-account" },
  },
  {
    name: "discover",
    path: "/discover",
    component: Discover,
    meta: { title: $gettext("Discover"), auth: true, background: "application-light" },
    props: { tab: 0 },
  },
  {
    name: "discover_similar",
    path: "/discover/similar",
    component: Discover,
    meta: { title: $gettext("Discover"), auth: true, background: "application-light" },
    props: { tab: 1 },
  },
  {
    name: "discover_season",
    path: "/discover/season",
    component: Discover,
    meta: { title: $gettext("Discover"), auth: true, background: "application-light" },
    props: { tab: 2 },
  },
  {
    name: "discover_random",
    path: "/discover/random",
    component: Discover,
    meta: { title: $gettext("Discover"), auth: true, background: "application-light" },
    props: { tab: 3 },
  },
  {
    path: "*",
    redirect: "/albums",
  },
];
