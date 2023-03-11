<template>
  <div class="p-tab p-settings-general">
    <v-form ref="form" lazy-validation
            dense class="p-form-settings pb-1" accept-charset="UTF-8"
            @submit.prevent="onChange">
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
                  v-model="settings.ui.theme"
                  :disabled="busy"
                  :items="themes"
                  :label="$gettext('Theme')"
                  :menu-props="{'maxHeight':346}"
                  color="secondary-dark"
                  background-color="secondary-light"
                  hide-details
                  box class="input-theme"
                  @change="onChangeTheme"
              ></v-select>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                  v-model="settings.ui.language"
                  :disabled="busy"
                  :items="languages"
                  :label="$gettext('Language')"
                  :menu-props="{'maxHeight':346}"
                  color="secondary-dark"
                  background-color="secondary-light"
                  hide-details
                  box class="input-language"
                  @change="onChange"
              ></v-select>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card v-if="isDemo || isSuperAdmin" flat tile class="mt-0 px-1 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.people"
                  :disabled="busy"
                  class="ma-0 pa-0 input-people"
                  color="secondary-dark"
                  :label="$gettext('People')"
                  :hint="$gettext('Recognizes faces so that specific people can be found.')"
                  prepend-icon="person"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.moments"
                  :disabled="busy"
                  class="ma-0 pa-0 input-moments"
                  color="secondary-dark"
                  :label="$gettext('Moments')"
                  :hint="$gettext('Automatically creates albums of special moments, trips, and places.')"
                  prepend-icon="star"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.labels"
                  :disabled="busy"
                  class="ma-0 pa-0 input-labels"
                  color="secondary-dark"
                  :label="$gettext('Labels')"
                  :hint="$gettext('Browse and edit image classification labels.')"
                  prepend-icon="label"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.private"
                  :disabled="busy"
                  class="ma-0 pa-0 input-private"
                  color="secondary-dark"
                  :label="$gettext('Private')"
                  :hint="$gettext('Exclude content marked as private from search results, shared albums, labels, and places.')"
                  prepend-icon="lock"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.upload"
                  :disabled="busy || config.readonly || isDemo"
                  class="ma-0 pa-0 input-upload"
                  color="secondary-dark"
                  :label="$gettext('Upload')"
                  :hint="$gettext('Add files to your library via Web Upload.')"
                  prepend-icon="cloud_upload"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.download"
                  :disabled="busy || isDemo"
                  class="ma-0 pa-0 input-download"
                  color="secondary-dark"
                  :label="$gettext('Download')"
                  :hint="$gettext('Download single files and zip archives.')"
                  prepend-icon="get_app"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.import"
                  :disabled="busy || config.readonly || isDemo"
                  class="ma-0 pa-0 input-import"
                  color="secondary-dark"
                  :label="$gettext('Import')"
                  :hint="$gettext('Imported files will be sorted by date and given a unique name.')"
                  prepend-icon="create_new_folder"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.share"
                  :disabled="busy"
                  class="ma-0 pa-0 input-share"
                  color="secondary-dark"
                  :label="$gettext('Share')"
                  :hint="$gettext('Upload to WebDAV and share links with friends.')"
                  prepend-icon="share"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.edit"
                  :disabled="busy || isDemo"
                  class="ma-0 pa-0 input-edit"
                  color="secondary-dark"
                  :label="$gettext('Edit')"
                  :hint="$gettext('Change photo titles, locations, and other metadata.')"
                  prepend-icon="edit"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.archive"
                  :disabled="busy || isDemo"
                  class="ma-0 pa-0 input-archive"
                  color="secondary-dark"
                  :label="$gettext('Archive')"
                  :hint="$gettext('Hide photos that have been moved to archive.')"
                  prepend-icon="archive"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.delete"
                  :disabled="busy"
                  class="ma-0 pa-0 input-delete"
                  color="secondary-dark"
                  :label="$gettext('Delete')"
                  :hint="$gettext('Permanently remove files to free up storage.')"
                  prepend-icon="delete"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.services"
                  :disabled="busy"
                  class="ma-0 pa-0 input-services"
                  color="secondary-dark"
                  :label="$gettext('Services')"
                  :hint="$gettext('Share your pictures with other apps and services.')"
                  prepend-icon="sync_alt"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.library"
                  :disabled="busy || isDemo"
                  class="ma-0 pa-0 input-library"
                  color="secondary-dark"
                  :label="$gettext('Library')"
                  :hint="$gettext('Index and import files through the user interface.')"
                  prepend-icon="camera_roll"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.files"
                  :disabled="busy"
                  class="ma-0 pa-0 input-files"
                  color="secondary-dark"
                  :label="$gettext('Originals')"
                  :hint="$gettext('Browse indexed files and folders in Library.')"
                  prepend-icon="account_tree"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.logs"
                  :disabled="busy"
                  class="ma-0 pa-0 input-logs"
                  color="secondary-dark"
                  :label="$gettext('Logs')"
                  :hint="$gettext('Show server logs in Library.')"
                  prepend-icon="grading"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.account"
                  :disabled="busy || isDemo"
                  class="ma-0 pa-0 input-account"
                  color="secondary-dark"
                  :label="$gettext('Account')"
                  :hint="$gettext('Change personal profile and security settings.')"
                  prepend-icon="admin_panel_settings"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

            <v-flex v-if="!config.disable.places" xs12 sm6 lg3 class="px-2 pb-2 pt-2">
              <v-checkbox
                  v-model="settings.features.places"
                  :disabled="busy || isDemo"
                  class="ma-0 pa-0 input-places"
                  color="secondary-dark"
                  :label="$gettext('Places')"
                  :hint="$gettext('Search and display photos on a map.')"
                  prepend-icon="place"
                  persistent-hint
                  @change="onChange"
              >
              </v-checkbox>
            </v-flex>

          </v-layout>
        </v-card-actions>
      </v-card>

      <v-card v-if="settings.features.download" flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-0">
          <h3 class="body-2 mb-0">
            <translate>Download</translate>
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
                  :hint="$gettext('Download only original media files, without any automatically generated files.')"
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

      <v-card v-if="settings.features.places && !config.disable.places" flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-2">
          <h3 class="body-2 mb-0">
            <translate key="Places">Places</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                  v-model="settings.maps.style"
                  :disabled="busy"
                  :items="options.MapsStyle()"
                  :label="$gettext('Maps')"
                  :menu-props="{'maxHeight':346}"
                  color="secondary-dark"
                  background-color="secondary-light"
                  hide-details
                  box class="input-style"
                  @change="onChangeMapsStyle"
              ></v-select>
            </v-flex>

            <v-flex xs12 sm6 class="px-2 pb-2">
              <v-select
                  v-model="settings.maps.animate"
                  :disabled="busy"
                  :items="options.MapsAnimate()"
                  :label="$gettext('Animation')"
                  :menu-props="{'maxHeight':346}"
                  color="secondary-dark"
                  background-color="secondary-light"
                  hide-details
                  box class="input-animate"
                  @change="onChange"
              ></v-select>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>

    <p-sponsor-dialog :show="dialog.sponsor" @close="dialog.sponsor = false"></p-sponsor-dialog>
  </div>
