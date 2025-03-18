<template>
  <div id="p-dialogs">
    <p-photo-edit-dialog
      :visible="edit.visible"
      :selection="edit.selection"
      :index="edit.index"
      :album="edit.album"
      :tab="edit.tab"
      @close="closeEditDialog"
    ></p-photo-edit-dialog>
    <p-photo-upload-dialog
      :visible="upload.visible"
      :data="upload.data"
      @close="closeUploadDialog"
      @confirm="closeUploadDialog"
    ></p-photo-upload-dialog>
    <p-update :visible="update.visible" @close="closeUpdateDialog"></p-update>
    <p-lightbox @enter="onLightboxEnter" @leave="onLightboxLeave"></p-lightbox>
  </div>
</template>
<script>
import Album from "model/album";

import PPhotoEditDialog from "component/photo/edit/dialog.vue";
import PPhotoUploadDialog from "component/photo/upload/dialog.vue";
import PUpdate from "component/update.vue";
import PLightbox from "component/lightbox.vue";

export default {
  name: "PDialogs",
  components: {
    PPhotoEditDialog,
    PPhotoUploadDialog,
    PUpdate,
    PLightbox,
  },
  data() {
    return {
      edit: {
        visible: false,
        album: null,
        selection: [],
        index: 0,
        tab: "",
      },
      upload: {
        visible: false,
        data: {},
      },
      update: {
        visible: false,
      },
      lightbox: {
        visible: false,
      },
      subscriptions: [],
    };
  },
  created() {
    // Opens the photo edit dialog.
    this.subscriptions.push(
      this.$event.subscribe("dialog.edit", (ev, data) => {
        this.onEdit(data);
      })
    );

    // Opens the web upload dialog.
    this.subscriptions.push(
      this.$event.subscribe("dialog.upload", (ev, data) => {
        this.onUpload(data);
      })
    );

    // Opens the update dialog so that users can reload the UI after updates.
    this.subscriptions.push(
      this.$event.subscribe("dialog.update", () => {
        this.onUpdate();
      })
    );
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    hasAuth() {
      return this.$session.auth || this.isPublic;
    },
    isReadOnly() {
      return this.$config.get("readonly");
    },
    onEdit(data) {
      if (this.edit.visible || !this.hasAuth()) {
        return;
      }

      this.edit.index = data.index;
      this.edit.selection = data.selection;
      this.edit.album = data.album;
      this.edit.tab = data?.tab ? data.tab : "";
      this.edit.visible = true;
    },
    closeEditDialog() {
      if (this.edit.visible) {
        this.edit.visible = false;
      }
    },
    onUpload(data) {
      if (this.upload.visible || !this.hasAuth() || this.isReadOnly() || !this.$config.feature("upload")) {
        return;
      }

      if (this.$route.name === "album" && this.$route.params?.album) {
        return new Album()
          .find(this.$route.params?.album)
          .then((m) => {
            this.showUploadDialog(Object.assign({ albums: [m] }, data));
          })
          .catch(() => {
            this.showUploadDialog(data);
          });
      } else {
        this.showUploadDialog(data);
      }
    },
    showUploadDialog(data) {
      this.upload.data = Object.assign({ albums: [] }, data);
      this.upload.visible = true;
    },
    closeUploadDialog() {
      if (this.upload.visible) {
        this.upload.visible = false;
      }
    },
    onUpdate() {
      if (this.$view.preventNavigation || this.update.visible || this.lightbox.visible) {
        return;
      }

      this.update.visible = true;
    },
    closeUpdateDialog() {
      if (this.update.visible) {
        this.update.visible = false;
      }
    },
    onLightboxEnter() {
      this.closeUpdateDialog();
      this.lightbox.visible = true;
    },
    onLightboxLeave() {
      this.lightbox.visible = false;
    },
  },
};
</script>
