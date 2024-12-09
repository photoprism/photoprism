/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import PNotify from "component/notify.vue";
import PNavigation from "component/navigation.vue";
import PScrollTop from "component/scroll-top.vue";
import PLoadingBar from "component/loading-bar.vue";
import PPhotoViewer from "component/photo-viewer.vue";
import PVideoPlayer from "component/video/player.vue";
import PPhotoToolbar from "component/photo/toolbar.vue";
import PPhotoCards from "component/photo/cards.vue";
import PPhotoMosaic from "component/photo/mosaic.vue";
import PPhotoList from "component/photo/list.vue";
import PPhotoPreview from "component/photo/preview.vue";
import PPhotoClipboard from "component/photo/clipboard.vue";
import PAlbumClipboard from "component/album/clipboard.vue";
import PAlbumToolbar from "component/album/toolbar.vue";
import PLabelClipboard from "component/label/clipboard.vue";
import PFileClipboard from "component/file/clipboard.vue";
import PSubjectClipboard from "component/subject/clipboard.vue";
import PAuthHeader from "component/auth/header.vue";
import PAuthFooter from "component/auth/footer.vue";
import PAboutFooter from "component/footer.vue";
import IconLivePhoto from "component/icon/live-photo.vue";
import IconSponsor from "component/icon/sponsor.vue";
import IconPrism from "component/icon/prism.vue";

export function installComponents(app) {
  app.component("PNotify", PNotify);
  app.component("PNavigation", PNavigation);
  app.component("PScrollTop", PScrollTop);
  app.component("PLoadingBar", PLoadingBar);
  app.component("PVideoPlayer", PVideoPlayer);
  app.component("PPhotoViewer", PPhotoViewer);
  app.component("PPhotoToolbar", PPhotoToolbar);
  app.component("PPhotoCards", PPhotoCards);
  app.component("PPhotoMosaic", PPhotoMosaic);
  app.component("PPhotoList", PPhotoList);
  app.component("PPhotoPreview", PPhotoPreview);
  app.component("PPhotoClipboard", PPhotoClipboard);
  app.component("PAlbumClipboard", PAlbumClipboard);
  app.component("PAlbumToolbar", PAlbumToolbar);
  app.component("PLabelClipboard", PLabelClipboard);
  app.component("PFileClipboard", PFileClipboard);
  app.component("PSubjectClipboard", PSubjectClipboard);
  app.component("PAuthHeader", PAuthHeader);
  app.component("PAuthFooter", PAuthFooter);
  app.component("PAboutFooter", PAboutFooter);
  app.component("IconLivePhoto", IconLivePhoto);
  app.component("IconSponsor", IconSponsor);
  app.component("IconPrism", IconPrism);
}
