<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="350"
    class="p-dialog p-photo-archive-dialog"
    @keydown.esc="close"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">mdi-package-down</v-icon>
        <p class="text-subtitle-1">{{ $gettext(`Are you sure you want to archive the selection?`) }}</p>
      </v-card-title>
      <v-card-actions class="action-buttons">
        <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
          {{ $gettext(`No`) }}
        </v-btn>
        <v-btn color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          {{ $gettext(`Yes`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PPhotoArchiveDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {};
  },
  methods: {
    close() {
      this.$emit("close");
    },
    confirm() {
      this.$emit("confirm");
    },
  },
};
</script>
