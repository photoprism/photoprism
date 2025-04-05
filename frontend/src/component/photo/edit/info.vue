<template>
  <div class="p-tab p-tab-photo-advanced">
    <v-form ref="form" validate-on="invalid-input" accept-charset="UTF-8" tabindex="1" @submit.prevent>
      <div class="v-table__overflow">
        <v-table tile hover density="compact" class="bg-table">
          <tbody>
            <tr>
              <td>UID</td>
              <td class="text-break">
                <span class="cursor-copy text-uppercase" @click.stop.prevent="$util.copyText(view.model.UID)">{{
                  view.model.UID
                }}</span>
              </td>
            </tr>
            <tr v-if="view.model.DocumentID">
              <td>Document ID</td>
              <td class="text-break">
                <span class="cursor-copy text-uppercase" @click.stop.prevent="$util.copyText(view.model.DocumentID)">{{
                  view.model.DocumentID
                }}</span>
              </td>
            </tr>
            <tr>
              <td>
                <span>{{ $gettext(`Type`) }}</span>
              </td>
              <td v-tooltip="formatSource(view.model?.TypeSrc, $gettext('Default'))">
                <v-select
                  v-model="view.model.Type"
                  :append-icon="view.model.TypeSrc === 'manual' ? 'mdi-check' : ''"
                  :list-props="{ density: 'compact' }"
                  max-width="160"
                  variant="solo"
                  bg-color="transparent"
                  density="compact"
                  autocomplete="off"
                  hide-details
                  :items="options.PhotoTypes()"
                  item-title="text"
                  item-value="value"
                  class="input-type"
                  @update:model-value="save"
                ></v-select>
              </td>
            </tr>
            <tr v-if="view.model.Path">
              <td>
                {{ $gettext(`Folder`) }}
              </td>
              <td class="text-break">
                <span class="cursor-copy" @click.stop.prevent="$util.copyText(view.model.Path)">{{
                  view.model.Path
                }}</span>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Name`) }}
              </td>
              <td class="text-break">
                <span class="cursor-copy" @click.stop.prevent="$util.copyText(view.model.Name)">{{
                  view.model.Name
                }}</span>
              </td>
            </tr>
            <tr v-if="view.model.OriginalName">
              <td>
                {{ $gettext(`Original Name`) }}
              </td>
              <td>
                <v-text-field
                  v-model="view.model.OriginalName"
                  flat
                  variant="solo"
                  bg-color="transparent"
                  density="compact"
                  hide-details
                  autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  @change="save"
                ></v-text-field>
              </td>
            </tr>
            <tr>
              <td>
                <span>{{ $gettext(`Title`) }}</span>
              </td>
              <td>
                <div v-tooltip="formatSource(view.model?.TitleSrc, $gettext('Generated'))" class="text-flex text-break">
                  <span class="cursor-copy text-break" @click.stop.prevent="$util.copyText(view.model.Title)">{{
                    view.model.Title
                  }}</span>
                  <v-icon v-if="view.model.TitleSrc === 'name'" icon="mdi-file" class="src"></v-icon>
                  <v-icon v-else-if="view.model.TitleSrc === 'manual'" icon="mdi-check" class="src"></v-icon>
                </div>
              </td>
            </tr>
            <tr>
              <td>
                <span>{{ $gettext(`Taken`) }}</span>
              </td>
              <td>
                <div v-tooltip="formatSource(view.model?.TakenSrc, $gettext('File'))" class="text-flex text-break">
                  <div>{{ view.model.getDateString() }}</div>
                  <v-icon v-if="view.model.TakenSrc === ''" icon="mdi-file-clock-outline" class="src"></v-icon>
                  <!-- v-icon v-else-if="view.model.TakenSrc === 'meta'" icon="mdi-camera" class="src"></v-icon -->
                  <v-icon v-else-if="view.model.TakenSrc === 'name'" icon="mdi-file-tree-outline" class="src"></v-icon>
                  <v-icon v-else-if="view.model.TakenSrc === 'estimate'" icon="mdi-file-question" class="src"></v-icon>
                  <v-icon v-else-if="view.model.TakenSrc === 'manual'" icon="mdi-check" class="src"></v-icon>
                </div>
              </td>
            </tr>
            <tr v-if="albums.length > 0">
              <td>
                {{ $gettext(`Albums`) }}
              </td>
              <td class="text-break">
                <a v-for="(a, i) in albums" :key="i" :href="a.url" class="text-primary text-link" target="_blank"
                  ><span v-if="i > 0">, </span>{{ a.title }}</a
                >
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Quality Score`) }}
              </td>
              <td>
                <v-rating v-model="view.model.Quality" :length="7" size="small" density="compact" readonly></v-rating>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Resolution`) }}
              </td>
              <td>{{ view.model.Resolution }} MP</td>
            </tr>
            <tr v-if="view.model.Faces > 0">
              <td>
                {{ $gettext(`Faces`) }}
              </td>
              <td>{{ view.model.Faces }}</td>
            </tr>
            <tr v-if="view.model.CameraSerial">
              <td>
                {{ $gettext(`Camera Serial`) }}
              </td>
              <td class="text-break">{{ view.model.CameraSerial }}</td>
            </tr>
            <tr v-if="view.model.Stack < 1">
              <td>
                {{ $gettext(`Stackable`) }}
              </td>
              <td>
                <v-switch
                  v-model="view.model.Stack"
                  hide-details
                  class="input-stackable"
                  :true-value="0"
                  :false-value="-1"
                  :label="view.model.Stack > -1 ? $gettext('Yes') : $gettext('No')"
                  @update:model-value="save"
                ></v-switch>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Favorite`) }}
              </td>
              <td>
                <v-switch
                  v-model="view.model.Favorite"
                  hide-details
                  class="input-favorite ml-2"
                  :label="view.model.Favorite ? $gettext('Yes') : $gettext('No')"
                  @update:model-value="save"
                ></v-switch>
              </td>
            </tr>
            <tr v-if="$config.feature('private')">
              <td>
                {{ $gettext(`Private`) }}
              </td>
              <td>
                <v-switch
                  v-model="view.model.Private"
                  hide-details
                  class="input-private ml-2"
                  :label="view.model.Private ? $gettext('Yes') : $gettext('No')"
                  @update:model-value="save"
                ></v-switch>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Scan`) }}
              </td>
              <td>
                <v-switch
                  v-model="view.model.Scan"
                  hide-details
                  class="input-scan ml-2"
                  :label="view.model.Scan ? $gettext('Yes') : $gettext('No')"
                  @update:model-value="save"
                ></v-switch>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Panorama`) }}
              </td>
              <td>
                <v-switch
                  v-model="view.model.Panorama"
                  hide-details
                  class="input-panorama ml-2"
                  :label="view.model.Panorama ? $gettext('Yes') : $gettext('No')"
                  @update:model-value="save"
                ></v-switch>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Place`) }}
              </td>
              <td>
                <div v-tooltip="formatSource(view.model.PlaceSrc, $gettext('Missing'))" class="text-flex">
                  <div>{{ view.model.locationInfo() }}</div>
                  <v-icon v-if="view.model.PlaceSrc === 'estimate'" icon="mdi-map-clock-outline" class="src"></v-icon>
                  <!-- v-icon v-else-if="view.model.PlaceSrc === 'meta'" icon="mdi-camera" class="src"></v-icon -->
                  <v-icon v-else-if="view.model.PlaceSrc === 'manual'" icon="mdi-check" class="src"></v-icon>
                </div>
              </td>
            </tr>
            <tr v-if="view.model.Lat">
              <td>
                {{ $gettext(`Latitude`) }}
              </td>
              <td>
                {{ view.model.Lat }}
              </td>
            </tr>
            <tr v-if="view.model.Lng">
              <td>
                {{ $gettext(`Longitude`) }}
              </td>
              <td>
                {{ view.model.Lng }}
              </td>
            </tr>
            <tr v-if="view.model.Altitude">
              <td>
                {{ $gettext(`Altitude`) }}
              </td>
              <td>{{ view.model.Altitude }} m</td>
            </tr>
            <tr v-if="view.model.Lat">
              <td>
                {{ $gettext(`Accuracy`) }}
              </td>
              <td>
                <v-text-field
                  v-model="view.model.CellAccuracy"
                  variant="solo"
                  bg-color="transparent"
                  density="compact"
                  hide-details
                  autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  type="number"
                  suffix="m"
                  :max-width="100"
                  @change="save"
                ></v-text-field>
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Created`) }}
              </td>
              <td class="text-break">
                {{ formatTime(view.model.CreatedAt) }}
              </td>
            </tr>
            <tr>
              <td>
                {{ $gettext(`Updated`) }}
              </td>
              <td class="text-break">
                {{ formatTime(view.model.UpdatedAt) }}
              </td>
            </tr>
            <tr v-if="view.model.EditedAt">
              <td>
                {{ $gettext(`Edited`) }}
              </td>
              <td class="text-break">
                {{ formatTime(view.model.EditedAt) }}
              </td>
            </tr>
            <tr v-if="view.model.CheckedAt">
              <td>
                {{ $gettext(`Checked`) }}
              </td>
              <td class="text-break">
                {{ formatTime(view.model.CheckedAt) }}
              </td>
            </tr>
            <tr v-if="view.model.DeletedAt">
              <td>
                {{ $gettext(`Archived`) }}
              </td>
              <td class="text-break">
                {{ formatTime(view.model.DeletedAt) }}
              </td>
            </tr>
          </tbody>
        </v-table>
      </div>
    </v-form>
  </div>
