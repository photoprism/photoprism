<template>
  <div class="p-tab p-settings-advanced">
    <v-form ref="form" lazy-validation
            dense class="p-form-settings pb-1" accept-charset="UTF-8"
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
                  :label="$gettext('Debug Logs')"
                  :hint="$gettext('Shows more detailed log messages. Requires a restart.')"
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
                  :label="$gettext('Read-Only Mode')"
                  :hint="$gettext('Don\'t modify originals folder. Disables import, upload, and delete.')"
                  prepend-icon="do_not_touch"
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
                  v-model="settings.DisableBackups"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Disable Backups')"
                  :hint="$gettext('Don\'t backup photo and album metadata to YAML files.')"
                  prepend-icon="healing"
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
                  prepend-icon="filter_alt_off"
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

        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0" :title="$gettext('Thumbnail Generation')">
            <translate>Image Quality</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="px-2 pb-2">
              <v-select
                  v-model="settings.ThumbFilter"
                  :disabled="busy"
                  :items="options.ThumbFilters()"
                  :label="$gettext('Downscaling Filter')"
                  color="secondary-dark"
                  background-color="secondary-light"
                  hide-details box
                  @change="onChange"
              ></v-select>
            </v-flex>

            <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.ThumbUncached"
                  :disabled="busy"
                  class="ma-0 pa-0"
                  color="secondary-dark"
                  :label="$gettext('Dynamic Previews')"
                  :hint="$gettext('Dynamic rendering requires a powerful server. It is not recommended for NAS devices.')"
                  prepend-icon="memory"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg8 class="px-2 pb-2">
              <v-subheader class="pa-0">
                {{ $gettextInterpolate($gettext('JPEG Quality: %{n}'), {n: settings.JpegQuality}) }}
              </v-subheader>
              <v-slider
                  v-model="settings.JpegQuality"
                  :min="25"
                  :max="100"
                  :disabled="busy"
                  hide-details
                  class="mt-0"
                  @change="onChange"
              ></v-slider>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-subheader class="pa-0">
                {{ $gettextInterpolate($gettext('Dynamic Size Limit: %{n}px'), {n: settings.ThumbSizeUncached}) }}
              </v-subheader>
              <v-slider
                  v-model="settings.ThumbSizeUncached"
                  :min="720"
                  :max="7680"
                  :step="4"
                  :disabled="busy"
                  hide-details
                  class="mt-0"
                  @change="onChange"
              ></v-slider>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-subheader class="pa-0">
                {{ $gettextInterpolate($gettext('Static Size Limit: %{n}px'), {n: settings.ThumbSize}) }}
              </v-subheader>
              <v-slider
                  v-model="settings.ThumbSize"
                  :min="720"
                  :max="7680"
                  :step="4"
                  :disabled="busy"
                  hide-details
                  class="mt-0"
                  @change="onChange"
              ></v-slider>
            </v-flex>
          </v-layout>
        </v-card-actions>

        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate>File Conversion</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-subheader class="pa-0">
                {{ $gettextInterpolate($gettext('JPEG Size Limit: %{n}px'), {n: settings.JpegSize}) }}
              </v-subheader>
              <v-flex class="pr-3">
                <v-slider
                    v-model="settings.JpegSize"
                    :min="720"
                    :max="30000"
                    :step="20"
                    :disabled="busy"
                    class="mt-0"
                    @change="onChange"
                ></v-slider>
              </v-flex>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-subheader class="pa-0">
                {{ $gettextInterpolate($gettext('PNG Size Limit: %{n}px'), {n: settings.PngSize}) }}
              </v-subheader>
              <v-flex class="pr-3">
                <v-slider
                    v-model="settings.PngSize"
                    :min="720"
                    :max="30000"
                    :step="20"
                    :disabled="busy"
                    class="mt-0"
                    @change="onChange"
                ></v-slider>
              </v-flex>
            </v-flex>

            <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableDarktable"
                  :disabled="busy || settings.DisableRaw"
                  class="ma-0 pa-0 input-disable-darktable"
                  color="secondary-dark"
                  :label="$gettext('Disable Darktable')"
                  :hint="$gettext('Don\'t use Darktable to convert RAW images.')"
                  prepend-icon="raw_off"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableRawTherapee"
                  :disabled="busy || settings.DisableRaw"
                  class="ma-0 pa-0 input-disable-rawtherapee"
                  color="secondary-dark"
                  :label="$gettext('Disable RawTherapee')"
                  :hint="$gettext('Don\'t use RawTherapee to convert RAW images.')"
                  prepend-icon="raw_off"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.RawPresets"
                  :disabled="busy || settings.DisableRaw"
                  class="ma-0 pa-0 input-raw-presets"
                  color="secondary-dark"
                  :label="$gettext('Use Presets')"
                  :hint="$gettext('Enables RAW converter presets. May reduce performance.')"
                  prepend-icon="tonality"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableImageMagick"
                  :disabled="busy"
                  class="ma-0 pa-0 input-disable-imagemagick"
                  color="secondary-dark"
                  :label="$gettext('Disable ImageMagick')"
                  :hint="$gettext('Don\'t use ImageMagick to convert images.')"
                  prepend-icon="auto_fix_off"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableFFmpeg"
                  :disabled="busy || (!experimental && !settings.DisableFFmpeg)"
                  class="ma-0 pa-0 input-disable-ffmpeg"
                  color="secondary-dark"
                  :label="$gettext('Disable FFmpeg')"
                  :hint="$gettext('Disables video transcoding and thumbnail extraction.')"
                  prepend-icon="videocam_off"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex v-if="isSponsor" xs12 sm6 lg4 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.DisableVectors"
                  :disabled="busy"
                  class="ma-0 pa-0 input-disable-vectors"
                  color="secondary-dark"
                  :label="$gettext('Disable Vectors')"
                  :hint="$gettext('Disables vector graphics support.')"
                  prepend-icon="font_download_off"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>
          </v-layout>
        </v-card-actions>

        <v-card-actions class="pt-2">
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="pa-2">
              <v-btn color="primary-button" :block="$vuetify.breakpoint.xsOnly" :disabled="busy || !$config.values.restart"
                     class="white--text ml-0" depressed @click.stop.p.prevent="onRestart">
                <translate>Restart</translate>
                <v-icon :right="!rtl" :left="rtl" dark>restart_alt</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>

        <!--
        TODO: Remaining options

        OriginalsLimit: 0,
        Workers: 0,
        WakeupInterval: 0,

        SiteUrl: config.values.siteUrl,
        SitePreview: config.values.siteUrl,
        SiteTitle: config.values.siteTitle,
        SiteCaption: config.values.siteCaption,
        SiteDescription: config.values.siteDescription,
        SiteAuthor: config.values.siteAuthor,
        -->
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import ConfigOptions from "model/config-options";
import * as options from "options/options";
import {restart} from "common/server";

export default {
  name: 'PSettingsAdvanced',
  data() {
    return {
      busy: this.$config.get("demo"),
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      isSponsor: this.$config.isSponsor(),
      readonly: this.$config.get("readonly"),
      experimental: this.$config.get("experimental"),
      config: this.$config.values,
      rtl: this.$rtl,
      settings: new ConfigOptions(false),
      options: options,
    };
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$router.push({ name: "settings" });
    } else {
      this.load();
    }
  },
  methods: {
    onRestart() {
      this.busy = true;
      restart().finally(() => {
        this.busy = false;
      });
    },
    load() {
      if (this.busy || this.isDemo) {
        return;
      }

      this.busy = true;
      this.settings.load().finally(() => this.busy = false);
    },
    onChange() {
      if (this.busy || this.isDemo) {
        return;
      }

      this.busy = true;

      this.settings.save().then(() => {
        this.$notify.success(this.$gettext("Changes successfully saved"));
      }).finally(() => this.busy = false);
    },
  },
};
</script>