</template>

<script>
import Settings from "model/settings";
import * as options from "options/options";
import * as themes from "options/themes";
import Event from "pubsub-js";

export default {
  name: 'PSettingsGeneral',
  data() {
    return {
      isDemo: this.$config.get("demo"),
      isSuperAdmin: this.$session.isSuperAdmin(),
      isPublic: this.$config.get("public"),
      config: this.$config.values,
      settings: new Settings(this.$config.settings()),
      options: options,
      busy: this.$config.loading(),
      subscriptions: [],
      themes: [],
      currentTheme: this.$config.themeName,
      mapsStyle: options.MapsStyle(),
      currentMapsStyle: this.$config.settings().maps.style,
      languages: options.Languages(),
      dialog: {
        sponsor: false,
      },
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
        this.themes = themes.Translated();
        this.settings.setValues(this.$config.settings());
        this.busy = false;
      });
    },
    onChangeTheme(value) {
      if(!value || !themes.Get(value)) {
        return false;
      }

      this.$sponsorFeatures().then(() => {
        this.currentTheme = value;
        this.onChange();
      }).catch(() => {
        if (themes.Get(value).sponsor) {
          this.dialog.sponsor = true;
          this.$nextTick(() => {
            this.settings.ui.theme = this.currentTheme;
          });
        } else {
          this.currentTheme = value;
          this.onChange();
        }
      });
    },
    onChangeMapsStyle(value) {
      if (!value) {
        return false;
      }

      const style = this.mapsStyle.find(s => s.value === value);

      if (!style) {
        return false;
      }

      this.$sponsorFeatures().then(() => {
        this.currentMapsStyle = value;
        this.onChange();
      }).catch(() => {
        if (style.sponsor) {
          this.dialog.sponsor = true;
          this.$nextTick(() => {
            this.settings.maps.style = this.currentMapsStyle;
          });
        } else {
          this.currentMapsStyle = value;
          this.onChange();
        }
      });
    },
    onChange() {
      const locale = this.settings.changed("ui", "language");

      if (locale) {
        this.busy = true;
      }

      this.settings.save().then(() => {
        this.$config.setSettings(this.settings);
        if (locale) {
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
