/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import PNotify from "./notify.vue";
import PNavigation from "./navigation.vue";
import PScrollTop from "./scroll-top.vue";
import PLoadingBar from "./loading-bar.vue";
import PVideoPlayer from "./video/player.vue";
import PPhotoViewer from "./photo/viewer.vue";
import PPhotoToolbar from "./photo/toolbar.vue";
import PPhotoCards from "./photo/cards.vue";
import PPhotoMosaic from "./photo/mosaic.vue";
import PPhotoList from "./photo/list.vue";
import PPhotoClipboard from "./photo/clipboard.vue";
import PAlbumClipboard from "./album/clipboard.vue";
import PAlbumToolbar from "./album/toolbar.vue";
import PLabelClipboard from "./label/clipboard.vue";
import PFileClipboard from "./file/clipboard.vue";
import PAboutFooter from "./footer.vue";

const components = {};

components.install = (Vue) => {
    Vue.component("p-notify", PNotify);
    Vue.component("p-navigation", PNavigation);
    Vue.component("p-scroll-top", PScrollTop);
    Vue.component("p-loading-bar", PLoadingBar);
    Vue.component("p-video-player", PVideoPlayer);
    Vue.component("p-photo-viewer", PPhotoViewer);
    Vue.component("p-photo-toolbar", PPhotoToolbar);
    Vue.component("p-photo-cards", PPhotoCards);
    Vue.component("p-photo-mosaic", PPhotoMosaic);
    Vue.component("p-photo-list", PPhotoList);
    Vue.component("p-photo-clipboard", PPhotoClipboard);
    Vue.component("p-album-clipboard", PAlbumClipboard);
    Vue.component("p-album-toolbar", PAlbumToolbar);
    Vue.component("p-label-clipboard", PLabelClipboard);
    Vue.component("p-file-clipboard", PFileClipboard);
    Vue.component("p-about-footer", PAboutFooter);
};

export default components;
