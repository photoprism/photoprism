<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="360"
    class="p-dialog p-confirm-dialog"
    @keydown.esc.exact="close"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon :icon="icon" :size="iconSize" color="primary"></v-icon>
        <div class="text-subtitle-1">{{ text ? text : $gettext(`Are you sure?`) }}</div>
      </v-card-title>
      <v-card-actions class="action-buttons">
        <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
          {{ $gettext(`Cancel`) }}
        </v-btn>
        <v-btn color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          {{ action ? action : $gettext(`Yes`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PConfirmDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    icon: {
      type: String,
      default: "mdi-delete-outline",
    },
    iconSize: {
      type: Number,
      default: 54,
    },
    text: {
      type: String,
      default: "",
    },
    action: {
      type: String,
      default: "",
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
