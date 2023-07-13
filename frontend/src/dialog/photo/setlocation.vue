<template>
  <v-dialog :value="show" persistent max-width="500" class="p-photo-set-location-dialog" @keydown.esc="cancel" @keydown.enter="confirm">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs10>
            <h3 class="headline mb-0">
              <translate>Set Location</translate>
            </h3>
          </v-flex>
          <v-flex xs2 text-xs-right>
            <v-icon>edit_location</v-icon>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text fluid class="pt-3 px-3">
        <v-layout fluid row wrap>
          <v-layout row wrap>
            <v-flex xs2>
              <v-icon>warning</v-icon>
            </v-flex>
            <v-flex xs10>
              <translate>Change is applied to all currently selected photos</translate>
            </v-flex>
          </v-layout>
          <v-flex fluid text-xs-left align-self-center>
            <v-container fluid fill-height>
              <div id="map" style="width: 300px; height: 300px;">
                <v-sheet fluid></v-sheet>
              </div>
            </v-container>
            <v-text-field
                  v-model="latitude"
                  :label="$gettext('Latitude')"
                  placeholder=""
                  color="secondary-dark"
                  class="input-latitude background-inherit elevation-0"
                  autofocus hide-details
            ></v-text-field>
            <v-text-field
                  v-model="longitude"
                  :label="$gettext('Longitude')"
                  placeholder=""
                  color="secondary-dark"
                  class="iniput-longitude background-inherit elevation-0"
                  hide-details
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
              <translate>Apply</translate>
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
  