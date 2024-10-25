<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-confirm-dialog" @keydown.esc="cancel">
    <v-card elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-row>
          <v-col cols="3" class="text-center">
            <!-- TODO: change this icon -->
            <v-icon size="54" color="secondary-dark lighten-1">{{ icon }}</v-icon>
          </v-col>
          <v-col cols="9" class="text-left" align-self="center">
            <div class="text-subtitle-1 pr-1">
              <translate>Are you sure?</translate>
            </div>
          </v-col>
          <v-col cols="12" class="pt-6 text-right">
            <v-btn variant="flat" color="secondary-light" class="action-cancel compact" @click.stop="cancel">
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn color="primary-button" variant="flat" theme="dark" class="action-confirm compact" @click.stop="confirm">
              <translate key="Delete">Yes</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
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
      default: "delete_outline",
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
