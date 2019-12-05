import Photos from "pages/photos.vue";
import Albums from "pages/albums.vue";
import AlbumPhotos from "pages/album/photos.vue";
import Places from "pages/places.vue";
import Labels from "pages/labels.vue";
import Events from "pages/events.vue";
import People from "pages/people.vue";
import Library from "pages/library.vue";
import Share from "pages/share.vue";
import Settings from "pages/settings.vue";
import Login from "pages/login.vue";
import Todo from "pages/todo.vue";

export default [
    {
        name: "home",
        path: "/",
        redirect: "/photos",
    },
    {
        name: "login",
        path: "/login",
        component: Login,
        meta: {title: "Login"},
    },
    {
        name: "photos",
        path: "/photos",
        component: Photos,
        meta: {title: "Browse your life"},
    },
    {
        name: "albums",
        path: "/albums",
        component: Albums,
        meta: {title: "Albums"},
    },
    {
        name: "album",
        path: "/albums/:uuid/:slug",
        component: AlbumPhotos,
        meta: {title: ""},
    },
    {
        name: "favorites",
        path: "/favorites",
        component: Photos,
        meta: {title: "Favorites"},
        props: {staticFilter: {favorites: true}},
    },
    {
        name: "places",
        path: "/places",
        component: Places,
        meta: {title: "Places"},
    },
    {
        name: "labels",
        path: "/labels",
        component: Labels,
        meta: {title: "Labels"},
    },
    {
        name: "events",
        path: "/events",
        component: Events,
        meta: {title: "Events"},
    },
    {
        name: "people",
        path: "/people",
        component: People,
        meta: {title: "People"},
    },
    {
        name: "filters",
        path: "/filters",
        component: Todo,
        meta: {title: "Filters"},
    },
    {
        name: "library",
        path: "/library",
        component: Library,
        meta: {title: "Library", auth: true},
        props: {tab: 0},
    },
    {
        name: "library_upload",
        path: "/library/upload",
        component: Library,
        meta: {title: "Library", auth: true},
        props: {tab: 0},
    },
    {
        name: "library_import",
        path: "/library/import",
        component: Library,
        meta: {title: "Library", auth: true},
        props: {tab: 1},
    },
    {
        name: "library_index",
        path: "/library/index",
        component: Library,
        meta: {title: "Library", auth: true},
        props: {tab: 2},
    },
    {
        name: "share",
        path: "/share",
        component: Share,
        meta: {title: "Share", auth: true},
    },
    {
        name: "settings",
        path: "/settings",
        component: Settings,
        meta: {title: "Settings", auth: true},
        props: {tab: 0},
    },
    {
        name: "settings_logs",
        path: "/settings/logs",
        component: Settings,
        meta: {title: "Settings", auth: true},
        props: {tab: 1},
    },
    {
        path: "*", redirect: "/photos",
    },
];
