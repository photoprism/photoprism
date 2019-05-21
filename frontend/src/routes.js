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
import Tags from "pages/tags.vue";
import Todo from "pages/todo.vue";
import Calendar from "pages/calendar.vue";

export default [
    { name: "Home", path: "/", redirect: "/photos" },
    { name: "Photos", path: "/photos", component: PhotosPage },
    { name: "Places", path: "/places", component: PlacesPage },
    { name: "PhotosEdit", path: "/photosEdit", component: PhotosEdit },
    { name: "Filters", path: "/filters", component: Todo },
    { name: "Calendar", path: "/calendar", component: Calendar },
    { name: "Tags", path: "/tags", component: Tags },
    { name: "Bookmarks", path: "/bookmarks", component: Todo },
    { name: "Favorites", path: "/favorites", component: Todo },
    { name: "Albums", path: "/albums", component: Albums },
    { name: "Albums2", path: "/albums2", component: Albums2 },
    { name: "Import", path: "/import", component: Import },
    { name: "Import2", path: "/import2", component: Import2 },
    { name: "Import3", path: "/import3", component: Import3 },
    { name: "Export", path: "/export", component: Export },
    { name: "Settings", path: "/settings", component: Settings },
    { path: "*", redirect: "/photos" },
];
