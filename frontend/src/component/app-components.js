import AppNavigation from './app-navigation.vue';
import AppLoadingBar from './app-loading-bar.vue';

const components = {};

components.install = (Vue) => {
    Vue.component('app-navigation', AppNavigation);
    Vue.component('app-loading-bar', AppLoadingBar);
};

export default components;