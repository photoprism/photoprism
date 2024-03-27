<template>
  <v-dialog v-model="visible" lazy max-width="500">
    <v-card class="pa-2">
      <v-card-title class="headline pa-0">
        <v-layout row wrap class="pa-2">
          <v-flex xs9 class="text-xs-left">
            <h3 class="headline pa-0">
              <translate>Connect via WebDAV</translate>
            </h3>
          </v-flex>
          <v-flex xs3 class="text-xs-right">
            <v-icon size="28" color="primary">sync_alt</v-icon>
          </v-flex>
        </v-layout>
      </v-card-title>

      <v-card-text class="pa-2 body-1">
        <translate>WebDAV clients can connect to PhotoPrism using the following URL:</translate>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <v-text-field autocorrect="off" autocapitalize="none" browser-autocomplete="off" hide-details readonly single-line outline color="secondary-dark" :value="webdavUrl()" class="input-url" @click.stop="selectText($event)"> </v-text-field>
      </v-card-text>

      <v-card-text class="pa-2 body-1 clickable" @click="windowsHelp($event)">
        <translate>On Windows, enter the following resource in the connection dialog:</translate>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <v-text-field autocorrect="off" autocapitalize="none" browser-autocomplete="off" hide-details readonly single-line outline color="secondary-dark" :value="windowsUrl()" class="input-url" @click.stop="selectText($event)"> </v-text-field>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <translate>This mounts the originals folder as a network drive and allows you to open, edit, and delete files from your computer or smartphone as if they were local.</translate>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <v-alert :value="true" color="primary darken-2" icon="info" class="pa-2" type="info" outline>
          <a style="color: inherit" href="https://docs.photoprism.app/user-guide/sync/webdav/" target="_blank">
            <translate>Detailed instructions can be found in our User Guide.</translate>
          </a>
        </v-alert>
      </v-card-text>
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
