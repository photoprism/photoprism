<template>
  <div class="p-tab p-settings-advanced py-2">
    <v-form ref="form" lazy-validation class="p-form-settings" accept-charset="UTF-8" @submit.prevent="onChange">
      <v-card flat tile class="mt-0 px-1 surface">
        <v-card-actions v-if="$config.values.restart">
          <v-row align="start">
            <v-col cols="12" class="pa-2 text-left">
              <v-alert color="primary" icon="mdi-information" class="pa-2" type="info" variant="outlined">
                <a style="color: inherit" href="#restart">
                  <translate>Changes to the advanced settings require a restart to take effect.</translate>
                </a>
              </v-alert>
            </v-col>
          </v-row>
        </v-card-actions>

        <v-card-title class="pb-0">
          <h3 class="text-body-2 mb-0">
            <translate>Global Options</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-row align="start">
            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.Debug"
                :disabled="busy"
                class="ma-0 pa-0 input-debug"
                color="surface-variant"
                :label="$gettext('Debug Logs')"
                :hint="$gettext('Enable debug mode to display additional logs and help with troubleshooting.')"
                prepend-icon="mdi-bug"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.Experimental"
                :disabled="busy"
                class="ma-0 pa-0 input-experimental"
                color="surface-variant"
                :label="$gettext('Experimental Features')"
                :hint="$gettext('Enable new features currently under development.')"
                prepend-icon="mdi-flask-empty"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.ReadOnly"
                :disabled="busy"
                class="ma-0 pa-0 input-readonly"
                color="surface-variant"
                :label="$gettext('Read-Only Mode')"
                :hint="$gettext('Disable features that require write permission for the originals folder.')"
                prepend-icon="mdi-hand-back-right-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableBackups"
                :disabled="busy"
                class="ma-0 pa-0 input-disable-backups"
                color="surface-variant"
                :label="$gettext('Disable Backups')"
                :hint="$gettext('Prevent database and album backups as well as YAML sidecar files from being created.')"
                prepend-icon="mdi-shield-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableWebDAV"
                :disabled="busy"
                class="ma-0 pa-0 input-disable-webdav"
                color="surface-variant"
                :label="$gettext('Disable WebDAV')"
                :hint="$gettext('Prevent other apps from accessing PhotoPrism as a shared network drive.')"
                prepend-icon="mdi-sync-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisablePlaces"
                :disabled="busy"
                class="ma-0 pa-0 input-disable-places"
                color="surface-variant"
                :label="$gettext('Disable Places')"
                :hint="$gettext('Disable interactive world maps and reverse geocoding.')"
                prepend-icon="mdi-map-marker-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableExifTool"
                :disabled="busy || (!settings.Experimental && !settings.DisableExifTool)"
                class="ma-0 pa-0 input-disable-exiftool"
                color="surface-variant"
                :label="$gettext('Disable ExifTool')"
                :hint="$gettext('ExifTool is required for full support of XMP metadata, videos and Live Photos.')"
                prepend-icon="mdi-camera-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableTensorFlow"
                :disabled="busy"
                class="ma-0 pa-0 input-disable-tensorflow"
                color="surface-variant"
                :label="$gettext('Disable TensorFlow')"
                :hint="$gettext('TensorFlow is required for image classification, facial recognition, and detecting unsafe content.')"
                prepend-icon="mdi-layers-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>

        <template v-if="!settings.DisableBackups">
          <v-card-title class="pb-0">
            <h3 class="text-body-2 mb-0">
              <translate>Backup</translate>
            </h3>
          </v-card-title>

          <v-card-actions>
            <v-row align="start">
              <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
                <v-checkbox
                  v-model="settings.BackupDatabase"
                  :disabled="busy || settings.BackupSchedule === ''"
                  class="ma-0 pa-0 input-backup-database"
                  color="surface-variant"
                  :label="$gettext('Database Backups')"
                  :hint="$gettext('Create regular backups based on the configured schedule.')"
                  prepend-icon="mdi-history"
                  persistent-hint
                  @update:model-value="onChange"
                >
                </v-checkbox>
              </v-col>

              <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
                <v-checkbox
                  v-model="settings.BackupAlbums"
                  :disabled="busy"
                  class="ma-0 pa-0 input-backup-albums"
                  color="surface-variant"
                  :label="$gettext('Album Backups')"
                  :hint="$gettext('Create YAML files to back up album metadata.')"
                  prepend-icon="mdi-image-album"
                  persistent-hint
                  @update:model-value="onChange"
                >
                </v-checkbox>
              </v-col>

              <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
                <v-checkbox
                  v-model="settings.SidecarYaml"
                  :disabled="busy"
                  class="ma-0 pa-0 input-sidecar-yaml"
                  color="surface-variant"
                  :label="$gettext('Sidecar Files')"
                  :hint="$gettext('Create YAML sidecar files to back up picture metadata.')"
                  prepend-icon="mdi-clipboard-file-outline"
                  persistent-hint
                  @update:model-value="onChange"
                >
                </v-checkbox>
              </v-col>
            </v-row>
          </v-card-actions>
        </template>

        <v-card-title class="pb-0">
          <h3 class="text-body-2 mb-0" :title="$gettext('Preview Images')">
            <translate>Preview Images</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-row align="start">
            <v-col v-if="settings.ThumbLibrary === 'imaging'" cols="12" class="px-2 pb-2">
              <v-select v-model="settings.ThumbFilter" :disabled="busy" :items="options.ThumbFilters()" :label="$gettext('Downscaling Filter')" color="surface-variant" bg-color="secondary-light" hide-details variant="solo" @update:model-value="onChange"></v-select>
            </v-col>

            <v-col cols="12" lg="4" class="px-2 pb-2">
              <v-list-subheader class="pa-0">
                {{ $gettextInterpolate($gettext("Static Size Limit: %{n}px"), { n: settings.ThumbSize }) }}
              </v-list-subheader>
              <v-slider v-model="settings.ThumbSize" :min="720" :max="7680" :step="4" :disabled="busy" hide-details class="mt-0 mx-2" @update:model-value="onChange"></v-slider>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2">
              <v-list-subheader class="pa-0">
                {{ $gettextInterpolate($gettext("Dynamic Size Limit: %{n}px"), { n: settings.ThumbSizeUncached }) }}
              </v-list-subheader>
              <v-slider v-model="settings.ThumbSizeUncached" :min="720" :max="7680" :step="4" :disabled="busy" hide-details class="mt-0 mx-2" @update:model-value="onChange"></v-slider>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.ThumbUncached"
                :disabled="busy"
                class="ma-0 pa-0"
                color="surface-variant"
                :label="$gettext('Dynamic Previews')"
                :hint="$gettext('On-demand generation of thumbnails may cause high CPU and memory usage. It is not recommended for resource-constrained servers and NAS devices.')"
                prepend-icon="mdi-memory"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>

        <v-card-title class="pb-0">
          <h3 class="text-body-2 mb-0">
            <translate>Image Quality</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-row align="start">
            <v-col cols="12" lg="4" class="px-2 pb-2">
              <v-list-subheader class="pa-0">
                {{ $gettextInterpolate($gettext("JPEG Quality: %{n}"), { n: settings.JpegQuality }) }}
              </v-list-subheader>
              <v-slider v-model="settings.JpegQuality" :min="25" :max="100" :disabled="busy" hide-details class="mt-0 mx-2" @update:model-value="onChange"></v-slider>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2">
              <v-list-subheader class="pa-0">
                {{ $gettextInterpolate($gettext("JPEG Size Limit: %{n}px"), { n: settings.JpegSize }) }}
              </v-list-subheader>
              <v-slider v-model="settings.JpegSize" :min="720" :max="30000" :step="20" :disabled="busy" class="mt-0 mx-2" @update:model-value="onChange"></v-slider>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2">
              <v-list-subheader class="pa-0">
                {{ $gettextInterpolate($gettext("PNG Size Limit: %{n}px"), { n: settings.PngSize }) }}
              </v-list-subheader>
              <v-slider v-model="settings.PngSize" :min="720" :max="30000" :step="20" :disabled="busy" class="mt-0 mx-2" @update:model-value="onChange"></v-slider>
            </v-col>
          </v-row>
        </v-card-actions>

        <v-card-title class="pb-0">
          <h3 class="text-body-2 mb-0">
            <translate>File Conversion</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-row align="start">
            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableDarktable"
                :disabled="busy || settings.DisableRaw"
                class="ma-0 pa-0 input-disable-darktable"
                color="surface-variant"
                :label="$gettext('Disable Darktable')"
                :hint="$gettext('Don\'t use Darktable to convert RAW images.')"
                prepend-icon="mdi-raw-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableRawTherapee"
                :disabled="busy || settings.DisableRaw"
                class="ma-0 pa-0 input-disable-rawtherapee"
                color="surface-variant"
                :label="$gettext('Disable RawTherapee')"
                :hint="$gettext('Don\'t use RawTherapee to convert RAW images.')"
                prepend-icon="mdi-raw-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.RawPresets"
                :disabled="busy || settings.DisableRaw"
                class="ma-0 pa-0 input-raw-presets"
                color="surface-variant"
                :label="$gettext('Use Presets')"
                :hint="$gettext('Enables RAW converter presets. May reduce performance.')"
                prepend-icon="mdi-circle-half-full"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableImageMagick"
                :disabled="busy"
                class="ma-0 pa-0 input-disable-imagemagick"
                color="surface-variant"
                :label="$gettext('Disable ImageMagick')"
                :hint="$gettext('Don\'t use ImageMagick to convert images.')"
                prepend-icon="mdi-auto-fix"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableFFmpeg"
                :disabled="busy || (!settings.Experimental && !settings.DisableFFmpeg)"
                class="ma-0 pa-0 input-disable-ffmpeg"
                color="surface-variant"
                :label="$gettext('Disable FFmpeg')"
                :hint="$gettext('Disables video transcoding and thumbnail extraction.')"
                prepend-icon="mdi-video-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col v-if="isSponsor" cols="12" sm="6" lg="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.DisableVectors"
                :disabled="busy"
                class="ma-0 pa-0 input-disable-vectors"
                color="surface-variant"
                :label="$gettext('Disable Vectors')"
                :hint="$gettext('Disables vector graphics support.')"
                prepend-icon="mdi-alpha-a-box"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>

        <v-card-actions v-if="!config.disable.restart" class="pt-6">
          <v-row align="start">
            <v-col cols="12" class="pa-2">
              <a id="restart"></a>
              <v-btn color="primary-button" :block="$vuetify.display.xs" :disabled="busy || !$config.values.restart" class="text-white" variant="flat" @click.stop.p.prevent="onRestart">
                <translate>Restart</translate>
                <v-icon :end="!rtl" :start="rtl">mdi-restart</v-icon>
              </v-btn>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import ConfigOptions from "model/config-options";
import * as options from "options/options";
import { restart } from "common/server";

export default {
  name: "PSettingsAdvanced",
  data() {
    return {
      busy: this.$config.get("demo"),
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      isSponsor: this.$config.isSponsor(),
      readonly: this.$config.get("readonly"),
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
      this.settings.load().finally(() => {
        this.busy = false;
      });
    },
    onChange() {
      if (this.busy || this.isDemo) {
        return;
      }

      this.busy = true;

      this.settings
        .save()
        .then(() => {
          this.$notify.success(this.$gettext("Changes successfully saved"));
        })
        .finally(() => (this.busy = false));
    },
  },
};
</script>
