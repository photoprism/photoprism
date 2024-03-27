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

import PServiceAddDialog from "dialog/service/add.vue";
import PServiceRemoveDialog from "dialog/service/remove.vue";
import PServiceEditDialog from "dialog/service/edit.vue";
import PPhotoArchiveDialog from "dialog/photo/archive.vue";
import PPhotoAlbumDialog from "dialog/photo/album.vue";
import PPhotoEditDialog from "dialog/photo/edit.vue";
import PPhotoDeleteDialog from "dialog/photo/delete.vue";
import PFileDeleteDialog from "dialog/file/delete.vue";
import PAlbumEditDialog from "dialog/album/edit.vue";
import PAlbumDeleteDialog from "dialog/album/delete.vue";
import PLabelDeleteDialog from "dialog/label/delete.vue";
import PPeopleMergeDialog from "dialog/people/merge.vue";
import PUploadDialog from "dialog/upload.vue";
import PVideoViewer from "dialog/video/viewer.vue";
import PShareDialog from "dialog/share.vue";
import PShareUploadDialog from "dialog/share/upload.vue";
import PWebdavDialog from "dialog/webdav.vue";
import PReloadDialog from "dialog/reload.vue";
import PSponsorDialog from "dialog/sponsor.vue";
import PConfirmDialog from "dialog/confirm.vue";
import PAccountAppsDialog from "dialog/account/apps.vue";
import PAccountPasscodeDialog from "dialog/account/passcode.vue";
import PAccountPasswordDialog from "dialog/account/password.vue";

const dialogs = {};

dialogs.install = (Vue) => {
  Vue.component("PServiceAddDialog", PServiceAddDialog);
  Vue.component("PServiceRemoveDialog", PServiceRemoveDialog);
  Vue.component("PServiceEditDialog", PServiceEditDialog);
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
  Vue.component("PConfirmDialog", PConfirmDialog);
  Vue.component("PAccountAppsDialog", PAccountAppsDialog);
  Vue.component("PAccountPasscodeDialog", PAccountPasscodeDialog);
  Vue.component("PAccountPasswordDialog", PAccountPasswordDialog);
};

export default dialogs;
