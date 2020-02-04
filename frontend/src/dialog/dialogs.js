import PPhotoArchiveDialog from "./p-photo-archive-dialog.vue";
import PPhotoAlbumDialog from "./p-photo-album-dialog.vue";
import PPhotoEditDialog from "./p-photo-edit-dialog.vue";
import PPhotoShareDialog from "./p-photo-share-dialog.vue";
import PAlbumDeleteDialog from "./p-album-delete-dialog.vue";
import PLabelDeleteDialog from "./p-label-delete-dialog.vue";
import PUploadDialog from "./p-upload-dialog.vue";

const dialogs = {};

dialogs.install = (Vue) => {
    Vue.component("p-photo-archive-dialog", PPhotoArchiveDialog);
    Vue.component("p-photo-album-dialog", PPhotoAlbumDialog);
    Vue.component("p-photo-edit-dialog", PPhotoEditDialog);
    Vue.component("p-photo-share-dialog", PPhotoShareDialog);
    Vue.component("p-album-delete-dialog", PAlbumDeleteDialog);
    Vue.component("p-label-delete-dialog", PLabelDeleteDialog);
    Vue.component("p-upload-dialog", PUploadDialog);
};

export default dialogs;
