import PhotosPage from "pages/photos.vue";
import PlacesPage from "pages/places.vue";
import PhotosEdit from "pages/photosEdit.vue";
import Albums from "pages/albums.vue";
import Albums2 from "pages/albums2.vue";
import Import2 from "pages/import2.vue";
import Import3 from "pages/import3.vue";
import Import from "pages/import.vue";
import Export from "pages/export.vue";
import Settings from "pages/settings.vue";
import Labels from "pages/labels.vue";
import Todo from "pages/todo.vue";
import Events from "pages/events.vue";
import People from "pages/people.vue";

export default [
    {
        name: "Home",
        path: "/",
        redirect: "/photos",
    },
    {
        name: "Photos",
        path: "/photos",
        component: PhotosPage,
        meta: {area: "Photos"},
    },
    {
        name: "Favorites",
        path: "/favorites",
        component: PhotosPage,
        meta: {area: "Favorites"},
        props: {staticFilter: {favorites: true}},
    },
    {
        name: "Places",
        path: "/places",
        component: PlacesPage,
        meta: {area: "Places"},
    },
    {
        name: "Labels",
        path: "/labels",
        component: Labels,
        meta: {area: "Labels"},
    },
    {
        name: "Events",
        path: "/events",
        component: Events,
        meta: {area: "Events"},
    },
    {
        name: "People",
        path: "/people",
        component: People,
        meta: {area: "People"},
    },
    {
        name: "PhotosEdit",
        path: "/photosEdit",
        component: Todo,
        meta: {area: "Photos"},
    },
    {
        name: "Filters",
        path: "/filters",
        component: Todo,
        meta: {area: "Filters"},
    },
    {
        name: "Albums",
        path: "/albums",
        component: Todo,
        meta: {area: "Albums"},
    },
    {
        name: "Albums2",
        path: "/albums2",
        component: Albums2,
        meta: {area: "Albums"},
    },
    {
        name: "Import",
        path: "/import",
        component: Todo,
        meta: {area: "Import"},
    },
    {
        name: "Import2",
        path: "/import2",
        component: Import2, meta: {area: "Import"},
    },
    {
        name: "Import3",
        path: "/import3",
        component: Import3, meta: {area: "Import"},
    },
    {
        name: "Export",
        path: "/export",
        component: Todo,
        meta: {area: "Export"},
    },
    {
        name: "Settings",
        path: "/settings",
        component: Todo,
        meta: {area: "Settings"},
    },
    {
        path: "*", redirect: "/photos",
    },
];
