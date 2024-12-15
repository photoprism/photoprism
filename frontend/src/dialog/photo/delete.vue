<template>
  <v-dialog :model-value="show" persistent max-width="360" class="p-photo-delete-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">mdi-delete-outline</v-icon>
        <p v-if="text === ''" class="text-subtitle-1"><translate>Are you sure you want to permanently delete these pictures?</translate></p>
        <p v-else class="text-subtitle-1">{{ text }}</p>
      </v-card-title>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <translate key="Cancel">Cancel</translate>
        </v-btn>
        <v-btn v-if="action === ''" color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          <translate key="Delete">Delete</translate>
        </v-btn>
        <v-btn v-else color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          {{ action }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PPhotoDeleteDialog",
  props: {
    show: Boolean,
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
    cancel() {
      this.$emit("cancel");
    },
    confirm() {
      this.$emit("confirm");
    },
  },
};
</script>
