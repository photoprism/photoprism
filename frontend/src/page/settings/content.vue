<template>
  <div class="p-tab p-settings-content py-2">
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="p-form-settings"
      accept-charset="UTF-8"
      @submit.prevent="onChange"
    >
      <v-card v-if="isSuperAdmin" flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Index`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" sm="4">
              <v-checkbox
                v-model="settings.features.review"
                class="ma-0 pa-0 input-review"
                density="compact"
                color="surface-variant"
                :label="$gettext('Quality Filter')"
                :hint="
                  $gettext(
                    'Require non-photographic and low-quality images to be reviewed before they appear in search results.'
                  )
                "
                prepend-icon="mdi-eye"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4">
              <v-checkbox
                v-model="settings.features.estimates"
                class="ma-0 pa-0 input-estimates"
                density="compact"
                color="surface-variant"
                :label="$gettext('Estimate Locations')"
                :hint="$gettext('Estimate the approximate location of pictures without GPS coordinates.')"
                prepend-icon="mdi-map-clock-outline"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4">
              <v-checkbox
                v-model="settings.index.convert"
                :disabled="isDemo || (!experimental && settings.index.convert)"
                class="ma-0 pa-0 input-convert"
                density="compact"
                color="surface-variant"
                :label="$gettext('Generate Previews')"
                :hint="$gettext('Extract still images and generate thumbnails while indexing.')"
                prepend-icon="mdi-image-size-select-large"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>

      <v-card v-if="isSuperAdmin" flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Stacks`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" sm="4">
              <v-checkbox
                v-model="settings.stack.meta"
                class="ma-0 pa-0 input-stack-meta"
                density="compact"
                color="surface-variant"
                :label="$gettext('Place & Time')"
                :hint="$gettext('Stack pictures taken at the exact same time and location based on their metadata.')"
                prepend-icon="mdi-clock-time-four-outline"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4">
              <v-checkbox
                v-model="settings.stack.uuid"
                class="ma-0 pa-0 input-stack-uuid"
                density="compact"
                color="surface-variant"
                :label="$gettext('Unique ID')"
                :hint="$gettext('Stack files sharing the same unique image or instance identifier.')"
                prepend-icon="mdi-fingerprint"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4">
              <v-checkbox
                v-model="settings.stack.name"
                class="ma-0 pa-0 input-stack-name"
                density="compact"
                color="surface-variant"
                :label="$gettext('Sequential Name')"
                :hint="
                  $gettext(
                    'Files with sequential names like \'IMG_1234 (2)\' and \'IMG_1234 (3)\' belong to the same picture.'
                  )
                "
                prepend-icon="mdi-format-list-numbered-rtl"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>

      <v-card flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Search`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.search.listView"
                class="ma-0 pa-0 input-search-listview"
                density="compact"
                :label="$gettext('List View')"
                :hint="$gettext('View search results as a list.')"
                prepend-icon="mdi-view-list"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.search.showTitles"
                class="ma-0 pa-0 input-search-titles"
                density="compact"
                :label="$gettext('Show Titles')"
                :hint="$gettext('Display picture titles in search results.')"
                prepend-icon="mdi-format-text"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.search.showCaptions"
                class="ma-0 pa-0 input-search-captions"
                density="compact"
                :label="$gettext('Show Captions')"
                :hint="$gettext('Display picture captions in search results.')"
                prepend-icon="mdi-text"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>
          </v-row>
        </v-card-actions>
      </v-card>

      <v-card v-if="settings.features.library && settings.features.download" flat tile class="mt-0 px-1 bg-background">
        <v-card-title class="pb-0 text-subtitle-2">
          {{ $gettext(`Download`) }}
        </v-card-title>

        <v-card-actions>
          <v-row align="start" dense>
            <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.download.originals"
                class="ma-0 pa-0 input-download-originals"
                density="compact"
                :label="$gettext('Originals')"
                :hint="$gettext('Download only original media files, without any automatically generated files.')"
                prepend-icon="mdi-camera"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.download.mediaRaw"
                class="ma-0 pa-0 input-download-raw"
                density="compact"
                :label="$gettext('RAW')"
                :hint="$gettext('Include RAW image files when downloading stacks and archives.')"
                prepend-icon="mdi-raw"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
            </v-col>

            <v-col cols="12" sm="4" class="px-2 pb-2 pt-2">
              <v-checkbox
                v-model="settings.download.mediaSidecar"
                class="ma-0 pa-0 input-download-sidecar"
                density="compact"
                :label="$gettext('Sidecar')"
                :hint="$gettext('Include sidecar files when downloading stacks and archives.')"
                prepend-icon="mdi-paperclip"
                persistent-hint
                @update:model-value="onChange"
              >
              </v-checkbox>
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
  name: "PSettingsContent",
  components: {
    PAboutFooter,
  },
  data() {
    return {
      isDemo: this.$config.isDemo(),
      isAdmin: this.$session.isAdmin(),
      isSuperAdmin: this.$session.isSuperAdmin(),
      readonly: this.$config.get("readonly"),
      experimental: this.$config.get("experimental"),
      config: this.$config.values,
      settings: new Settings(this.$config.getSettings()),
      options: options,
      busy: this.$config.loading(),
      subscriptions: [],
    };
  },
  created() {
    this.load();
    this.subscriptions.push(
      this.$event.subscribe("config.updated", (ev, data) => this.settings.setValues(data.config.settings))
    );
  },
  beforeUnmount() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.$event.unsubscribe(this.subscriptions[i]);
    }
  },
  methods: {
    load() {
      this.busy = true;
      this.$notify.blockUI();

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
    onChange() {
      if (this.busy) {
        return;
      }

      this.busy = true;
      this.$notify.blockUI();

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
