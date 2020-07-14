<template>
  <div class="p-tab p-settings-general">
    <v-form lazy-validation dense
            ref="form" class="p-form-settings" accept-charset="UTF-8"
            @submit.prevent="onChange">
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate>Library</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-private"
                      v-model="settings.features.private"
                      color="secondary-dark"
                      :label="$gettext('Hide Private')"
                      :hint="$gettext('Exclude content marked as private from search results, shared albums, labels and places.')"
                      prepend-icon="lock"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-review"
                      v-model="settings.features.review"
                      color="secondary-dark"
                      :label="$gettext('Quality Filter')"
                      :hint="$gettext('Non-photographic and low-quality images require a review before they appear in search results.')"
                      prepend-icon="remove_red_eye"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy || readonly"
                      class="ma-0 pa-0 input-convert"
                      v-model="settings.index.convert"
                      color="secondary-dark"
                      :label="$gettext('Convert to JPEG')"
                      :hint="$gettext('File types like RAW might need to be converted so that they can be displayed in a browser. JPEGs will be stored in the same folder next to the original using the best possible quality.')"
                      prepend-icon="photo_camera"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-sequences"
                      v-model="settings.index.sequences"
                      color="secondary-dark"
                      :label="$gettext('Group Sequential')"
                      :hint="$gettext('Files with sequential names like \'IMG_1234 (2)\' or \'IMG_1234 copy 2\' belong to the same photo.')"
                      prepend-icon="photo_library"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-2">
          <h3 class="body-2 mb-0">
            <translate key="User Interface">User Interface</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                      @change="onChange"
                      :disabled="busy"
                      :items="options.Themes()"
                      :label="$gettext('Theme')"
                      color="secondary-dark"
                      background-color="secondary-light"
                      v-model="settings.theme"
                      hide-details box
                      class="input-theme"
              ></v-select>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                      @change="onChange"
                      :disabled="busy"
                      :items="options.Languages()"
                      :label="$gettext('Language')"
                      color="secondary-dark"
                      background-color="secondary-light"
                      v-model="settings.language"
                      hide-details box
                      class="input-language"
              ></v-select>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card flat tile class="mt-0 px-1 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy || readonly"
                      class="ma-0 pa-0 input-upload"
                      v-model="settings.features.upload"
                      color="secondary-dark"
                      :label="$gettext('Upload')"
                      :hint="$gettext('Add files to your library via Web Upload.')"
                      prepend-icon="cloud_upload"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-download"
                      v-model="settings.features.download"
                      color="secondary-dark"
                      :label="$gettext('Download')"
                      :hint="$gettext('Download single files and zip archives.')"
                      prepend-icon="get_app"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-share"
                      v-model="settings.features.share"
                      color="secondary-dark"
                      :label="$gettext('Share')"
                      :hint="$gettext('Upload to WebDAV and share links with friends.')"
                      prepend-icon="share"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-archive"
                      v-model="settings.features.archive"
                      color="secondary-dark"
                      :label="$gettext('Archive')"
                      :hint="$gettext('Hide photos that have been moved to archive.')"
                      prepend-icon="archive"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-edit"
                      v-model="settings.features.edit"
                      color="secondary-dark"
                      :label="$gettext('Edit')"
                      :hint="$gettext('Change photo titles, locations and other metadata.')"
                      prepend-icon="edit"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-files"
                      v-model="settings.features.files"
                      color="secondary-dark"
                      :label="$gettext('Originals')"
                      :hint="$gettext('Browse indexed files and folders in Library.')"
                      prepend-icon="insert_drive_file"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-moments"
                      v-model="settings.features.moments"
                      color="secondary-dark"
                      :label="$gettext('Moments')"
                      :hint="$gettext('Let PhotoPrism create albums from past events.')"
                      prepend-icon="star"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-labels"
                      v-model="settings.features.labels"
                      color="secondary-dark"
                      :label="$gettext('Labels')"
                      :hint="$gettext('Browse and edit image classification labels.')"
                      prepend-icon="label"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-library"
                      v-model="settings.features.library"
                      color="secondary-dark"
                      :label="$gettext('Library')"
                      :hint="$gettext('Show Library in navigation menu.')"
                      prepend-icon="camera_roll"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy || readonly"
                      class="ma-0 pa-0 input-import"
                      v-model="settings.features.import"
                      color="secondary-dark"
                      :label="$gettext('Import')"
                      :hint="$gettext('Imported files will be sorted by date and given a unique name.')"
                      prepend-icon="create_new_folder"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-logs"
                      v-model="settings.features.logs"
                      color="secondary-dark"
                      :label="$gettext('Logs')"
                      :hint="$gettext('Show server logs in Library.')"
                      prepend-icon="notes"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-places"
                      v-model="settings.features.places"
                      color="secondary-dark"
                      :label="$gettext('Places')"
                      :hint="$gettext('Search and display photos on a map.')"
                      prepend-icon="place"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card flat tile class="mt-0 px-1 application" v-if="settings.features.places">
        <v-card-title primary-title class="pb-2">
          <h3 class="body-2 mb-0">
            <translate key="Places">Places</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                      @change="onChange"
                      :disabled="busy"
                      :items="options.MapsStyle()"
                      :label="$gettext('Style')"
                      color="secondary-dark"
                      background-color="secondary-light"
                      v-model="settings.maps.style"
                      hide-details box
                      class="input-style"
              ></v-select>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                      @change="onChange"
                      :disabled="busy"
                      :items="options.MapsAnimate()"
                      :label="$gettext('Animation')"
                      color="secondary-dark"
                      background-color="secondary-light"
                      v-model="settings.maps.animate"
                      hide-details box
                      class="input-animate"
              ></v-select>
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

    export default {
        name: 'p-settings-general',
        data() {
            return {
                readonly: this.$config.get("readonly"),
                experimental: this.$config.get("experimental"),
                settings: new Settings(this.$config.settings()),
                options: options,
                busy: false,
            };
        },
        methods: {
            load() {
                this.settings.load();
            },
            onChange() {
                const reload = this.settings.changed("language");

                if (reload) {
                    this.busy = true;
                }

                this.settings.save().then((s) => {
                    if (reload) {
                        this.$notify.info(this.$gettext("Reloading…"));
                        this.$notify.blockUI();
                        setTimeout(() => window.location.reload(), 100);
                    } else {
                        this.$notify.success(this.$gettext("Settings saved"));
                    }
                }).finally(() => this.busy = false)
            },
        },
        created() {
            this.load();
        },
    };
</script>
