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
          <td>
            <translate key="Type">Type</translate>
          </td>
          <td>{{ model.Type | capitalize }}</td>
        </tr>
        <tr v-if="model.Path">
          <td>
            <translate key="Folder">Folder</translate>
          </td>
          <td>{{ model.Path }}</td>
        </tr>
        <tr>
          <td>
            <translate key="Name">Name</translate>
          </td>
          <td>{{ model.Name }}</td>
        </tr>
        <tr v-if="model.OriginalName">
          <td>
            <translate key="Original Name">Original Name</translate>
          </td>
          <td>
            <v-text-field
                    @change="save"
                    flat solo dense hide-details v-model="model.OriginalName"
                    color="secondary-dark"
                    style="font-weight: 400; font-size: 13px;"
            ></v-text-field>
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Title">Title</translate>
          </td>
          <td>{{ model.Title }}</td>
        </tr>
        <tr v-if="model.TitleSrc">
          <td>
            <translate key="Title Source">Title Source</translate>
          </td>
          <td>{{ model.TitleSrc | capitalize }}</td>
        </tr>
        <tr>
          <td>
            <translate key="Quality Score">Quality Score</translate>
          </td>
          <td>
            <v-rating v-model="model.Quality" :length="7" readonly small></v-rating>
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Resolution">Resolution</translate>
          </td>
          <td>{{ model.Resolution }} MP</td>
        </tr>
        <tr v-if="model.CameraSerial">
          <td>
            <translate key="Camera Serial">Camera Serial</translate>
          </td>
          <td>{{ model.CameraSerial }}
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Favorite">Favorite</translate>
          </td>
          <td>
            <v-switch
                    @change="save"
                    hide-details
                    v-model="model.Favorite"
                    :label="model.Favorite ? $gettext('Yes') : $gettext('No')"
            ></v-switch>
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Private">Private</translate>
          </td>
          <td>
            <v-switch
                    @change="save"
                    hide-details
                    v-model="model.Private"
                    :label="model.Private ? $gettext('Yes') : $gettext('No')"
            ></v-switch>
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Analog">Analog</translate>
          </td>
          <td>
            <v-switch
                    @change="save"
                    hide-details
                    v-model="model.Analog"
                    :label="model.Analog ? $gettext('Yes') : $gettext('No')"
            ></v-switch>
          </td>
        </tr>
        <tr v-if="model.Lat">
          <td>
            <translate>Coordinates</translate>
          </td>
          <td>
            <translate>Latitude</translate>: {{ model.Lat }}, <translate>Longitude</translate>: {{ model.Lat }}<span v-if="model.Altitude > 0">, <translate>Altitude</translate>: {{ model.Altitude }} m</span>
          </td>
        </tr>
        <tr v-if="model.Lat">
          <td>
            <translate>Accuracy</translate>
          </td>
          <td>
            <v-text-field
                    @change="save"
                    flat solo dense hide-details v-model="model.GPSAccuracy"
                    color="secondary-dark"
                    type="number"
                    suffix="m"
                    style="font-weight: 400; font-size: 13px; width: 100px;"
            ></v-text-field>
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Created">Created</translate>
          </td>
          <td>
            {{ formatTime(model.CreatedAt) }}
          </td>
        </tr>
        <tr>
          <td>
            <translate key="Updated">Updated</translate>
          </td>
          <td>
            {{ formatTime(model.UpdatedAt) }}
          </td>
        </tr>
        <tr v-if="model.EditedAt">
          <td>
            <translate key="Edited">Edited</translate>
          </td>
          <td>
            {{ formatTime(model.EditedAt) }}
          </td>
        </tr>
        <tr v-if="model.CheckedAt">
          <td>
            <translate key="Checked">Checked</translate>
          </td>
          <td>
            {{ formatTime(model.CheckedAt) }}
          </td>
        </tr>
        <tr v-if="model.DeletedAt">
          <td>
            <translate key="Archived">Archived</translate>
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

    export default {
        name: 'p-tab-photo-advanced',
        props: {
            model: Object,
            uid: String,
        },
        data() {
            return {
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
                    result.push({"Month": i + 1, "Name": months[i]});
                }

                return result;
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
                this.$viewer.show(Thumb.fromFiles([this.model]), 0)
            },
        },
    };
</script>
