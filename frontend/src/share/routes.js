import Albums from "share/albums.vue";
import AlbumPhotos from "share/photos.vue";
import {$gettext} from "common/vm";

const c = window.__CONFIG__;

export default [
    {
        name: "home",
        path: "/",
        redirect: {name: "albums"},
    },
    {
        name: "albums",
        path: "/s/:token",
        component: Albums,
        meta: {title: c.siteAuthor, auth: true},
        props: {view: "album", staticFilter: {type: "album"}},
    },
    {
        name: "album",
        path: "/s/:token/:uid",
        component: AlbumPhotos,
        meta: {title: c.siteAuthor, auth: true},
    },
    {
        path: "*", redirect: {name: "albums"},
    },
];
