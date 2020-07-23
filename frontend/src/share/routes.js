import Albums from "share/albums.vue";
import AlbumPhotos from "share/photos.vue";
import {$gettext} from "common/vm";

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
        meta: {title: $gettext("Albums"), auth: true},
        props: {view: "album", staticFilter: {type: "album"}},
    },
    {
        name: "album",
        path: "/s/:token/:uid",
        component: AlbumPhotos,
        meta: {title: $gettext("Albums"), auth: true},
    },
    {
        path: "*", redirect: {name: "albums"},
    },
];
