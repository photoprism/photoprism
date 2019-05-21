import PAlert from "./p-alert.vue";
import PNavigation from "./p-navigation.vue";
import PLoadingBar from "./p-loading-bar.vue";
import PPhotoDetails from "./p-photo-details.vue";
import PPhotoTiles from "./p-photo-tiles.vue";
import PPhotoMosaic from "./p-photo-mosaic.vue";
import PPhotoList from "./p-photo-list.vue";
import PPhotoViewer from "./p-photo-viewer.vue";
import PPhotoSearch from "./p-photo-search.vue";
import PPhotoClipboard from "./p-photo-clipboard.vue";

const components = {};

components.install = (Vue) => {
    Vue.component("p-alert", PAlert);
    Vue.component("p-navigation", PNavigation);
    Vue.component("p-loading-bar", PLoadingBar);
    Vue.component("p-photo-details", PPhotoDetails);
    Vue.component("p-photo-tiles", PPhotoTiles);
    Vue.component("p-photo-mosaic", PPhotoMosaic);
    Vue.component("p-photo-list", PPhotoList);
    Vue.component("p-photo-viewer", PPhotoViewer);
    Vue.component("p-photo-search", PPhotoSearch);
    Vue.component("p-photo-clipboard", PPhotoClipboard);
};

export default components;
