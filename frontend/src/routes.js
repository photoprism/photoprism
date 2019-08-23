import Vue from "vue";

import Photos from "pages/photos.vue";
import Albums from "pages/albums.vue";
import Places from "pages/places.vue";
import Labels from "pages/labels.vue";
import Events from "pages/events.vue";
import People from "pages/people.vue";
import Library from "pages/library.vue";
import Share from "pages/share.vue";
import Settings from "pages/settings.vue";
import Todo from "pages/todo.vue";
import Login from "pages/login.vue";
import Home from "pages/home.vue";

const requireAuth = async (to, from, next) => {
    if (!!Vue.prototype.$session.getKey()) {
        // there is a session key. the server requires a password and
        // we are authenticated. so continue to the route
        return next();
    }
    if (Vue.prototype.$session.getNoAuthMode()) {
        // there was not a session key but apparently the server doesn't
        // require authentication anyway. so continue to the route
        return next();
    }
    // at this point we don't know if the server needs auth or not
    if (await Vue.prototype.$session.isAuthed()) {
        // it doesn't, so let's remember that for the next time and
        // continue to the route
        Vue.prototype.$session.setNoAuthMode(true)
        return next();
    }
    // the server requires authentication so redirect to the login screen
    next({
        name: "Login",
        query: { redirect: to.fullPath },
    });
};

const pages = [
    {
        name: "Photos",
        path: "/photos",
        component: Photos,
        meta: {area: "Photos"},
    },
    {
        name: "Albums",
        path: "/albums",
        component: Albums,
        meta: {area: "Albums"},
    },
    {
        name: "Favorites",
        path: "/favorites",
        component: Photos,
        meta: {area: "Favorites"},
        props: {staticFilter: {favorites: true}},
    },
    {
        name: "Places",
        path: "/places",
        component: Places,
        meta: {area: "Places"},
    },
    {
        name: "Labels",
        path: "/labels",
        component: Labels,
        meta: {area: "Labels"},
    },
    {
        name: "Events",
        path: "/events",
        component: Events,
        meta: {area: "Events"},
    },
    {
        name: "People",
        path: "/people",
        component: People,
        meta: {area: "People"},
    },
    {
        name: "Filters",
        path: "/filters",
        component: Todo,
        meta: {area: "Filters"},
    },
    {
        name: "Library",
        path: "/library",
        component: Library,
        meta: {area: "Library"},
    },
    {
        name: "Share",
        path: "/share",
        component: Share,
        meta: {area: "Share"},
    },
    {
        name: "Settings",
        path: "/settings",
        component: Settings,
        meta: {area: "Settings"},
    },
];

export default [
    {
        name: "Login",
        path: "/login",
        component: Login,
    },
    {
        name: "Home",
        path: "/",
        component: Home,
        beforeEnter: requireAuth,
        children: pages,
        redirect: {name: "Photos"},
    },
    {
        path: "*",
        redirect: {name: "Home"},
    },
];
