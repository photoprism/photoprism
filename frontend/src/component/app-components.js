import AppAlert from './app-alert.vue';
import AppNavigation from './app-navigation.vue';
import AppLoadingBar from './app-loading-bar.vue';
import PhotoSwipe from './photoswipe.vue';

const components = {};

components.install = (Vue) => {
    Vue.component('app-alert', AppAlert);
    Vue.component('photoswipe', PhotoSwipe);
    Vue.component('app-navigation', AppNavigation);
    Vue.component('app-loading-bar', AppLoadingBar);
};

export default components;
