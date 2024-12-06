<template>
  <v-dialog :model-value="show" max-width="300">
    <v-card>
      <v-card-title class="text-subtitle-1 pa-6">
        <translate>PhotoPrism has been updated…</translate>
      </v-card-title>
      <v-card-actions class="pa-6">
        <v-spacer></v-spacer>
        <v-btn color="secondary-light" class="compact mx-2" variant="flat" @click="close">
          <translate>Cancel</translate>
        </v-btn>
        <v-btn color="primary-button" class="action-update-reload compact" density="comfortable" variant="flat" @click="reload">
          <translate>Reload</translate>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "PReloadDialog",
  props: {
    show: Boolean,
  },
  data() {
    return {
      visible: this.show,
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
    close() {
      this.$emit("close");
    },
    reload() {
      this.$notify.info(this.$gettext("Reloading…"));
      this.$notify.blockUI();
      setTimeout(() => window.location.reload(), 100);
    },
  },
};
</script>
