/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

import PNotify from "component/notify.vue";
import PNavigation from "navigation.vue";
import PScrollTop from "component/scroll-top.vue";
import PLoadingBar from "component/loading-bar.vue";
import PVideoPlayer from "component/video/player.vue";
import PPhotoViewer from "component/photo/viewer.vue";
import PPhotoCards from "component/photo/cards.vue";
import PPhotoMosaic from "component/photo/mosaic.vue";
import PPhotoList from "component/photo/list.vue";
import PPhotoClipboard from "component/photo/clipboard.vue";
import PAlbumClipboard from "component/album/clipboard.vue";

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
