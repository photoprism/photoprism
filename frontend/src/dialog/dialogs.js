import PPhotoDeleteDialog from "./p-photo-delete-dialog.vue";
import PPhotoAlbumDialog from "./p-photo-album-dialog.vue";
import PPhotoEditDialog from "./p-photo-edit-dialog.vue";
import PAlbumDeleteDialog from "./p-album-delete-dialog.vue";

const dialogs = {};

dialogs.install = (Vue) => {
    Vue.component("p-photo-delete-dialog", PPhotoDeleteDialog);
    Vue.component("p-photo-album-dialog", PPhotoAlbumDialog);
    Vue.component("p-photo-edit-dialog", PPhotoEditDialog);
    Vue.component("p-album-delete-dialog", PAlbumDeleteDialog);
};

export default dialogs;
