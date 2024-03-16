<template>
  <v-dialog :value="show" max-width="300">
    <v-card>
      <v-card-title class="subheading pa-3">
        <translate>PhotoPrism has been updated…</translate>
      </v-card-title>
      <v-card-actions class="pa-3">
        <v-spacer></v-spacer>
        <v-btn color="secondary-light" class="compact mx-2" depressed @click="close">
          <translate>Cancel</translate>
        </v-btn>
        <v-btn
          color="primary-button"
          class="action-update-reload compact"
          dark
          small
          depressed
          @click="reload"
        >
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
