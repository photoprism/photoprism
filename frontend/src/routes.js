import Photos from "pages/photos.vue";
import Albums from "pages/albums.vue";
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
        meta: {area: "Login"},
    },
    {
        name: "photos",
        path: "/photos",
        component: Photos,
        meta: {area: "Photos"},
    },
    {
        name: "albums",
        path: "/albums",
        component: Albums,
        meta: {area: "Albums"},
    },
    {
        name: "favorites",
        path: "/favorites",
        component: Photos,
        meta: {area: "Favorites"},
        props: {staticFilter: {favorites: true}},
    },
    {
        name: "places",
        path: "/places",
        component: Places,
        meta: {area: "Places"},
    },
    {
        name: "labels",
        path: "/labels",
        component: Labels,
        meta: {area: "Labels"},
    },
    {
        name: "events",
        path: "/events",
        component: Events,
        meta: {area: "Events"},
    },
    {
        name: "people",
        path: "/people",
        component: People,
        meta: {area: "People"},
    },
    {
        name: "filters",
        path: "/filters",
        component: Todo,
        meta: {area: "Filters"},
    },
    {
        name: "library",
        path: "/library",
        component: Library,
        meta: {area: "Library", auth: true},
        props: {tab: 0},
    },
    {
        name: "library_upload",
        path: "/library/upload",
        component: Library,
        meta: {area: "Library", auth: true},
        props: {tab: 0},
    },
    {
        name: "library_import",
        path: "/library/import",
        component: Library,
        meta: {area: "Library", auth: true},
        props: {tab: 1},
    },
    {
        name: "library_index",
        path: "/library/index",
        component: Library,
        meta: {area: "Library", auth: true},
        props: {tab: 2},
    },
    {
        name: "share",
        path: "/share",
        component: Share,
        meta: {area: "Share", auth: true},
    },
    {
        name: "settings",
        path: "/settings",
        component: Settings,
        meta: {area: "Settings", auth: true},
        props: {tab: 0},
    },
    {
        name: "settings_logs",
        path: "/settings/logs",
        component: Settings,
        meta: {area: "Settings", auth: true},
        props: {tab: 1},
    },
    {
        path: "*", redirect: "/photos",
    },
];
