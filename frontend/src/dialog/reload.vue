<template>
  <v-dialog :model-value="show" max-width="330">
    <v-card>
      <v-card-title class="d-flex justify-start align-center flex-nowrap ga-3">
        <h6 class="text-subtitle-1"><translate>PhotoPrism has been updated…</translate></h6>
      </v-card-title>
      <v-card-actions>
        <v-btn color="button" variant="flat" @click="close">
          <translate>Cancel</translate>
        </v-btn>
        <v-btn color="highlight" class="action-update-reload" variant="flat" @click="reload">
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
