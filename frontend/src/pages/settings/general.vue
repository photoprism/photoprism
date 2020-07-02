<template>
  <div class="p-tab p-settings-general">
    <v-form lazy-validation dense
            ref="form" class="p-form-settings" accept-charset="UTF-8"
            @submit.prevent="onChange">
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate key="Library">Library</translate>
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
                      :label="labels.private"
                      :hint="hints.private"
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
                      :label="labels.review"
                      :hint="hints.review"
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
                      :label="labels.convert"
                      :hint="hints.convert"
                      prepend-icon="photo_camera"
                      persistent-hint
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                      @change="onChange"
                      :disabled="busy"
                      class="ma-0 pa-0 input-group"
                      v-model="settings.index.group"
                      color="secondary-dark"
                      :label="labels.group"
                      :hint="hints.group"
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
                      :items="options.Themes"
                      :label="labels.theme"
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
                      :items="options.Languages"
                      :label="labels.language"
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
                      :label="labels.upload"
                      :hint="hints.upload"
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
                      :label="labels.download"
                      :hint="hints.download"
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
                      :label="labels.share"
                      :hint="hints.share"
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
                      :label="labels.archive"
                      :hint="hints.archive"
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
                      :label="labels.edit"
                      :hint="hints.edit"
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
                      :label="labels.originals"
                      :hint="hints.originals"
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
                      :label="labels.moments"
                      :hint="hints.moments"
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
                      :label="labels.labels"
                      :hint="hints.labels"
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
                      :label="labels.library"
                      :hint="hints.library"
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
                      :label="labels.import"
                      :hint="hints.import"
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
                      :label="labels.logs"
                      :hint="hints.logs"
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
                      :label="labels.places"
                      :hint="hints.places"
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
                      :items="options.MapsStyle"
                      :label="labels.mapsStyle"
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
                      :items="options.MapsAnimate"
                      :label="labels.mapsAnimate"
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

      <v-card flat tile class="mt-0 px-1 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="px-2 pb-2 body-1">
              <router-link to="/about">
                PhotoPrism™
                {{$config.get("version")}}
                <br>© 2018-2020 Michael Mayer
              </router-link>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2 body-1 text-xs-left text-sm-right">
              <a href="https://docs.photoprism.org/credits/" class="secondary-dark--text"
                 target="_blank">Thank you</a> to everyone who made this possible!
              <br>
              <a href="https://raw.githubusercontent.com/photoprism/photoprism/develop/NOTICE"
                 class="secondary-dark--text" target="_blank">
                3rd-party software packages</a>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>
  </div>
</template>

<script>
    import Settings from "model/settings";
    import * as options from "resources/options";
    import labels from "resources/labels";
    import hints from "resources/hints";

    export default {
        name: 'p-settings-general',
        data() {
            return {
                readonly: this.$config.get("readonly"),
                experimental: this.$config.get("experimental"),
                settings: new Settings(this.$config.settings()),
                options: options,
                labels: labels,
                hints: hints,
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
                        this.$notify.info(this.$gettext("Reloading..."));
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
