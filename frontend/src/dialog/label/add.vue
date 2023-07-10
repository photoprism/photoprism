<template>
  <v-dialog :value="show" lazy persistent max-width="356" class="p-label-add-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-text class="pt-3 px-3">
        <v-layout row wrap>
          <v-flex xs9 text-xs-left align-self-center>
            <v-text-field
                  v-model="label"
                  :rules="[nameRule]"
                  :label="$gettext('Label')"
                  color="secondary-dark"
                  class="input-rename background-inherit elevation-0"
                  single-line autofocus solo hide-details
            ></v-text-field>
          </v-flex>
        </v-layout>
      </v-card-text>
      <v-card-actions class="pt-0 pb-3 px-3">
        <v-layout row wrap class="pa-0">
          <v-flex xs12 text-xs-right>
            <v-btn depressed color="secondary-light" class="action-cancel mx-1" @click.stop="cancel">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn depressed color="primary-button"
                   class="action-confirm white--text compact mx-0"
                   @click.stop="confirm">
              <translate>Add Label</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

export default {
  name: 'PLabelAddDialog',
  props: {
    show: Boolean,
  },
  data() {
    return {
      label: "",
      nameRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
    };
  },
  watch: {
  },
  methods: {
    cancel() {
      this.$emit('cancel');
    },
    confirm() {
      this.$emit('confirm', this.label);
    },
  },
};
</script>
  