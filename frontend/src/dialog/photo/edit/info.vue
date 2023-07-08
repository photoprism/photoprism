<template>
  <div class="p-tab p-tab-photo-advanced">
    <div class="v-table__overflow">
      <table class="v-datatable v-table theme--light">
        <tbody>
        <tr>
          <td>UID</td>
          <td>{{ model.UID | uppercase }}</td>
        </tr>
        <tr v-if="model.DocumentID">
          <td>Document ID</td>
          <td>{{ model.DocumentID | uppercase }}</td>
        </tr>
        <tr>
          <td :title="model.TypeSrc">
            <translate>Type</translate>
            <v-icon v-if="model.TypeSrc === 'manual'" class="src">check</v-icon>
          </td>
          <td>
            <v-select
                v-model="model.Type"
                flat solo
                browser-autocomplete="off"
                hide-details
                color="secondary-dark"
                :items="options.PhotoTypes()"
                class="input-type"
                @change="save">
            </v-select>
          </td>
        </tr>
        <tr v-if="model.Path">
          <td>
            <translate>Folder</translate>
          </td>
          <td>{{ model.Path }}</td>
        </tr>
        <tr>
          <td>
            <translate>Name</translate>
          </td>
          <td>{{ model.Name }}</td>
        </tr>
        <tr v-if="model.OriginalName">
          <td>
            <translate>Original Name</translate>
          </td>
          <td>
            <v-text-field
                v-model="model.OriginalName"
                flat solo dense hide-details
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                color="secondary-dark"
                @change="save"
            ></v-text-field>
          </td>
        </tr>
        <tr>
          <td :title="model.TitleSrc">
            <translate>Title</translate>
            <v-icon v-if="model.TitleSrc === 'manual'" class="src">check</v-icon>
          </td>
          <td>{{ model.Title }}</td>
        </tr>
        <tr>
          <td :title="model.TakenSrc">
            <translate>Taken</translate>
            <v-icon v-if="model.TakenSrc === 'manual'" class="src">check</v-icon>
          </td>
          <td>{{ model.getDateString() }}</td>
        </tr>
        <tr v-if="albums.length > 0">
          <td>
            <translate>Albums</translate>
          </td>
          <td>
            <a v-for="(a, i) in albums" :key="i" :href="a.url" class="primary--text text-link" target="_blank"><span v-if="i > 0">, </span>{{ a.title }}</a>
          </td>
        </tr>
        <tr>
          <td>
            <translate>Quality Score</translate>
          </td>
          <td>
            <v-rating v-model="model.Quality" :length="7" readonly small></v-rating>
          </td>
        </tr>
        <tr>
          <td>
            <translate>Resolution</translate>
          </td>
          <td>{{ model.Resolution }} MP</td>
        </tr>
        <tr v-if="model.Faces > 0">
          <td>
            <translate>Faces</translate>
          </td>
          <td>{{ model.Faces }}</td>
        </tr>
        <tr v-if="model.CameraSerial">
          <td>
            <translate>Camera Serial</translate>
          </td>
          <td>{{ model.CameraSerial }}
          </td>
        </tr>
        <tr v-if="model.Stack < 1">
          <td>
            <translate>Stackable</translate>
          </td>
          <td>
            <v-switch
                v-model="model.Stack"
                hide-details
                class="input-stackable"
                :true-value="0"
                :false-value="-1"
                :label="model.Stack > - 1 ? $gettext('Yes') : $gettext('No')"
                @change="save"
            ></v-switch>
          </td>
        </tr>
        <tr>
          <td>
            <translate>Favorite</translate>
          </td>
          <td>
            <v-switch
                v-model="model.Favorite"
                hide-details
                class="input-favorite"
                :label="model.Favorite ? $gettext('Yes') : $gettext('No')"
                @change="save"
            ></v-switch>
          </td>
        </tr>
        <tr v-if="$config.feature('private')">
          <td>
            <translate>Private</translate>
          </td>
          <td>
            <v-switch
                v-model="model.Private"
                hide-details
                class="input-private"
                :label="model.Private ? $gettext('Yes') : $gettext('No')"
                @change="save"
            ></v-switch>
          </td>
        </tr>
        <tr>
          <td>
            <translate>Scan</translate>
          </td>
          <td>
            <v-switch
                v-model="model.Scan"
                hide-details
                class="input-scan"
                :label="model.Scan ? $gettext('Yes') : $gettext('No')"
                @change="save"
            ></v-switch>
          </td>
        </tr>
        <tr>
          <td>
            <translate>Panorama</translate>
          </td>
          <td>
            <v-switch
                v-model="model.Panorama"
                hide-details
                class="input-panorama"
                :label="model.Panorama ? $gettext('Yes') : $gettext('No')"
                @change="save"
            ></v-switch>
          </td>
        </tr>
        <tr>
          <td :title="model.PlaceSrc">
            <translate>Place</translate>
            <v-icon v-if="model.PlaceSrc === 'manual'" class="src">check</v-icon>
          </td>
          <td>
            {{ model.locationInfo() }}
          </td>
        </tr>
        <tr v-if="model.Lat">
          <td>
            <translate>Latitude</translate>
          </td>
          <td>
            {{ model.Lat }}
          </td>
        </tr>
        <tr v-if="model.Lng">
          <td>
            <translate>Longitude</translate>
          </td>
          <td>
            {{ model.Lng }}
          </td>
        </tr>
        <tr v-if="model.Altitude">
          <td>
            <translate>Altitude</translate>
          </td>
          <td>
            {{ model.Altitude }} m
          </td>
        </tr>
        <tr v-if="model.Lat">
          <td>
            <translate>Accuracy</translate>
          </td>
          <td>
            <v-text-field
                v-model="model.CellAccuracy"
                flat solo dense hide-details
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                color="secondary-dark"
                type="number"
                suffix="m"
                style="width: 100px;"
                @change="save"
            ></v-text-field>
          </td>
        </tr>
        <tr>
          <td>
            <translate>Created</translate>
          </td>
          <td>
            {{ formatTime(model.CreatedAt) }}
          </td>
        </tr>
        <tr>
          <td>
            <translate>Updated</translate>
          </td>
          <td>
            {{ formatTime(model.UpdatedAt) }}
          </td>
        </tr>
        <tr v-if="model.EditedAt">
          <td>
            <translate>Edited</translate>
          </td>
          <td>
            {{ formatTime(model.EditedAt) }}
          </td>
        </tr>
        <tr v-if="model.CheckedAt">
          <td>
            <translate>Checked</translate>
          </td>
          <td>
            {{ formatTime(model.CheckedAt) }}
          </td>
        </tr>
        <tr v-if="model.DeletedAt">
          <td>
            <translate>Archived</translate>
          </td>
          <td>
            {{ formatTime(model.DeletedAt) }}
          </td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import Thumb from "model/thumb";
