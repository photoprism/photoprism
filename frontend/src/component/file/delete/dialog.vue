<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="350"
    class="p-dialog p-file-delete-dialog"
    @keydown.esc="close"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">mdi-delete-outline</v-icon>
        <p class="text-subtitle-1">{{ $gettext(`Are you sure you want to permanently delete this file?`) }}</p>
      </v-card-title>
      <v-card-actions class="action-buttons mt-1">
        <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
          {{ $gettext(`Cancel`) }}
        </v-btn>
        <v-btn color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          {{ $gettext(`Delete`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PFileDeleteDialog",
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
