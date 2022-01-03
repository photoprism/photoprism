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

import PAccountAddDialog from "./account/add.vue";
import PAccountRemoveDialog from "./account/remove.vue";
import PAccountEditDialog from "./account/edit.vue";
import PPhotoArchiveDialog from "./photo/archive.vue";
import PPhotoAlbumDialog from "./photo/album.vue";
import PPhotoEditDialog from "./photo/edit.vue";
import PPhotoDeleteDialog from "./photo/delete.vue";
import PFileDeleteDialog from "./file/delete.vue";
import PAlbumEditDialog from "./album/edit.vue";
import PAlbumDeleteDialog from "./album/delete.vue";
import PLabelDeleteDialog from "./label/delete.vue";
import PPeopleMergeDialog from "./people/merge.vue";
import PUploadDialog from "./upload.vue";
import PVideoViewer from "./video/viewer.vue";
import PShareDialog from "./share.vue";
import PShareUploadDialog from "./share/upload.vue";
import PWebdavDialog from "./webdav.vue";
import PReloadDialog from "./reload.vue";
import PSponsorDialog from "./sponsor.vue";

const dialogs = {};

dialogs.install = (Vue) => {
  Vue.component("PAccountAddDialog", PAccountAddDialog);
  Vue.component("PAccountRemoveDialog", PAccountRemoveDialog);
  Vue.component("PAccountEditDialog", PAccountEditDialog);
  Vue.component("PPhotoArchiveDialog", PPhotoArchiveDialog);
  Vue.component("PPhotoAlbumDialog", PPhotoAlbumDialog);
  Vue.component("PPhotoEditDialog", PPhotoEditDialog);
  Vue.component("PPhotoDeleteDialog", PPhotoDeleteDialog);
  Vue.component("PFileDeleteDialog", PFileDeleteDialog);
  Vue.component("PAlbumEditDialog", PAlbumEditDialog);
  Vue.component("PAlbumDeleteDialog", PAlbumDeleteDialog);
  Vue.component("PLabelDeleteDialog", PLabelDeleteDialog);
  Vue.component("PPeopleMergeDialog", PPeopleMergeDialog);
  Vue.component("PUploadDialog", PUploadDialog);
  Vue.component("PVideoViewer", PVideoViewer);
  Vue.component("PShareDialog", PShareDialog);
  Vue.component("PShareUploadDialog", PShareUploadDialog);
  Vue.component("PWebdavDialog", PWebdavDialog);
  Vue.component("PReloadDialog", PReloadDialog);
  Vue.component("PSponsorDialog", PSponsorDialog);
};

export default dialogs;
