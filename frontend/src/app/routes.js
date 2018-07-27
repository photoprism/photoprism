import Welcome from 'app/pages/welcome.vue';

export default [
    { path: '/', redirect: '/welcome' },
    { path: '/welcome', component: Welcome },
    { path: '*', redirect: '/welcome' },
];
