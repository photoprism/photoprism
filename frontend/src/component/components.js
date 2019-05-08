import AppAlert from './app-alert.vue';
import AppNavigation from './app-navigation.vue';
import AppLoadingBar from './app-loading-bar.vue';
import AppGallery from './app-gallery.vue';
import {LMap, LTileLayer, LMarker, LControl} from 'vue2-leaflet';
import {Icon} from 'leaflet';

const components = {};

components.install = (Vue) => {
    Vue.component('app-alert', AppAlert);
    Vue.component('app-gallery', AppGallery);
    Vue.component('app-navigation', AppNavigation);
    Vue.component('app-loading-bar', AppLoadingBar);

    Vue.component('l-map', LMap);
    Vue.component('l-tile-layer', LTileLayer);
    Vue.component('l-marker', LMarker);
    Vue.component('l-control', LControl);

    delete Icon.Default.prototype._getIconUrl;

    Icon.Default.mergeOptions({
        iconRetinaUrl: require('leaflet/dist/images/marker-icon-2x.png'),
        iconUrl: require('leaflet/dist/images/marker-icon.png'),
        shadowUrl: require('leaflet/dist/images/marker-shadow.png')
    });
};

export default components;
