import Vue from 'vue';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import '../css/app.css';
import App from 'app/main.vue';
import routes from 'app/routes';
import Api from 'common/api';
import AppComponents from 'component/app-components';
import Alert from 'common/alert';
import Session from 'common/session';
import Event from 'pubsub-js';
import Moment from 'vue-moment';

const session = new Session(window.localStorage);

Vue.prototype.$event = Event;
Vue.prototype.$alert = Alert;
Vue.prototype.$session = session;
Vue.prototype.$api = Api;
Vue.prototype.$config = window.appConfig;

Vue.use(Vuetify, {
    theme: {
        primary: '#FDD835',
        secondary: '#b0bec5',
        accent: '#8c9eff',
        error: '#b71c1c',
    },
});

Vue.use(Moment);
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
