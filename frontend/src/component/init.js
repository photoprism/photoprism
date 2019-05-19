import AppAlert from "./app-alert.vue";
import AppNavigation from "./app-navigation.vue";
import AppLoadingBar from "./app-loading-bar.vue";
import AppGallery from "./app-gallery.vue";
import AppPhotoDetails from "./app-photo-details.vue";
import AppPhotoTiles from "./app-photo-tiles.vue";
import AppPhotoMosaic from "./app-photo-mosaic.vue";
import AppPhotoList from "./app-photo-list.vue";

const components = {};

components.install = (Vue) => {
    Vue.component("app-alert", AppAlert);
    Vue.component("app-gallery", AppGallery);
    Vue.component("app-navigation", AppNavigation);
    Vue.component("app-loading-bar", AppLoadingBar);
    Vue.component("app-photo-details", AppPhotoDetails);
    Vue.component("app-photo-tiles", AppPhotoTiles);
    Vue.component("app-photo-mosaic", AppPhotoMosaic);
    Vue.component("app-photo-list", AppPhotoList);
};

export default components;
