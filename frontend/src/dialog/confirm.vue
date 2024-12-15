<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-confirm-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">{{ icon }}</v-icon>
        <p class="text-subtitle-1"><translate>Are you sure?</translate></p>
      </v-card-title>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <translate key="Cancel">Cancel</translate>
        </v-btn>
        <v-btn color="primary-button" variant="flat" class="action-confirm" @click.stop="confirm">
          <translate key="Delete">Yes</translate>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PConfirmDialog",
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    icon: {
      type: String,
      default: "mdi-delete-outline",
    },
  },
  data() {
    return {};
  },
  methods: {
    cancel() {
      this.$emit("cancel");
    },
    confirm() {
      this.$emit("confirm");
    },
  },
};
</script>
