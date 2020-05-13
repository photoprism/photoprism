import PNotify from "./p-notify.vue";
import PNavigation from "./p-navigation.vue";
import PLoadingBar from "./p-loading-bar.vue";
import PVideoPlayer from "./p-video-player.vue";
import PPhotoViewer from "./p-photo-viewer.vue";
import PPhotoCards from "./p-photo-cards.vue";
import PPhotoMosaic from "./p-photo-mosaic.vue";
import PPhotoList from "./p-photo-list.vue";
import PPhotoSearch from "./p-photo-search.vue";
import PPhotoClipboard from "./p-photo-clipboard.vue";
import PLabelClipboard from "./p-label-clipboard.vue";
import PAlbumClipboard from "./p-album-clipboard.vue";
import PAlbumToolbar from "./p-album-toolbar.vue";
import PScrollTop from "./p-scroll-top.vue";

const components = {};

components.install = (Vue) => {
    Vue.component("p-notify", PNotify);
    Vue.component("p-navigation", PNavigation);
    Vue.component("p-loading-bar", PLoadingBar);
    Vue.component("p-video-player", PVideoPlayer);
    Vue.component("p-photo-viewer", PPhotoViewer);
    Vue.component("p-photo-cards", PPhotoCards);
    Vue.component("p-photo-mosaic", PPhotoMosaic);
    Vue.component("p-photo-list", PPhotoList);
    Vue.component("p-photo-search", PPhotoSearch);
    Vue.component("p-photo-clipboard", PPhotoClipboard);
    Vue.component("p-label-clipboard", PLabelClipboard);
    Vue.component("p-album-clipboard", PAlbumClipboard);
    Vue.component("p-album-toolbar", PAlbumToolbar);
    Vue.component("p-scroll-top", PScrollTop);
};

export default components;
