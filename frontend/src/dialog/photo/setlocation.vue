<template>
  <v-dialog :value="show" lazy persistent max-width="356" class="p-photo-set-location-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-text class="pt-3 px-3">
        <v-layout row wrap>
          <v-flex xs9 text-xs-left align-self-center>
            <v-text-field
                  v-model="latitude"
                  :label="$gettext('Latitude')"
                  color="secondary-dark"
                  class="input-latitude background-inherit elevation-0"
                  single-line autofocus solo hide-details
            ></v-text-field>
            <v-text-field
                  v-model="longitude"
                  :label="$gettext('Longitude')"
                  color="secondary-dark"
                  class="input-longitude background-inherit elevation-0"
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
              <translate>Set Location</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

export default {
  name: 'PPhotoSetLocationDialog',
  props: {
    show: Boolean,
  },
  data() {
    return {
      latitude: "",
      longitude: "",
    };
  },
  watch: {
  },
  methods: {
    cancel() {
      this.$emit('cancel');
    },
    confirm() {
      this.$emit('confirm', this.latitude, this.longitude);
    },
  },
};
</script>
  