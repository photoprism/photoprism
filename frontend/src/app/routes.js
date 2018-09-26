import Photos from 'app/pages/photos.vue';
import Albums from 'app/pages/albums.vue';
import Import from 'app/pages/import.vue';
import Export from 'app/pages/export.vue';
import Settings from 'app/pages/settings.vue';
import Todo from 'app/pages/todo.vue';

export default [
    { name: 'Home', path: '/', redirect: '/photos' },
    { name: 'Photos', path: '/photos', component: Photos },
    { name: 'Filters', path: '/filters', component: Todo },
    { name: 'Calendar', path: '/calendar', component: Todo },
    { name: 'Tags', path: '/tags', component: Todo },
    { name: 'Bookmarks', path: '/bookmarks', component: Todo },
    { name: 'Favorites', path: '/favorites', component: Todo },
    { name: 'Places', path: '/places', component: Todo },
    { name: 'Albums', path: '/albums', component: Albums },
    { name: 'Import', path: '/import', component: Import },
    { name: 'Export', path: '/export', component: Export },
    { name: 'Settings', path: '/settings', component: Settings },
    { path: '*', redirect: '/photos' },
];
