import PAccountAddDialog from "./account/add.vue";
import PAccountRemoveDialog from "./account/remove.vue";
import PAccountEditDialog from "./account/edit.vue";
import PPhotoArchiveDialog from "./p-photo-archive-dialog.vue";
import PPhotoAlbumDialog from "./p-photo-album-dialog.vue";
import PPhotoEditDialog from "./p-photo-edit-dialog.vue";
import PAlbumDeleteDialog from "./album/delete.vue";
import PLabelDeleteDialog from "./label/delete.vue";
import PUploadDialog from "./p-upload-dialog.vue";
import PVideoDialog from "./p-video-dialog.vue";
import PShareDialog from "./p-share-dialog.vue";
import PShareUploadDialog from "./share/upload.vue";

const dialogs = {};

dialogs.install = (Vue) => {
    Vue.component("p-account-add-dialog", PAccountAddDialog);
    Vue.component("p-account-remove-dialog", PAccountRemoveDialog);
    Vue.component("p-account-edit-dialog", PAccountEditDialog);
    Vue.component("p-photo-archive-dialog", PPhotoArchiveDialog);
    Vue.component("p-photo-album-dialog", PPhotoAlbumDialog);
    Vue.component("p-photo-edit-dialog", PPhotoEditDialog);
    Vue.component("p-album-delete-dialog", PAlbumDeleteDialog);
    Vue.component("p-label-delete-dialog", PLabelDeleteDialog);
    Vue.component("p-upload-dialog", PUploadDialog);
    Vue.component("p-video-dialog", PVideoDialog);
    Vue.component("p-share-dialog", PShareDialog);
    Vue.component("p-share-upload-dialog", PShareUploadDialog);
};

export default dialogs;
