/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

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

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import PNotify from "component/notify.vue";
import PNavigation from "./navigation.vue";
import PScrollTop from "component/scroll-top.vue";
import PLoadingBar from "component/loading-bar.vue";
import PVideoPlayer from "component/video/player.vue";
import PPhotoViewer from "component/photo/viewer.vue";
import PPhotoCards from "./photo/cards.vue";
import PPhotoMosaic from "./photo/mosaic.vue";
import PPhotoList from "./photo/list.vue";
import PPhotoClipboard from "./photo/clipboard.vue";
import PAlbumClipboard from "./album/clipboard.vue";

const components = {};

components.install = (Vue) => {
  Vue.component("PNotify", PNotify);
  Vue.component("PNavigation", PNavigation);
  Vue.component("PScrollTop", PScrollTop);
  Vue.component("PLoadingBar", PLoadingBar);
  Vue.component("PVideoPlayer", PVideoPlayer);
  Vue.component("PPhotoViewer", PPhotoViewer);
  Vue.component("PPhotoCards", PPhotoCards);
  Vue.component("PPhotoMosaic", PPhotoMosaic);
  Vue.component("PPhotoList", PPhotoList);
  Vue.component("PPhotoClipboard", PPhotoClipboard);
  Vue.component("PAlbumClipboard", PAlbumClipboard);
};

export default components;
