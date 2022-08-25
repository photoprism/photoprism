<template>
  <div class="p-tab p-tab-photo-files">
    <v-expansion-panel expand class="pa-0 elevation-0 secondary" :value="state">
      <template v-for="file in model.fileModels()">
        <v-expansion-panel-content v-if="!file.Missing" :key="file.UID" class="pa-0 elevation-0 secondary-light"
                                   style="margin-top: 1px;">
          <template #header>
            <div class="caption">{{ file.baseName(70) }}</div>
          </template>
          <v-card>
            <v-card-text class="white pa-0">
              <v-container fluid class="pa-0">
                <v-alert
                    :value="file.Error"
                    type="error" class="my-0 text-capitalize"
                >{{ file.Error }}</v-alert>
                <v-layout row wrap fill-height
                          align-center
                          justify-center>
                  <v-flex xs12 class="pa-0">
                    <div class="v-table__overflow">
                      <table class="v-datatable v-table theme--light photo-files">
                        <tbody>
                        <tr v-if="file.FileType === 'jpg'">
                          <td>
                            <translate>Preview</translate>
                          </td>
                          <td>
                            <v-img :src="file.thumbnailUrl('tile_224')"
                                   aspect-ratio="1"
                                   max-width="112"
                                   max-height="112"
                                   class="accent lighten-2 elevation-0 clickable my-1"
                                   @click.exact="openFile(file)"
                            >
                            </v-img>
                          </td>
                        </tr>
                        <tr>
                          <td>
                            <translate>Actions</translate>
                          </td>
                          <td>
                            <v-btn v-if="features.download" small depressed dark color="primary-button" class="btn-action action-download"
                                   @click.stop.prevent="downloadFile(file)">
                              <translate>Download</translate>
                            </v-btn>
                            <v-btn v-if="features.edit && file.FileType === 'jpg' && !file.Error && !file.Primary" small depressed dark
                                   color="primary-button"
                                   class="btn-action action-primary"
                                   @click.stop.prevent="primaryFile(file)">
                              <translate>Primary</translate>
                            </v-btn>
                            <v-btn v-if="features.edit && !file.Sidecar && !file.Error && !file.Primary && file.Root === '/'" small
                                   depressed dark color="primary-button"
                                   class="btn-action action-unstack"
                                   @click.stop.prevent="unstackFile(file)">
                              <translate>Unstack</translate>
                            </v-btn>
                            <v-btn v-if="features.delete && !file.Primary" small depressed dark color="primary-button"
                                   class="btn-action action-delete"
                                   @click.stop.prevent="showDeleteDialog(file)">
                              <translate>Delete</translate>
                            </v-btn>
                          </td>
                        </tr>
                        <tr>
                          <td title="Unique ID">
                            UID
                          </td>
                          <td>{{ file.UID | uppercase }}</td>
                        </tr>
                        <tr v-if="file.InstanceID" title="XMP">
                          <td>
                            <translate>Instance ID</translate>
                          </td>
                          <td>{{ file.InstanceID | uppercase }}</td>
                        </tr>
                        <tr>
                          <td title="SHA-1">
                            <translate>Hash</translate>
                          </td>
                          <td>{{ file.Hash }}</td>
                        </tr>
                        <tr v-if="file.Name">
                          <td>
                            <translate>Filename</translate>
                          </td>
                          <td>{{ file.Name }}</td>
                        </tr>
                        <tr v-if="file.Root">
                          <td>
                            <translate>Storage</translate>
                          </td>
                          <td>{{ file.storageInfo() }}</td>
                        </tr>
                        <tr v-if="file.OriginalName">
                          <td>
                            <translate>Original Name</translate>
                          </td>
                          <td>{{ file.OriginalName }}</td>
                        </tr>
                        <tr>
                          <td>
                            <translate>Size</translate>
                          </td>
                          <td>{{ file.sizeInfo() }}</td>
                        </tr>
                        <tr v-if="file.FileType">
                          <td>
                            <translate>Type</translate>
                          </td>
                          <td>{{ file.typeInfo() }}</td>
                        </tr>
                        <tr v-if="file.Codec && file.Codec !== file.FileType">
                          <td>
                            <translate>Codec</translate>
                          </td>
                          <td>{{ codecName(file) }}</td>
                        </tr>
                        <tr v-if="file.Duration && file.Duration > 0">
                          <td>
                            <translate>Duration</translate>
                          </td>
                          <td>{{ formatDuration(file) }}</td>
                        </tr>
                        <tr v-if="file.Frames">
                          <td>
                            <translate>Frames</translate>
                          </td>
                          <td>{{ file.Frames }}</td>
                        </tr>
                        <tr v-if="file.FPS">
                          <td>
                            <translate>FPS</translate>
                          </td>
                          <td>{{ file.FPS.toFixed(1) }}</td>
                        </tr>
                        <tr v-if="file.Primary">
                          <td>
                            <translate>Primary</translate>
                          </td>
                          <td>
                            <translate>Yes</translate>
                          </td>
                        </tr>
                        <tr v-if="file.HDR">
                          <td>
                            <translate>High Dynamic Range (HDR)</translate>
                          </td>
                          <td>
                            <translate>Yes</translate>
                          </td>
                        </tr>
                        <tr v-if="file.Portrait">
                          <td>
                            <translate>Portrait</translate>
                          </td>
                          <td>
                            <translate>Yes</translate>
                          </td>
                        </tr>
                        <tr v-if="file.Projection">
                          <td>
                            <translate>Projection</translate>
                          </td>
                          <td>{{ file.Projection | capitalize }}</td>
                        </tr>
                        <tr v-if="file.AspectRatio">
                          <td>
                            <translate>Aspect Ratio</translate>
                          </td>
                          <td>{{ file.AspectRatio }} : 1</td>
                        </tr>
                        <tr v-if="file.Orientation">
                          <td>
                            <translate>Orientation</translate>
                          </td>
                          <td>
                            <v-icon :class="`orientation-${file.Orientation}`">portrait</v-icon>
                          </td>
                        </tr>
                        <tr v-if="file.ColorProfile">
                          <td>
                            <translate>Color Profile</translate>
                          </td>
                          <td>{{ file.ColorProfile }}</td>
                        </tr>
                        <tr v-if="file.MainColor">
                          <td>
                            <translate>Main Color</translate>
                          </td>
                          <td>{{ file.MainColor | capitalize }}</td>
                        </tr>
                        <tr v-if="file.Chroma">
                          <td>
                            <translate>Chroma</translate>
                          </td>
                          <td><v-progress-linear :value="file.Chroma" style="max-width: 300px;" :title="`${file.Chroma}%`"></v-progress-linear></td>
                        </tr>
                        <tr v-if="file.Missing">
                          <td>
                            <translate>Missing</translate>
                          </td>
                          <td>
                            <translate>Yes</translate>
                          </td>
                        </tr>
                        <tr>
                          <td>
                            <translate>Added</translate>
                          </td>
                          <td>{{ formatTime(file.CreatedAt) }}
                            <translate>in</translate>
                            {{ Math.round(file.CreatedIn / 1000000) | number('0,0') }} ms
                          </td>
                        </tr>
                        <tr v-if="file.UpdatedIn">
                          <td>
                            <translate>Updated</translate>
                          </td>
                          <td>{{ formatTime(file.UpdatedAt) }}
                            <translate>in</translate>
                            {{ Math.round(file.UpdatedIn / 1000000) | number('0,0') }} ms
                          </td>
                        </tr>
                        </tbody>
                      </table>
                    </div>
                  </v-flex>
                </v-layout>
              </v-container>
            </v-card-text>
          </v-card>
        </v-expansion-panel-content>
      </template>
    </v-expansion-panel>
    <p-file-delete-dialog :show="deleteFile.dialog" @cancel="closeDeleteDialog"
                          @confirm="confirmDeleteFile"></p-file-delete-dialog>
  </div>
