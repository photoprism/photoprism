<template>
  <v-dialog :value="show" lazy persistent max-width="360" class="p-photo-delete-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <v-flex xs3 text-xs-center>
            <v-icon size="54" color="secondary-dark lighten-1">delete_outline</v-icon>
          </v-flex>
          <v-flex xs9 text-xs-left align-self-center>
            <div v-if="text === ''" class="subheading pr-1">
              <translate>Are you sure you want to permanently delete these pictures?</translate>
            </div>
            <div v-else class="subheading pr-1">
              {{ text }}
            </div>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-3">
            <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate key="Cancel">Cancel</translate>
            </v-btn>
            <v-btn v-if="action === ''" color="primary-button" depressed dark class="action-confirm"
                   @click.stop="confirm">
              <translate key="Delete">Delete</translate>
            </v-btn>
            <v-btn v-else color="primary-button" depressed dark class="action-confirm"
                   @click.stop="confirm">
              {{ action }}
            </v-btn>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: 'PPhotoDeleteDialog',
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
      this.$emit('cancel');
    },
    confirm() {
      this.$emit('confirm');
    },
  }
};
</script>
