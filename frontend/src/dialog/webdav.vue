<template>
  <v-dialog
      v-model="visible"
      lazy
      max-width="500"
  >
    <v-card class="pa-2">
      <v-card-title class="headline pa-2">
        <translate>Connect via WebDAV</translate>
      </v-card-title>

      <v-card-text class="pa-2 body-1">
        <translate>WebDAV clients can connect to PhotoPrism using the following URL:</translate>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <v-text-field
            autocorrect="off"
            autocapitalize="none"
            browser-autocomplete="off"
            hide-details readonly
            single-line
            outline
            color="secondary-dark"
            :value="webdavUrl()"
            class="input-url"
            @click.stop="selectText($event)">
        </v-text-field>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <translate>This mounts the originals folder as a network drive and allows you to open, edit, and delete files from your computer or smartphone as if they were local.</translate>
      </v-card-text>

      <v-card-text class="pa-2 body-1">
        <v-alert
            :value="true"
            color="primary darken-2"
            icon="info"
            class="pa-2"
            type="info"
            outline
        >
          <a style="color: inherit;" href="https://docs.photoprism.app/user-guide/sync/webdav/" target="_blank">
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
  name: 'PDialogWebdav',
  props: {
    show: Boolean,
  },
  data() {
    return {
      visible: false,
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
        this.$notify.error(this.$gettext("Failed copying to clipboard"))
      }
    },
    webdavUrl() {
      return `${window.location.protocol}//admin@${window.location.host}/originals/`;
    },
    close() {
      this.$emit('close');
    },
  },
};
</script>
