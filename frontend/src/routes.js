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
import Discover from "pages/discover.vue";
import Todo from "pages/todo.vue";

const c = window.clientConfig;

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
        meta: {title: "Sign In"},
    },
    {
        name: "photos",
        path: "/photos",
        component: Photos,
        meta: {title: c.subtitle},
    },
    {
        name: "albums",
        path: "/albums",
        component: Albums,
        meta: {title: "Albums"},
    },
    {
        name: "album",
        path: "/albums/:uuid",
        component: AlbumPhotos,
        meta: {title: "Album"},
    },
    {
        name: "favorites",
        path: "/favorites",
        component: Photos,
        meta: {title: "Favorites"},
        props: {staticFilter: {favorites: true}},
    },
    {
        name: "archive",
        path: "/archive",
        component: Photos,
        meta: {title: "Archive"},
        props: {staticFilter: {archived: true}},
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
        name: "library_logs",
        path: "/library/logs",
        component: Library,
        meta: {title: "Server Logs", auth: true},
        props: {tab: 3},
    },
    {
        name: "library_upload",
        path: "/library/upload",
        component: Library,
        meta: {title: "Photo Upload", auth: true},
        props: {tab: 2},
    },
    {
        name: "library_import",
        path: "/library/import",
        component: Library,
        meta: {title: "Import Photos", auth: true},
        props: {tab: 1},
    },
    {
        name: "library",
        path: "/library",
        component: Library,
        meta: {title: "Photo Library", auth: true},
        props: {tab: 0},
    },
    {
        name: "share",
        path: "/share",
        component: Share,
        meta: {title: "Share with friends", auth: true},
    },
    {
        name: "settings",
        path: "/settings",
        component: Settings,
        meta: {title: "Application Settings", auth: true},
        props: {tab: 0},
    },
    {
        name: "discover",
        path: "/discover",
        component: Discover,
        meta: {title: "Discover", auth: false},
        props: {tab: 0},
    },
    {
        name: "discover_similar",
        path: "/discover/similar",
        component: Discover,
        meta: {title: "Discover", auth: false},
        props: {tab: 1},
    },
    {
        name: "discover_season",
        path: "/discover/season",
        component: Discover,
        meta: {title: "Discover", auth: false},
        props: {tab: 2},
    },
    {
        name: "discover_random",
        path: "/discover/random",
        component: Discover,
        meta: {title: "Discover", auth: false},
        props: {tab: 3},
    },
    {
        path: "*", redirect: "/photos",
    },
];
