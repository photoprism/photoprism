import Albums from "share/albums.vue";
import AlbumPhotos from "share/photos.vue";

export default [
    {
        name: "home",
        path: "/",
        redirect: {name: 'albums'},
    },
    {
        name: "albums",
        path: "/s/:token",
        component: Albums,
        meta: {title: "PhotoPrism", auth: true},
        props: {view: "album", staticFilter: {type: "album"}},
    },
    {
        name: "album",
        path: "/s/:token/:uid",
        component: AlbumPhotos,
        meta: {title: "PhotoPrism", auth: true},
    },
    {
        path: "*", redirect: {name: 'albums'},
    },
];
