/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

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
import Login from "pages/login.vue";
import Discover from "pages/discover.vue";
import About from "pages/about/about.vue";
import License from "pages/about/license.vue";
import Help from "pages/help.vue";
import {$gettext} from "common/vm";

const c = window.__CONFIG__;

export default [
    {
        name: "home",
        path: "/",
        redirect: "/photos",
    },
    {
        name: "about",
        path: "/about",
        component: About,
        meta: {title: c.name, auth: false},
    },
    {
        name: "license",
        path: "/about/license",
        component: License,
        meta: {title: c.name, auth: false},
    },
    {
        name: "help",
        path: "/help*",
        component: Help,
        meta: {title: c.name, auth: false},
    },
    {
        name: "login",
        path: "/login",
        component: Login,
        meta: {auth: false},
    },
    {
        name: "photos",
        path: "/photos",
        component: Photos,
        meta: {title: c.siteCaption, auth: true},
        props: {staticFilter: {photo: "true"}},
    },
    {
        name: "moments",
        path: "/moments",
        component: Albums,
        meta: {title: $gettext("Moments"), auth: true},
        props: {view: "moment", staticFilter: {type: "moment"}},
    },
    {
        name: "moment",
        path: "/moments/:uid/:slug",
        component: AlbumPhotos,
        meta: {title: $gettext("Moments"), auth: true},
    },
    {
        name: "albums",
        path: "/albums",
        component: Albums,
        meta: {title: $gettext("Albums"), auth: true},
        props: {view: "album", staticFilter: {type: "album"}},
    },
    {
        name: "album",
        path: "/albums/:uid/:slug",
        component: AlbumPhotos,
        meta: {auth: true},
    },
    {
        name: "calendar",
        path: "/calendar",
        component: Albums,
        meta: {title: $gettext("Calendar"), auth: true},
        props: {view: "month", staticFilter: {type: "month"}},
    },
    {
        name: "month",
        path: "/calendar/:uid/:slug",
        component: AlbumPhotos,
        meta: {title: $gettext("Calendar"), auth: true},
    },
    {
        name: "folders",
        path: "/folders",
        component: Albums,
        meta: {title: $gettext("Folders"), auth: true},
        props: {view: "folder", staticFilter: {type: "folder"}},
    },
    {
        name: "folder",
        path: "/folders/:uid:/:slug",
        component: AlbumPhotos,
        meta: {title: $gettext("Folders"), auth: true},
    },
    {
        name: "unsorted",
        path: "/unsorted",
        component: Photos,
        meta: {title: $gettext("Unsorted"), auth: true},
        props: {staticFilter: {unsorted: true}},
    },
    {
        name: "favorites",
        path: "/favorites",
        component: Photos,
        meta: {title: $gettext("Favorites"), auth: true},
        props: {staticFilter: {favorite: true}},
    },
    {
        name: "videos",
        path: "/videos",
        component: Photos,
        meta: {title: $gettext("Videos"), auth: true},
        props: {staticFilter: {video: "true"}},
    },
    {
        name: "review",
        path: "/review",
        component: Photos,
        meta: {title: $gettext("Review"), auth: true},
        props: {staticFilter: {review: true}},
    },
    {
        name: "private",
        path: "/private",
        component: Photos,
        meta: {title: $gettext("Private"), auth: true},
        props: {staticFilter: {private: true}},
    },
    {
        name: "archive",
        path: "/archive",
        component: Photos,
        meta: {title: $gettext("Archive"), auth: true},
        props: {staticFilter: {archived: true}},
    },
    {
        name: "places",
        path: "/places",
        component: Places,
        meta: {title: $gettext("Places"), auth: true},
    },
    {
        name: "place",
        path: "/places/:q",
        component: Places,
        meta: {title: $gettext("Places"), auth: true},
    },
    {
        name: "states",
        path: "/states",
        component: Albums,
        meta: {title: $gettext("Places"), auth: true},
        props: {view: "state", staticFilter: {type: "state"}},
    },
    {
        name: "state",
        path: "/states/:uid/:slug",
        component: AlbumPhotos,
        meta: {title: $gettext("Places"), auth: true},
    },
    {
        name: "files",
        path: "/library/files*",
        component: Files,
        meta: {title: $gettext("File Browser"), auth: true},
    },
    {
        name: "hidden",
        path: "/library/hidden",
        component: Photos,
        meta: {title: $gettext("Hidden Files"), auth: true},
        props: {staticFilter: {hidden: true}},
    },
    {
        name: "errors",
        path: "/library/errors",
        component: Errors,
        meta: {title: c.name, auth: true},
    },
    {
        name: "labels",
        path: "/labels",
        component: Labels,
        meta: {title: $gettext("Labels"), auth: true},
    },
    {
        name: "browse",
        path: "/browse",
        component: Photos,
        meta: {title: $gettext("Search"), auth: true},
        props: {staticFilter: {quality: 0}},
    },
    {
        name: "people",
        path: "/people",
        component: People,
        meta: {title: $gettext("People"), auth: true},
    },
    {
        name: "library_logs",
        path: "/library/logs",
        component: Library,
        meta: {title: $gettext("Library"), auth: true, background: "application-light"},
        props: {tab: 2},
    },
    {
        name: "library_import",
        path: "/library/import",
        component: Library,
        meta: {title: $gettext("Library"), auth: true, background: "application-light"},
        props: {tab: 1},
    },
    {
        name: "library",
        path: "/library",
        component: Library,
        meta: {title: $gettext("Library"), auth: true, background: "application-light"},
        props: {tab: 0},
    },
    {
        name: "settings",
        path: "/settings",
        component: Settings,
        meta: {title: $gettext("Settings"), auth: true, background: "application-light"},
        props: {tab: 0},
    },
    {
        name: "settings_sync",
        path: "/settings/sync",
        component: Settings,
        meta: {title: $gettext("Settings"), auth: true, background: "application-light"},
        props: {tab: 1},
    },
    {
        name: "settings_account",
        path: "/settings/account",
        component: Settings,
        meta: {title: $gettext("Settings"), auth: true, background: "application-light"},
        props: {tab: 2},
    },
    {
        name: "discover",
        path: "/discover",
        component: Discover,
        meta: {title: $gettext("Discover"), auth: true, background: "application-light"},
        props: {tab: 0},
    },
    {
        name: "discover_similar",
        path: "/discover/similar",
        component: Discover,
        meta: {title: $gettext("Discover"), auth: true, background: "application-light"},
        props: {tab: 1},
    },
    {
        name: "discover_season",
        path: "/discover/season",
        component: Discover,
        meta: {title: $gettext("Discover"), auth: true, background: "application-light"},
        props: {tab: 2},
    },
    {
        name: "discover_random",
        path: "/discover/random",
        component: Discover,
        meta: {title: $gettext("Discover"), auth: true, background: "application-light"},
        props: {tab: 3},
    },
    {
        path: "*", redirect: "/photos",
    },
];
