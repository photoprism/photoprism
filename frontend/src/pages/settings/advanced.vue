<template>
  <div class="p-tab p-settings-general">
    <v-form ref="form" lazy-validation
            dense class="p-form-settings" accept-charset="UTF-8"
            @submit.prevent="onChange">
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate>Global Options</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.Debug"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Debug Mode')"
                  :hint="$gettext('Shows more detailed log messages.')"
                  prepend-icon="pest_control"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.ReadOnly"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Read-Only Library')"
                  :hint="$gettext('Don\'t modify originals directory. Disables import and upload.')"
                  prepend-icon="camera_roll"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.Experimental"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Experimental Features')"
                  :hint="$gettext('Enable new features currently under development.')"
                  prepend-icon="science"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableWebDAV"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Disable WebDAV')"
                  :hint="$gettext('Disable built-in WebDAV server. Requires a restart.')"
                  prepend-icon="sync_disabled"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableBackups"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Disable Backups')"
                  :hint="$gettext('Don\'t backup photo and album metadata to YAML files.')"
                  prepend-icon="update_disabled"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisablePlaces"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Disable Places')"
                  :hint="$gettext('Disables reverse geocoding and maps.')"
                  prepend-icon="location_off"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableExifTool"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Disable ExifTool')"
                  :hint="$gettext('Don\'t create ExifTool JSON files for improved metadata extraction.')"
                  prepend-icon="no_photography"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableTensorFlow"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Disable TensorFlow')"
                  :hint="$gettext('Don\'t use TensorFlow for image classification.')"
                  prepend-icon="layers_clear"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import ConfigOptions from "model/config-options";
import * as options from "options/options";

export default {
  name: 'PSettingsServer',
  data() {
    const isDemo = this.$config.get("demo");

    return {
      demo: isDemo,
      readonly: this.$config.get("readonly"),
      experimental: this.$config.get("experimental"),
      config: this.$config.values,
      settings: new ConfigOptions(),
      options: options,
      busy: false,
    };
  },
  created() {
    this.load();
  },
  methods: {
    load() {
      this.settings.load();
    },
    onChange() {
      this.busy = true;

      this.settings.save().then(() => {
        this.$notify.success(this.$gettext("Settings saved"));
      }).finally(() => this.busy = false);
    },
  },
};
</script>
