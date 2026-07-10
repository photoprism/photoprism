<template>
  <div class="p-tab p-settings-collections py-2">
    <v-form ref="form" validate-on="invalid-input" class="p-form-settings" accept-charset="UTF-8" @submit.prevent="onChange">
      <v-card v-if="isSuperAdmin && !hasScope && settings.features.download" flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Features`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.albums.download.disabled"
                :disabled="busy"
                class="ma-0 pa-0 input-album-download-disabled"
                density="compact"
                :label="$gettext('Disable Downloads')"
                :hint="$gettext('Prevent collections from being downloaded as zip archives.')"
                prepend-icon="mdi-download-off"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>

      <v-card v-if="isSuperAdmin && !hasScope && settings.features.download && !settings.albums.download.disabled" flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Download`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" md="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.albums.download.originals"
                :disabled="busy"
                class="ma-0 pa-0 input-album-download-originals"
                density="compact"
                :label="$gettext('Originals')"
                :hint="$gettext('Download only original media files, without any automatically generated files.')"
                prepend-icon="mdi-camera"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" md="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.albums.download.mediaRaw"
                :disabled="busy"
                class="ma-0 pa-0 input-album-download-raw"
                density="compact"
                :label="$gettext('RAW')"
                :hint="$gettext('Include RAW image files when downloading stacks and archives.')"
                prepend-icon="mdi-raw"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" md="3" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.albums.download.mediaSidecar"
                :disabled="busy"
                class="ma-0 pa-0 input-album-download-sidecar"
                density="compact"
                :label="$gettext('Sidecar')"
                :hint="$gettext('Include sidecar files when downloading stacks and archives.')"
                prepend-icon="mdi-paperclip"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" md="3" class="px-2 pb-2 pt-2">
              <v-select
                v-model="settings.albums.download.name"
                :disabled="busy"
                :items="options.DownloadName()"
                :label="$gettext('Filename')"
                prepend-icon="mdi-file-download"
                item-title="text"
                item-value="value"
                :menu-props="{ maxHeight: 346 }"
                class="input-album-download-name"
                @update:model-value="onChange"
              ></v-select>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>

      <v-card v-if="isSuperAdmin && !hasScope" flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Sort Order`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" sm="6" md="4">
              <v-select
                v-model="settings.albums.order.album"
                :disabled="busy"
                :items="options.AlbumSortOrder()"
                item-title="text"
                item-value="value"
                :label="$gettext('Albums')"
                :menu-props="{ maxHeight: 346 }"
                hide-details
                class="input-album-order"
                @update:model-value="onChange"
              ></v-select>
            </v-col>

            <v-col cols="12" sm="6" md="4">
              <v-select
                v-model="settings.albums.order.folder"
                :disabled="busy"
                :items="options.AlbumSortOrder()"
                item-title="text"
                item-value="value"
                :label="$gettext('Folders')"
                :menu-props="{ maxHeight: 346 }"
                hide-details
                class="input-folder-order"
                @update:model-value="onChange"
              ></v-select>
            </v-col>

            <v-col cols="12" sm="6" md="4">
              <v-select
                v-model="settings.albums.order.moment"
                :disabled="busy"
                :items="options.AlbumSortOrder()"
                item-title="text"
                item-value="value"
                :label="$gettext('Moments')"
                :menu-props="{ maxHeight: 346 }"
                hide-details
                class="input-moment-order"
                @update:model-value="onChange"
              ></v-select>
            </v-col>

            <v-col cols="12" sm="6" md="4">
              <v-select
                v-model="settings.albums.order.state"
                :disabled="busy"
                :items="options.AlbumSortOrder()"
                item-title="text"
                item-value="value"
                :label="$gettext('Places')"
                :menu-props="{ maxHeight: 346 }"
                hide-details
                class="input-state-order"
                @update:model-value="onChange"
              ></v-select>
            </v-col>

            <v-col cols="12" sm="6" md="4">
              <v-select
                v-model="settings.albums.order.month"
                :disabled="busy"
                :items="options.AlbumSortOrder()"
                item-title="text"
                item-value="value"
                :label="$gettext('Calendar')"
                :menu-props="{ maxHeight: 346 }"
                hide-details
                class="input-month-order"
                @update:model-value="onChange"
              ></v-select>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import Settings from "model/settings";
import * as options from "options/options";
import PAboutFooter from "component/about/footer.vue";

export default {
  name: "PSettingsCollections",
  components: {
    PAboutFooter,
  },
  data() {
    return {
      isSuperAdmin: this.$session.isSuperAdmin(),
      hasScope: this.$session.hasScope(),
      settings: new Settings(this.$config.getSettings()),
      options: options,
      busy: this.$config.loading(),
      subscriptions: [],
    };
  },
  created() {
    this.load();
    this.subscriptions.push(this.$event.subscribe("config.updated", (ev, data) => this.settings.setValues(data.config.settings)));
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    // load refreshes the client config and syncs the local settings copy.
    load() {
      this.busy = true;
      this.$notify.blockUI("busy");

      this.$config
        .load()
        .then(() => {
          this.settings.setValues(this.$config.getSettings());
        })
        .finally(() => {
          this.busy = false;
          this.$notify.unblockUI();
        });
    },
    // onChange persists the current settings and updates the client config.
    onChange() {
      if (this.busy) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI("busy");

      this.settings
        .save()
        .then(() => {
          this.$config.setSettings(this.settings);
          this.$notify.success(this.$gettext("Changes successfully saved"));
        })
        .finally(() => {
          this.busy = false;
          this.$notify.unblockUI();
        });
    },
  },
};
</script>
