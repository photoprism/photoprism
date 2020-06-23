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

import PNotify from "./p-notify.vue";
import PNavigation from "./p-navigation.vue";
import PLoadingBar from "./p-loading-bar.vue";
import PVideoPlayer from "./p-video-player.vue";
import PPhotoViewer from "./p-photo-viewer.vue";
import PPhotoCards from "./p-photo-cards.vue";
import PPhotoMosaic from "./p-photo-mosaic.vue";
import PPhotoList from "./p-photo-list.vue";
import PPhotoClipboard from "./p-photo-clipboard.vue";
import PLabelClipboard from "./p-label-clipboard.vue";
import PFileClipboard from "./p-file-clipboard.vue";
import PAlbumClipboard from "./p-album-clipboard.vue";
import PAlbumToolbar from "./p-album-toolbar.vue";
import PPhotoToolbar from "./p-photo-toolbar.vue";
import PScrollTop from "./p-scroll-top.vue";

const components = {};

components.install = (Vue) => {
    Vue.component("p-notify", PNotify);
    Vue.component("p-navigation", PNavigation);
    Vue.component("p-loading-bar", PLoadingBar);
    Vue.component("p-video-player", PVideoPlayer);
    Vue.component("p-photo-viewer", PPhotoViewer);
    Vue.component("p-photo-cards", PPhotoCards);
    Vue.component("p-photo-mosaic", PPhotoMosaic);
    Vue.component("p-photo-list", PPhotoList);
    Vue.component("p-photo-clipboard", PPhotoClipboard);
    Vue.component("p-label-clipboard", PLabelClipboard);
    Vue.component("p-file-clipboard", PFileClipboard);
    Vue.component("p-album-clipboard", PAlbumClipboard);
    Vue.component("p-album-toolbar", PAlbumToolbar);
    Vue.component("p-photo-toolbar", PPhotoToolbar);
    Vue.component("p-scroll-top", PScrollTop);
};

export default components;
