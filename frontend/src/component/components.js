import PNotify from "./p-notify.vue";
import PNavigation from "./p-navigation.vue";
import PLoadingBar from "./p-loading-bar.vue";
import PPhotoSearch from "./p-photo-search.vue";
import PPhotoClipboard from "./p-photo-clipboard.vue";
import PPhotoDetails from "./p-photo-details.vue";
import PPhotoTiles from "./p-photo-tiles.vue";
import PPhotoMosaic from "./p-photo-mosaic.vue";
import PPhotoList from "./p-photo-list.vue";
import PAlbumPhotoSearch from "./album/p-photo-search.vue";
import PAlbumPhotoClipboard from "./album/p-photo-clipboard.vue";
import PAlbumPhotoDetails from "./album/p-photo-details.vue";
import PAlbumPhotoTiles from "./album/p-photo-tiles.vue";
import PAlbumPhotoMosaic from "./album/p-photo-mosaic.vue";
import PAlbumPhotoList from "./album/p-photo-list.vue";
import PPhotoViewer from "./p-photo-viewer.vue";
import PScrollTop from "./p-scroll-top.vue";

const components = {};

components.install = (Vue) => {
    Vue.component("p-notify", PNotify);
    Vue.component("p-navigation", PNavigation);
    Vue.component("p-loading-bar", PLoadingBar);
    Vue.component("p-photo-viewer", PPhotoViewer);
    Vue.component("p-photo-details", PPhotoDetails);
    Vue.component("p-photo-tiles", PPhotoTiles);
    Vue.component("p-photo-mosaic", PPhotoMosaic);
    Vue.component("p-photo-list", PPhotoList);
    Vue.component("p-photo-search", PPhotoSearch);
    Vue.component("p-photo-clipboard", PPhotoClipboard);
    Vue.component("p-album-photo-details", PAlbumPhotoDetails);
    Vue.component("p-album-photo-tiles", PAlbumPhotoTiles);
    Vue.component("p-album-photo-mosaic", PAlbumPhotoMosaic);
    Vue.component("p-album-photo-list", PAlbumPhotoList);
    Vue.component("p-album-photo-search", PAlbumPhotoSearch);
    Vue.component("p-album-photo-clipboard", PAlbumPhotoClipboard);
    Vue.component("p-scroll-top", PScrollTop);
};

export default components;
