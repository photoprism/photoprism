<template>
  <div class="p-tab p-settings-library">
    <v-form ref="form" lazy-validation
            dense class="p-form-settings pb-1" accept-charset="UTF-8"
            @submit.prevent="onChange">
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate>Index</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.estimates"
                  :disabled="busy"
                  class="ma-0 pa-0 input-estimates"
                  color="secondary-dark"
                  :label="$gettext('Estimates')"
                  :hint="$gettext('Estimate the approximate location of pictures without coordinates.')"
                  prepend-icon="timeline"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.review"
                  :disabled="busy"
                  class="ma-0 pa-0 input-review"
                  color="secondary-dark"
                  :label="$gettext('Quality Filter')"
                  :hint="$gettext('Non-photographic and low-quality images require a review before they appear in search results.')"
                  prepend-icon="remove_red_eye"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.index.convert"
                  :disabled="busy || demo"
                  class="ma-0 pa-0 input-convert"
                  color="secondary-dark"
                  :label="$gettext('Convert to JPEG')"
                  :hint="$gettext('Automatically create JPEGs for other file types so that they can be displayed in a browser.')"
                  prepend-icon="photo_camera"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0"
                      :title="$gettext('Stacks group files with a similar frame of reference, but differences of quality, format, size or color.')">
          <h3 class="body-2 mb-0">
            <translate>Stacks</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.stack.meta"
                  :disabled="busy"
                  class="ma-0 pa-0 input-stack-meta"
                  color="secondary-dark"
                  :label="$gettext('Place & Time')"
                  :hint="$gettext('Stack pictures taken at the exact same time and location based on their metadata.')"
                  prepend-icon="schedule"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.stack.uuid"
                  :disabled="busy"
                  class="ma-0 pa-0 input-stack-uuid"
                  color="secondary-dark"
                  :label="$gettext('Unique ID')"
                  :hint="$gettext('Stack files sharing the same unique image or instance identifier.')"
                  prepend-icon="fingerprint"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.stack.name"
                  :disabled="busy"
                  class="ma-0 pa-0 input-stack-name"
                  color="secondary-dark"
                  :label="$gettext('Sequential Name')"
                  :hint="$gettext('Files with sequential names like \'IMG_1234 (2)\' and \'IMG_1234 (3)\' belong to the same picture.')"
                  prepend-icon="format_list_numbered_rtl"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate>Downloads</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.download.originals"
                  :disabled="busy"
                  class="ma-0 pa-0 input-download-originals"
                  color="secondary-dark"
                  :label="$gettext('Originals')"
                  :hint="$gettext('Download only original media files, without any automatically generated sidecar files.')"
                  prepend-icon="camera"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.download.mediaRaw"
                  :disabled="busy"
                  class="ma-0 pa-0 input-download-raw"
                  color="secondary-dark"
                  :label="$gettext('RAW')"
                  :hint="$gettext('Include RAW image files when downloading stacks and archives.')"
                  prepend-icon="raw_on"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.download.mediaSidecar"
                  :disabled="busy"
                  class="ma-0 pa-0 input-download-sidecar"
                  color="secondary-dark"
                  :label="$gettext('Sidecar')"
                  :hint="$gettext('Include sidecar files when downloading stacks and archives.')"
                  prepend-icon="attach_file"
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
import Settings from "model/settings";
import * as options from "options/options";
import Event from "pubsub-js";

export default {
  name: 'PSettingsLibrary',
  data() {
    const isDemo = this.$config.get("demo");

    return {
      demo: isDemo,
      readonly: this.$config.get("readonly"),
      experimental: this.$config.get("experimental"),
      config: this.$config.values,
      settings: new Settings(this.$config.settings()),
      options: options,
      busy: this.$config.loading(),
      subscriptions: [],
    };
  },
  created() {
    this.load();
    this.subscriptions.push(Event.subscribe("config.updated", (ev, data) => this.settings.setValues(data.config.settings)));
  },
  destroyed() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      Event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    load() {
      this.$config.load().then(() => {
        this.settings.setValues(this.$config.settings());
        this.busy = false;
      });
    },
    onChange() {
      const reload = this.settings.changed("ui", "language");

      if (reload) {
        this.busy = true;
      }

      this.settings.save().then(() => {
        if (reload) {
          this.$notify.info(this.$gettext("Reloading…"));
          this.$notify.blockUI();
          setTimeout(() => window.location.reload(), 100);
        } else {
          this.$notify.success(this.$gettext("Changes successfully saved"));
        }
      }).finally(() => this.busy = false);
    },
  },
};
</script>
