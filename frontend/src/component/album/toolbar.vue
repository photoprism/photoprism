<template>
  <v-form
    ref="form"
    validate-on="invalid-input"
    autocomplete="off"
    class="p-photo-toolbar p-album-toolbar"
    accept-charset="UTF-8"
    @submit.prevent="updateQuery()"
  >
    <v-toolbar
      flat
      :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
      class="page-toolbar"
      color="secondary"
    >
      <v-toolbar-title :title="album.Title" class="page__title">
        <router-link :to="{ name: collectionRoute }" class="hidden-xs">
          {{ T(collectionTitle) }}
          <v-icon>{{ navIcon }}</v-icon>
        </router-link>
        <router-link :to="{ name: collectionRoute }">
          {{ album.Title }}
        </router-link>
      </v-toolbar-title>

      <v-btn-toggle
        :model-value="settings.view"
        :title="$gettext('Toggle View')"
        :density="$vuetify.display.smAndDown ? 'comfortable' : 'default'"
        base-color="secondary"
        variant="flat"
        rounded="pill"
        mandatory
        border
        group
        class="ms-1"
      >
        <v-btn value="cards" icon="mdi-view-column" class="ps-1 action-view-cards" @click="setView('cards')"></v-btn>
        <v-btn
          v-if="listView"
          value="list"
          icon="mdi-view-list"
          class="action-view-list"
          @click="setView('list')"
        ></v-btn>
        <v-btn value="mosaic" icon="mdi-view-comfy" class="pe-1 action-view-mosaic" @click="setView('mosaic')"></v-btn>
      </v-btn-toggle>

      <p-action-menu :items="menuActions" button-class="ms-1"></p-action-menu>
    </v-toolbar>

    <div v-if="album.Description" class="toolbar-details-panel">
      {{ album.Description }}
    </div>

    <p-share-dialog
      :visible="dialog.share"
      :model="album"
      @upload="webdavUpload"
      @close="dialog.share = false"
    ></p-share-dialog>
    <p-service-upload
      :visible="dialog.upload"
      :items="{ albums: album.getId() }"
      :model="album"
      @close="dialog.upload = false"
      @confirm="dialog.upload = false"
    ></p-service-upload>
    <p-album-edit-dialog :visible="dialog.edit" :album="album" @close="dialog.edit = false"></p-album-edit-dialog>
    <p-confirm-dialog
      :visible="dialog.delete"
      :text="$gettext('Are you sure you want to delete this album?')"
      :action="$gettext('Delete')"
      icon="mdi-delete-outline"
      @close="dialog.delete = false"
      @confirm="onDeleteConfirm"
    ></p-confirm-dialog>
  </v-form>
</template>
<script>
import $notify from "common/notify";
import download from "common/download";
import { T } from "common/gettext";
import $api from "common/api";

import PActionMenu from "component/action/menu.vue";
import PConfirmDialog from "component/confirm/dialog.vue";

export default {
  name: "PAlbumToolbar",
  components: {
    PActionMenu,
    PConfirmDialog,
  },
  props: {
    album: {
      type: Object,
      default: () => {},
    },
    filter: {
      type: Object,
      default: () => {},
    },
    updateFilter: {
      type: Function,
      default: () => {},
    },
    updateQuery: {
      type: Function,
      default: () => {},
    },
    settings: {
      type: Object,
      default: () => {},
    },
    refresh: {
      type: Function,
      default: () => {},
    },
  },
  data() {
    const settings = this.$config.getSettings();
    const features = settings.features;
    return {
      expanded: false,
      canUpload: this.$config.allow("files", "upload") && features.upload,
      canDownload:
        this.$config.allow("albums", "download") && features.download && !settings?.albums?.download?.disabled,
      canShare: this.$config.allow("albums", "share") && features.share,
      canManage: this.$config.allow("albums", "manage"),
      canDelete: this.$config.allow("albums", "delete"),
      experimental: this.$config.get("experimental"),
      isFullScreen: !!document.fullscreenElement,
      categories: this.$config.albumCategories(),
      collectionTitle: this.$route.meta?.collectionTitle ? this.$route.meta.collectionTitle : this.$gettext("Albums"),
      collectionRoute: this.$route.meta?.collectionRoute ? this.$route.meta.collectionRoute : "albums",
      navIcon: this.$isRtl ? "mdi-chevron-left" : "mdi-chevron-right",
      listView: this.$config.getSettings()?.search?.listView,
      dialog: {
        share: false,
        upload: false,
        edit: false,
        delete: false,
      },
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
    };
  },
  methods: {
    showExpansionPanel() {
      if (!this.expanded) {
        this.expanded = true;
      }
    },
    hideExpansionPanel() {
      if (this.expanded) {
        this.expanded = false;
      }
    },
    menuActions() {
      return [
        {
          name: "refresh",
          icon: "mdi-refresh",
          text: this.$gettext("Refresh"),
          shortcut: "Ctrl-R",
          visible: true,
          click: () => {
            this.refresh();
          },
        },
        {
          name: "edit",
          icon: "mdi-pencil",
          text: this.$gettext("Edit"),
          visible: this.canManage,
          click: () => {
            this.dialog.edit = true;
          },
        },
        {
          name: "share",
          icon: "mdi-share-variant",
          text: this.$gettext("Share"),
          class: "action-share",
          visible: this.canShare,
          click: () => {
            this.dialog.share = true;
          },
        },
        {
          name: "upload",
          icon: "mdi-cloud-upload",
          text: this.$gettext("Upload"),
          shortcut: "Ctrl-U",
          visible: this.canUpload,
          click: () => {
            this.showUpload();
          },
        },
        {
          name: "download",
          icon: "mdi-download",
          text: this.$gettext("Download"),
          visible: this.canDownload,
          click: () => {
            this.download();
          },
        },
        {
          name: "delete",
          icon: "mdi-delete-outline",
          text: this.$gettext("Delete Album"),
          visible: this.canDelete && ["album", "moment", "state"].includes(this.album.Type),
          class: "text-error",
          color: "error",
          click: () => {
            this.dialog.delete = true;
          },
        },
      ];
    },
    T() {
      return T.apply(this, arguments);
    },
    webdavUpload() {
      this.dialog.share = false;
      this.dialog.upload = true;
    },
    showUpload() {
      // Pre-select manually managed albums in upload dialog.
      if (this.album.Type === "album") {
        this.$event.publish("dialog.upload", { albums: [this.album] });
      } else {
        this.$event.publish("dialog.upload", { albums: [] });
      }
    },
    onUpdate(v) {
      this.updateQuery(v);
    },
    setView(name) {
      if (name) {
        if (name === "list" && !this.listView) {
          name = "mosaic";
        }

        this.refresh({ view: name });
      }
    },
    download() {
      this.onDownload(`${this.$config.apiUri}/albums/${this.album.UID}/dl?t=${this.$config.downloadToken}`);
    },
    onDownload(path) {
      $notify.success(this.$gettext("Downloading…"));

      download(path, "album.zip");
    },
    onDeleteConfirm() {
      $api
        .delete(`albums/${this.album.UID}`)
        .then(() => {
          $notify.success(this.$gettext("Album deleted"));
          this.$router.push({ name: this.collectionRoute });
        })
        .catch(() => {
          $notify.error(this.$gettext("Failed to delete album"));
        });
      this.dialog.delete = false;
    },
  },
};
</script>
