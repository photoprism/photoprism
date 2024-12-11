<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-photo-archive-dialog" @keydown.esc="cancel">
    <v-card>
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-row>
          <v-col cols="3" class="text-center">
            <v-icon size="54" color="surface-variant lighten-1">mdi-package-down</v-icon>
          </v-col>
          <v-col cols="9" class="text-left" align-self="center">
            <div class="text-subtitle-1 pr-1">
              <translate>Are you sure you want to archive the selection?</translate>
            </div>
          </v-col>
          <v-col cols="12" class="text-right pt-6">
            <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
              <translate>No</translate>
            </v-btn>
            <v-btn color="primary-button" variant="flat" class="action-confirm" @click.stop="confirm">
              <translate>Yes</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PPhotoArchiveDialog",
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
