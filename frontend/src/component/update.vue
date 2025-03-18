<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="400"
    class="p-dialog p-update"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center flex-nowrap ga-3">
        <v-icon icon="mdi-alert-decagram-outline" size="28" color="primary"></v-icon>
        <h6 class="text-h6">{{ $gettext(`Software Update`) }}</h6>
      </v-card-title>
      <v-card-text class="d-flex justify-start flex-column ga-3">
        <div class="text-body-2 data-message">{{ getMessage() }}</div>
        <div dir="ltr" class="text-caption data-version">
          {{ getVersion() }}
        </div>
      </v-card-text>
      <v-card-actions>
        <v-btn color="button" variant="flat" @click="close">
          {{ $gettext(`Close`) }}
        </v-btn>
        <v-btn color="highlight" class="action-update-reload" variant="flat" @click="reload">
          {{ $gettext(`Reload`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "PUpdate",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    getMessage() {
      return this.$gettext("A new version of %{s} is available:", { s: this.$config.getAbout() });
    },
    getVersion() {
      return this.$config.getServerVersion();
    },
    close() {
      this.$emit("close");
    },
    reload() {
      this.$notify.info(this.$gettext("Reloading…"));
      this.$notify.blockUI();
      setTimeout(() => {
        window.location.reload();
      }, 100);
      this.close();
    },
  },
};
</script>
