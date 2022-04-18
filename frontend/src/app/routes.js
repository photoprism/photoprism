/*

Copyright (c) 2018 - 2022 PhotoPrism UG. All rights reserved.

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

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import Photos from "pages/photos.vue";
import Albums from "pages/albums.vue";
import AlbumPhotos from "pages/album/photos.vue";
import Places from "pages/places.vue";
import Files from "pages/library/files.vue";
import Errors from "pages/library/errors.vue";
import Labels from "pages/labels.vue";
import People from "pages/people.vue";
import Library from "pages/library.vue";
import Settings from "pages/settings.vue";
import AuthLogin from "pages/auth/login.vue";
import Discover from "pages/discover.vue";
import About from "pages/about/about.vue";
import Feedback from "pages/about/feedback.vue";
import License from "pages/about/license.vue";
import Help from "pages/help.vue";
import { $gettext } from "common/vm";
import { config, session } from "./session";

const c = window.__CONFIG__;
const appName = c.name;
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
    meta: { title: appName, auth: false },
  },
  {
    name: "feedback",
    path: "/feedback",
    component: Feedback,
    meta: { title: appName, auth: true },
  },
  {
    name: "license",
    path: "/about/license",
    component: License,
    meta: { title: appName, auth: false },
  },
  {
    name: "help",
    path: "/help*",
    component: Help,
    meta: { title: appName, auth: false },
  },
  {
    name: "login",
    path: "/auth/login",
    component: AuthLogin,
    meta: { title: siteTitle, auth: false, hideNav: true },
    beforeEnter: (to, from, next) => {
      if (session.isUser()) {
        next({ name: "home" });
      } else {
        next();
      }
    },
  },
  {
    name: "browse",
    path: "/browse",
    component: Photos,
    meta: { title: appName, icon: true, auth: true },
  },
  {
    name: "all",
    path: "/all",
    component: Photos,
    meta: { title: $gettext("Search"), auth: true },
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
    props: { view: "moment", staticFilter: { type: "moment", order: "moment" } },
  },
  {
    name: "moment",
    path: "/moments/:uid/:slug",
    component: AlbumPhotos,
    meta: { title: $gettext("Moments"), auth: true },
  },
  {
    name: "albums",
    path: "/albums",
    component: Albums,
    meta: { title: $gettext("Albums"), auth: true },
    props: { view: "album", staticFilter: { type: "album", order: "name" } },
  },
  {
    name: "album",
    path: "/albums/:uid/:slug",
    component: AlbumPhotos,
    meta: { title: $gettext("Albums"), auth: true },
  },
  {
    name: "calendar",
    path: "/calendar",
    component: Albums,
    meta: { title: $gettext("Calendar"), auth: true },
    props: { view: "month", staticFilter: { type: "month", order: "newest" } },
  },
  {
    name: "month",
    path: "/calendar/:uid/:slug",
    component: AlbumPhotos,
    meta: { title: $gettext("Calendar"), auth: true },
  },
  {
    name: "folders",
    path: "/folders",
    component: Albums,
    meta: { title: $gettext("Folders"), auth: true },
    props: { view: "folder", staticFilter: { type: "folder", order: "newest" } },
  },
  {
    name: "folder",
    path: "/folders/:uid/:slug",
    component: AlbumPhotos,
    meta: { title: $gettext("Folders"), auth: true },
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
    name: "place",
    path: "/places/:q",
    component: Places,
    meta: { title: $gettext("Places"), auth: true },
  },
  {
    name: "states",
    path: "/states",
    component: Albums,
    meta: { title: $gettext("Places"), auth: true },
    props: { view: "state", staticFilter: { type: "state", order: "place" } },
  },
  {
    name: "state",
    path: "/states/:uid/:slug",
    component: AlbumPhotos,
    meta: { title: $gettext("Places"), auth: true },
  },
  {
    name: "files",
    path: "/library/files*",
    component: Files,
    meta: { title: $gettext("File Browser"), auth: true },
  },
  {
    name: "hidden",
    path: "/library/hidden",
    component: Photos,
    meta: { title: $gettext("Hidden Files"), auth: true },
    props: { staticFilter: { hidden: "true" } },
  },
  {
    name: "errors",
    path: "/library/errors",
    component: Errors,
    meta: { title: appName, auth: true },
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
            next({ name: "people_faces" });
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
    name: "library",
    path: "/library",
    component: Library,
    meta: { title: $gettext("Library"), auth: true, background: "application-light" },
    props: { tab: "library-index" },
  },
  {
    name: "library_import",
    path: "/library/import",
    component: Library,
    meta: { title: $gettext("Library"), auth: true, background: "application-light" },
    props: { tab: "library-import" },
  },
  {
    name: "library_logs",
    path: "/library/logs",
    component: Library,
    meta: { title: $gettext("Library"), auth: true, background: "application-light" },
    props: { tab: "library-logs" },
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
    name: "settings_library",
    path: "/settings/library",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-library" },
  },
  {
    name: "settings_sync",
    path: "/settings/sync",
    component: Settings,
    meta: {
      title: $gettext("Settings"),
      auth: true,
      settings: true,
      background: "application-light",
    },
    props: { tab: "settings-sync" },
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
    redirect: "/browse",
  },
];
