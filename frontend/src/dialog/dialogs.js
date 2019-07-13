import PPhotoDeleteDialog from "./p-photo-delete-dialog.vue";

const dialogs = {};

dialogs.install = (Vue) => {
    Vue.component("p-photo-delete-dialog", PPhotoDeleteDialog);
};

export default dialogs;
