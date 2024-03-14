<template>
  <v-dialog
    :value="show"
    lazy
    persistent
    max-width="350"
    class="p-label-delete-dialog"
    @keydown.esc="cancel"
  >
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <v-flex xs3 text-xs-center>
            <v-icon size="54" color="secondary-dark lighten-1">delete_outline</v-icon>
          </v-flex>
          <v-flex xs9 text-xs-left align-self-center>
            <div class="subheading pr-1">
              <translate key="Are you sure you want to delete these labels?"
                >Are you sure you want to delete these labels?</translate
              >
            </div>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-3">
            <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn
              color="primary-button"
              depressed
              dark
              class="action-confirm"
              @click.stop="confirm"
            >
              <translate key="Delete">Delete</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: "PLabelDeleteDialog",
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