import {DateTime, Info} from "luxon";
import * as options from "options/options";

export default {
  name: 'PTabPhotoAdvanced',
  props: {
    model: {
      type: Object,
      default: () => {},
    },
    uid: String,
  },
  data() {
    return {
      options: options,
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
    };
  },
  computed: {
    monthOptions() {
      let result = [
        {"Month": -1, "Name": this.$gettext("Unknown")},
      ];

      const months = Info.months("long");

      for (let i = 0; i < months.length; i++) {
        result.push({"Month": i + 1, "UserName": months[i]});
      }

      return result;
    },
    albums() {
      if (!this.model || !this.model.Albums || this.model.Albums.length < 1) {
        return [];
      }

      const results = [];

      this.model.Albums.forEach(a => results.push({"title": a.Title, "url": this.albumUrl(a)}));

      return results;
    },
  },
  methods: {
    formatTime(s) {
      return DateTime.fromISO(s).toLocaleString(DateTime.DATETIME_MED);
    },
    save() {
      this.model.update();
    },
    close() {
      this.$emit('close');
    },
    openPhoto() {
      this.$viewer.show(Thumb.fromFiles([this.model]), 0);
    },
    albumUrl(m) {
      if (!m) {
        return '#';
      }

      return this.$router.resolve({ name: 'album', params: { uid: m.UID, slug: 'view' }}).href;
    },
  },
};
</script>
