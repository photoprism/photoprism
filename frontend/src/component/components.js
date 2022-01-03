/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.org>

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

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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
import PSubjectClipboard from "./subject/clipboard.vue";
import PAboutFooter from "./footer.vue";

const components = {};

components.install = (Vue) => {
  Vue.component("PNotify", PNotify);
  Vue.component("PNavigation", PNavigation);
  Vue.component("PScrollTop", PScrollTop);
  Vue.component("PLoadingBar", PLoadingBar);
  Vue.component("PVideoPlayer", PVideoPlayer);
  Vue.component("PPhotoViewer", PPhotoViewer);
  Vue.component("PPhotoToolbar", PPhotoToolbar);
  Vue.component("PPhotoCards", PPhotoCards);
  Vue.component("PPhotoMosaic", PPhotoMosaic);
  Vue.component("PPhotoList", PPhotoList);
  Vue.component("PPhotoClipboard", PPhotoClipboard);
  Vue.component("PAlbumClipboard", PAlbumClipboard);
  Vue.component("PAlbumToolbar", PAlbumToolbar);
  Vue.component("PLabelClipboard", PLabelClipboard);
  Vue.component("PFileClipboard", PFileClipboard);
  Vue.component("PSubjectClipboard", PSubjectClipboard);
  Vue.component("PAboutFooter", PAboutFooter);
};

export default components;
