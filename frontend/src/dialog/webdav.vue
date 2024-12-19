<template>
  <v-dialog v-model="visible" max-width="580">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-swap-horizontal</v-icon>
        <h6 class="text-h6"><translate>Connect via WebDAV</translate></h6>
      </v-card-title>

      <v-card-text class="text-body-2">
        <translate>WebDAV clients can connect to PhotoPrism using the following URL:</translate>
      </v-card-text>

      <v-card-text class="text-body-2">
        <v-text-field autocorrect="off" autocapitalize="none" autocomplete="off" hide-details readonly single-line :model-value="webdavUrl()" class="input-url" @click.stop="selectText($event)"> </v-text-field>
      </v-card-text>

      <v-card-text class="text-body-2 clickable" @click="windowsHelp($event)">
        <translate>On Windows, enter the following resource in the connection dialog:</translate>
      </v-card-text>

      <v-card-text class="text-body-2">
        <v-text-field autocorrect="off" autocapitalize="none" autocomplete="off" hide-details readonly single-line :model-value="windowsUrl()" class="input-url" @click.stop="selectText($event)"> </v-text-field>
      </v-card-text>

      <v-card-text class="text-body-2">
        <translate>This mounts the originals folder as a network drive and allows you to open, edit, and delete files from your computer or smartphone as if they were local.</translate>
      </v-card-text>

      <v-card-text class="pt-3 text-body-2">
        <v-alert color="surface-variant" icon="mdi-information" class="pa-2" variant="outlined">
          <a class="text-link" style="color: inherit" href="https://docs.photoprism.app/user-guide/sync/webdav/" target="_blank">
            <translate>Detailed instructions can be found in our User Guide.</translate>
          </a>
        </v-alert>
      </v-card-text>

      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-close" @click.stop="close">
          <translate>Close</translate>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import Util from "common/util";

export default {
  name: "PWebdavDialog",
  props: {
    show: Boolean,
  },
  data() {
    return {
      visible: false,
      user: this.$session.getUser(),
    };
  },
  watch: {
    show(val) {
      this.visible = val;
    },
    visible(val) {
      if (!val) {
        this.close();
      }
    },
  },
  methods: {
    selectText(ev) {
      if (!ev || !ev.target) {
        return;
      }

      ev.target.select();

      this.copyUrl();
    },
    async copyUrl() {
      try {
        await Util.copyToMachineClipboard(this.webdavUrl());
        this.$notify.success(this.$gettext("Copied to clipboard"));
      } catch (error) {
        this.$notify.error(this.$gettext("Failed copying to clipboard"));
      }
    },
    webdavUrl() {
      let baseUrl = `${window.location.protocol}//${encodeURIComponent(this.user.Name)}@${window.location.host}/originals/`;

      if (this.user.BasePath) {
        baseUrl = `${baseUrl}${this.user.BasePath}/`;
      }

      return baseUrl;
    },
    windowsUrl() {
      let baseUrl = "";

      if (window.location.protocol === "https") {
        baseUrl = `\\\\${window.location.host}@SSL\\originals\\`;
      } else {
        baseUrl = `\\\\${window.location.host}\\originals\\`;
      }

      if (this.user.BasePath) {
        const basePath = this.user.BasePath.replace(/\//g, "\\");
        baseUrl = `${baseUrl}${basePath}\\`;
      }

      return baseUrl;
    },
    windowsHelp(ev) {
      window.open("https://docs.photoprism.app/user-guide/sync/webdav/#connect-to-a-webdav-server", "_blank");
      ev.preventDefault();
      ev.stopPropagation();
    },
    close() {
      this.$emit("close");
    },
  },
};
</script>
