import Vue from 'vue';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import '../css/app.css';
import App from 'app/main.vue';
import routes from 'app/routes';
import Api from 'common/api';
import Config from 'common/config';
import AppComponents from 'component/app-components';
import Alert from 'common/alert';
import Session from 'common/session';
import Event from 'pubsub-js';
import Moment from 'vue-moment';
import InfiniteScroll from 'vue-infinite-scroll';
import VueTruncate from 'vue-truncate-filter';

const session = new Session(window.localStorage);
const config = new Config(window.localStorage, window.appConfig);

Vue.prototype.$event = Event;
Vue.prototype.$alert = Alert;
Vue.prototype.$session = session;
Vue.prototype.$api = Api;
Vue.prototype.$config = config;

Vue.use(Vuetify, {
    theme: {
        primary: '#FFD600',
        secondary: '#b0bec5',
        accent: '#00B8D4',
        error: '#E57373',
        info: '#00B8D4',
        success: '#00BFA5',
        warning: '#FFD600',
        delete: '#E57373',
    },
});

Vue.use(Moment);
Vue.use(InfiniteScroll);
Vue.use(VueTruncate);
Vue.use(AppComponents);
Vue.use(Router);

const router = new Router({
    routes,
    mode: 'history',
    saveScrollPosition: true,
});

/* eslint-disable no-unused-vars */
const app = new Vue({
    el: '#app',
    router,
    render: h => h(App),
});