</template>

<script>
import { DateTime } from "luxon";
import * as options from "options/options";
import { $gettext, T } from "common/gettext";

export default {
  name: "PTabPhotoAdvanced",
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      view: this.$view.getData(),
      options: options,
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
    };
  },
  computed: {
    albums() {
      if (!this.view?.model || !Array.isArray(this.view.model?.Albums) || this.view.model.Albums?.length < 1) {
        return [];
      }

      const results = [];

      this.view.model.Albums.forEach((a) => results.push({ title: a.Title, url: this.albumUrl(a) }));

      return results;
    },
  },
  methods: {
    $gettext,
    formatSource(s, defaultValue) {
      switch (s) {
        case null:
        case false:
        case undefined:
        case "":
        case "auto":
          return defaultValue ? defaultValue : this.$gettext("Auto");
        case "default":
          return this.$gettext("Default");
        case "manual":
          return this.$gettext("Manual");
        case "meta":
          return this.$gettext("Metadata");
        case "xmp":
          return "XMP";
        case "estimate":
          return this.$gettext("Estimate");
        case "name":
          return this.$gettext("Name");
        case "title":
          return this.$gettext("Title");
        case "caption":
          return this.$gettext("Caption");
        case "image":
          return this.$gettext("Image");
        case "location":
          return this.$gettext("Location");
        default:
          return T(this.$util.capitalize(s));
      }
    },
    formatTime(s) {
      return DateTime.fromISO(s).toLocaleString(DateTime.DATETIME_MED);
    },
    save() {
      this.view.model.update();
    },
    close() {
      this.$emit("close");
    },
    albumUrl(m) {
      if (!m) {
        return "#";
      }

      return this.$router.resolve({
        name: "album",
        params: { album: m.UID, slug: "view" },
      }).href;
    },
  },
};
</script>