</template>

<script>
import Thumb from "model/thumb";
import {DateTime} from "luxon";
import Notify from "common/notify";
import Util from "common/util";

export default {
  name: 'PTabPhotoFiles',
  props: {
    model: {
      type: Object,
      default: () => {},
    },
    uid: String,
  },
  data() {
    return {
      state: [true],
      deleteFile: {
        dialog: false,
        file: null,
      },
      features: this.$config.settings().features,
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      selected: [],
      listColumns: [
        {
          text: this.$gettext('Primary'),
          value: 'Primary',
          sortable: false,
          align: 'center',
          class: 'p-col-primary'
        },
        {text: this.$gettext('Name'), value: 'Name', sortable: false, align: 'left'},
        {text: this.$gettext('Dimensions'), value: '', sortable: false, class: 'hidden-sm-and-down'},
        {text: this.$gettext('Size'), value: 'Size', sortable: false, class: 'hidden-xs-only'},
        {text: this.$gettext('Type'), value: '', sortable: false, align: 'left'},
        {text: this.$gettext('Status'), value: '', sortable: false, align: 'left'},
      ],
    };
  },
  computed: {},
  methods: {
    formatDuration(file) {
      if (!file || !file.Duration) {
        return "";
      }

      return Util.duration(file.Duration);
    },
    fileType(file) {
      if (!file || !file.FileType) {
        return "";
      }

      return Util.fileType(file.FileType);
    },
    codecName(file) {
      if (!file || !file.Codec) {
        return "";
      }

      return Util.codecName(file.Codec);
    },
    openFile(file) {
      this.$viewer.show([Thumb.fromFile(this.model, file)], 0);
    },
    downloadFile(file) {
      Notify.success(this.$gettext("Downloading…"));

      file.download();
    },
    showDeleteDialog(file) {
      this.deleteFile.dialog = true;
      this.deleteFile.file = file;
    },
    closeDeleteDialog() {
      this.deleteFile.dialog = false;
      this.deleteFile.file = null;
    },
    confirmDeleteFile() {
      if (this.deleteFile.file && this.deleteFile.file.UID) {
        this.model.deleteFile(this.deleteFile.file.UID).finally(() => this.closeDeleteDialog());
      } else {
        this.closeDeleteDialog();
      }
    },
    unstackFile(file) {
      this.model.unstackFile(file.UID);
    },
    primaryFile(file) {
      this.model.primaryFile(file.UID);
    },
    formatTime(s) {
      return DateTime.fromISO(s).toLocaleString(DateTime.DATETIME_MED);
    },
    refresh() {
    },
  },
};
</script>
