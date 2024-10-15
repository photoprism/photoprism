<template>
  <v-dialog :value="show" persistent max-width="360" class="p-photo-delete-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-row>
          <v-col cols="3" class="text-center">
            <v-icon size="54" color="secondary-dark lighten-1">mdi-delete-outline</v-icon>
          </v-col>
          <v-col cols="9" class="text-left" align-self="center">
            <div v-if="text === ''" class="subtitle-1 pr-1">
              <translate>Are you sure you want to permanently delete these pictures?</translate>
            </div>
            <div v-else class="subtitle-1 pr-1">
              {{ text }}
            </div>
          </v-col>
          <v-col cols="12" class="text-right pt-6">
            <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn v-if="action === ''" color="primary-button" depressed theme="dark" class="action-confirm" @click.stop="confirm">
              <translate key="Delete">Delete</translate>
            </v-btn>
            <v-btn v-else color="primary-button" depressed theme="dark" class="action-confirm" @click.stop="confirm">
              {{ action }}
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
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
