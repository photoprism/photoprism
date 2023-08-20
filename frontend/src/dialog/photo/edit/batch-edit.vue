<template>
  <v-dialog :value="show" lazy persistent max-width="350" class="p-photo-batch-edit-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <v-flex xs3 text-xs-center>
            <v-icon size="54" color="secondary-dark lighten-1">edit</v-icon>
          </v-flex>
          <v-flex xs9 text-xs-left align-self-center>
            <div class="subheading pr-1">
              <translate>Are you sure you want to make changes to all {{ photoCount }} photos in the selection?</translate>
            </div>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-3">
            <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate>No</translate>
            </v-btn>
            <v-btn color="primary-button" depressed dark class="action-confirm"
                   @click.stop="confirm">
              <translate>Yes</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
export default {
  name: 'PPhotoBatchEditDialog',
  props: {
    show: Boolean,
  },
  data() {
    return {
      photoCount: this.$clipboard.selection?.length,
    };
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
