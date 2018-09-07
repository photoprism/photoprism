import Photos from 'app/pages/photos.vue';
import Albums from 'app/pages/albums.vue';
import Import from 'app/pages/import.vue';
import Export from 'app/pages/export.vue';
import Settings from 'app/pages/settings.vue';
import Todo from 'app/pages/todo.vue';

export default [
    { name: 'home', path: '/', redirect: '/photos' },
    { name: 'photos', path: '/photos', component: Photos },
    { name: 'filters', path: '/filters', component: Todo },
    { name: 'bookmarks', path: '/bookmarks', component: Todo },
    { name: 'favorites', path: '/favorites', component: Todo },
    { name: 'places', path: '/places', component: Todo },
    { name: 'albums', path: '/albums', component: Albums },
    { name: 'import', path: '/import', component: Import },
    { name: 'export', path: '/export', component: Export },
    { name: 'settings', path: '/settings', component: Settings },
    { path: '*', redirect: '/photos' },
];
