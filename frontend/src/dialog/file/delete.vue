<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-file-delete-dialog" @keydown.esc="cancel">
    <v-card elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-row>
          <v-col cols="3" class="text-center">
            <v-icon size="54" color="secondary-dark lighten-1">mdi-delete-outline</v-icon>
          </v-col>
          <v-col cols="9" class="text-left" align-self="center">
            <div class="text-subtitle-1 pr-1">
              <translate>Are you sure you want to permanently delete this file?</translate>
            </div>
          </v-col>
          <v-col cols="12" class="text-right pt-6">
            <v-btn variant="flat" color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn color="primary-button" variant="flat" theme="dark" class="action-confirm" @click.stop="confirm">
              <translate key="Delete">Delete</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PFileDeleteDialog",
  props: {
    show: Boolean,
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
