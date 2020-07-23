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

import PAccountAddDialog from "./account/add.vue";
import PAccountRemoveDialog from "./account/remove.vue";
import PAccountEditDialog from "./account/edit.vue";
import PPhotoArchiveDialog from "./photo/archive.vue";
import PPhotoAlbumDialog from "./photo/album.vue";
import PPhotoEditDialog from "./photo/edit.vue";
import PAlbumEditDialog from "./album/edit.vue";
import PAlbumDeleteDialog from "./album/delete.vue";
import PLabelDeleteDialog from "./label/delete.vue";
import PUploadDialog from "./upload.vue";
import PVideoDialog from "./video.vue";
import PShareDialog from "./share.vue";
import PShareUploadDialog from "./share/upload.vue";
import PWebdavDialog from "./webdav.vue";
import PReloadDialog from "./reload.vue";

const dialogs = {};

dialogs.install = (Vue) => {
    Vue.component("p-account-add-dialog", PAccountAddDialog);
    Vue.component("p-account-remove-dialog", PAccountRemoveDialog);
    Vue.component("p-account-edit-dialog", PAccountEditDialog);
    Vue.component("p-photo-archive-dialog", PPhotoArchiveDialog);
    Vue.component("p-photo-album-dialog", PPhotoAlbumDialog);
    Vue.component("p-photo-edit-dialog", PPhotoEditDialog);
    Vue.component("p-album-edit-dialog", PAlbumEditDialog);
    Vue.component("p-album-delete-dialog", PAlbumDeleteDialog);
    Vue.component("p-label-delete-dialog", PLabelDeleteDialog);
    Vue.component("p-upload-dialog", PUploadDialog);
    Vue.component("p-video-dialog", PVideoDialog);
    Vue.component("p-share-dialog", PShareDialog);
    Vue.component("p-share-upload-dialog", PShareUploadDialog);
    Vue.component("p-webdav-dialog", PWebdavDialog);
    Vue.component("p-reload-dialog", PReloadDialog);
};

export default dialogs;
