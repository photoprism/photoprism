import Photos from 'app/pages/photos.vue';
import Albums from 'app/pages/albums.vue';
import Import from 'app/pages/import.vue';
import Export from 'app/pages/export.vue';
import Settings from 'app/pages/settings.vue';

export default [
    { path: '/', redirect: '/photos' },
    { path: '/photos', component: Photos },
    { path: '/albums', component: Albums },
    { path: '/import', component: Import },
    { path: '/export', component: Export },
    { path: '/settings', component: Settings },
    { path: '*', redirect: '/photos' },
];
