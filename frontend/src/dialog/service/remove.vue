<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-account-delete-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">mdi-delete-outline</v-icon>
        <p class="text-subtitle-1"><translate key="Are you sure you want to delete this account?">Are you sure you want to delete this account?</translate></p>
      </v-card-title>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <translate key="Cancel">Cancel</translate>
        </v-btn>
        <v-btn variant="flat" color="highlight" class="action-confirm" @click.stop="confirm">
          <translate key="Delete">Delete</translate>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PAccountDeleteDialog",
  props: {
    show: Boolean,
    model: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
      loading: false,
    };
  },
  methods: {
    cancel() {
      this.$emit("cancel");
    },
    confirm() {
      this.loading = true;

      this.model.remove().then(() => {
        this.loading = false;
        this.$notify.success(this.$gettext("Account deleted"));
        this.$emit("confirm");
      });
    },
  },
};
</script>
