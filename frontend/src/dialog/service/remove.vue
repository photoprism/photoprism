<template>
  <v-dialog :value="show" persistent max-width="350" class="p-account-delete-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-row>
          <v-col cols="3" class="text-xs-center">
            <v-icon size="54" color="secondary-dark lighten-1">mdi-delete-outline</v-icon>
          </v-col>
          <v-col cols="9" class="text-xs-left" align-self="center">
            <div class="subtitle-1 pr-1">
              <translate key="Are you sure you want to delete this account?">Are you sure you want to delete this account?</translate>
            </div>
          </v-col>
          <v-col cols="12" class="text-xs-right pt-6">
            <v-btn depressed color="secondary-light" class="action-cancel ml-2" @click.stop="cancel">
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn depressed dark color="primary-button" class="action-confirm" @click.stop="confirm">
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
