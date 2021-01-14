<template>
  <div class="p-tab p-tab-photo-files">
    <v-expansion-panel expand class="pa-0 elevation-0 secondary" :value="state">
      <v-expansion-panel-content v-for="(file, index) in model.fileModels()" :key="index"
                                 class="pa-0 elevation-0 secondary-light" style="margin-top: 1px;">
        <template #header>
          <div class="caption">{{ file.baseName(70) }}</div>
        </template>
        <v-card>
          <v-card-text class="white pa-0">
            <v-container fluid class="pa-0">
              <v-layout row wrap fill-height
                        align-center
                        justify-center>
                <v-flex xs12 class="pa-0">
                  <div class="v-table__overflow">
                    <table class="v-datatable v-table theme--light photo-files">
                      <tbody>
                      <tr v-if="file.Type === 'jpg'">
                        <td>
                          <translate>Preview</translate>
                        </td>
                        <td>
                          <v-img :src="file.thumbnailUrl('tile_224')"
                                 aspect-ratio="1"
                                 max-width="112"
                                 max-height="112"
                                 class="accent lighten-2 elevation-0 clickable"
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
                          <v-btn small depressed dark color="secondary-dark" class="ma-0 action-download"
                                 @click.stop.prevent="downloadFile(file)">
                            <translate>Download</translate>
                          </v-btn>
                          <v-btn v-if="features.edit && file.Type === 'jpg' && !file.Primary" small depressed dark color="secondary-dark"
                                 class="ma-0 action-primary"
                                 @click.stop.prevent="primaryFile(file)">
                            <translate>Primary</translate>
                          </v-btn>
                          <v-btn v-if="features.edit && !file.Sidecar && !file.Primary && file.Root === '/'" small depressed dark color="secondary-dark"
                                 class="ma-0 action-unstack"
                                 @click.stop.prevent="unstackFile(file)">
                            <translate>Unstack</translate>
                          </v-btn>
                          <v-btn v-if="features.edit && !file.Primary" small depressed dark color="secondary-dark"
                                 class="ma-0 action-delete"
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
                      <tr v-if="file.Root.length > 1">
                        <td>
                          <translate>Storage Folder</translate>
                        </td>
                        <td>{{ file.Root | capitalize }}</td>
                      </tr>
                      <tr v-if="file.Name">
                        <td>
                          <translate>Name</translate>
                        </td>
                        <td>{{ file.Name }}</td>
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
                      <tr v-if="file.Type">
                        <td>
                          <translate>Type</translate>
                        </td>
                        <td>{{ file.typeInfo() }}</td>
                      </tr>
                      <tr v-if="file.Codec">
                        <td>
                          <translate>Codec</translate>
                        </td>
                        <td>{{ file.Codec | uppercase }}</td>
                      </tr>
                      <tr v-if="file.Primary">
                        <td>
                          <translate>Primary</translate>
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
                        <td>{{ file.AspectRatio }}</td>
                      </tr>
                      <tr v-if="file.MainColor">
                        <td>
                          <translate>Main Color</translate>
                        </td>
                        <td>{{ file.MainColor | capitalize }}</td>
                      </tr>
                      <tr v-if="file.Type === 'jpg'">
                        <td>
                          <translate>Chroma</translate>
                        </td>
                        <td>{{ file.Chroma }} / 100</td>
                      </tr>
                      <tr v-if="file.Error">
                        <td>
                          <translate>Error</translate>
                        </td>
                        <td>{{ file.Error }}</td>
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
    </v-expansion-panel>
    <p-file-delete-dialog :show="deleteFile.dialog" @cancel="closeDeleteDialog"
                          @confirm="confirmDeleteFile"></p-file-delete-dialog>
  </div>
</template>

<script>
import Thumb from "model/thumb";
import {DateTime} from "luxon";

export default {
  name: 'PTabPhotoFiles',
  props: {
    model: Object,
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
    openFile(file) {
      this.$viewer.show([Thumb.fromFile(this.model, file)], 0);
    },
    downloadFile(file) {
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
