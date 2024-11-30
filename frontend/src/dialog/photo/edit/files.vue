<template>
  <div class="p-tab p-tab-photo-files">
    <v-expansion-panels v-model="state" class="pa-0 elevation-0 bg-secondary" variant="accordion">
      <template v-for="file in model.fileModels()">
        <v-expansion-panel v-if="!file.Missing" :key="file.UID" class="pa-0 elevation-0 secondary-light" style="margin-top: 1px" :title="file.baseName(70)">
          <v-expansion-panel-text>
            <v-card>
              <v-card-text class="bg-white pa-0">
                <v-container fluid class="pa-0">
                  <v-alert v-if="file.Error" type="error" class="my-0 text-capitalize">
                    {{ file.Error }}
                  </v-alert>
                  <v-row class="d-flex align-stretch" align="center" justify="center">
                    <v-col cols="12" class="pa-0 flex-grow-1">
                      <div class="v-table__overflow">
                        <v-table class="v-datatable v-table photo-files d-flex">
                          <tbody>
                            <tr v-if="file.FileType === 'jpg' || file.FileType === 'png'">
                              <td>
                                <translate>Preview</translate>
                              </td>
                              <td>
                                <v-img :src="file.thumbnailUrl('tile_224')" aspect-ratio="1" max-width="112" max-height="112" class="card elevation-0 clickable my-1" @click.exact="openFile(file)"></v-img>
                              </td>
                            </tr>
                            <tr>
                              <td>
                                <translate>Actions</translate>
                              </td>
                              <td>
                                <v-btn v-if="features.download" size="small" variant="flat" color="primary-button" class="btn-action action-download ma-1" :disabled="busy" @click.stop.prevent="downloadFile(file)">
                                  <translate>Download</translate>
                                </v-btn>
                                <v-btn
                                  v-if="features.edit && (file.FileType === 'jpg' || file.FileType === 'png') && !file.Error && !file.Primary"
                                  size="small"
                                  variant="flat"
                                  theme="dark"
                                  color="primary-button"
                                  class="btn-action action-primary ma-1"
                                  :disabled="busy"
                                  @click.stop.prevent="primaryFile(file)"
                                >
                                  <translate>Primary</translate>
                                </v-btn>
                                <v-btn v-if="features.edit && !file.Sidecar && !file.Error && !file.Primary && file.Root === '/'" size="small" variant="flat" theme="dark" color="primary-button" class="btn-action action-unstack ma-1" :disabled="busy" @click.stop.prevent="unstackFile(file)">
                                  <translate>Unstack</translate>
                                </v-btn>
                                <v-btn v-if="features.delete && !file.Primary" size="small" variant="flat" theme="dark" color="primary-button" class="btn-action action-delete ma-1" :disabled="busy" @click.stop.prevent="showDeleteDialog(file)">
                                  <translate>Delete</translate>
                                </v-btn>
                                <v-btn v-if="experimental && canAccessPrivate && file.Primary" size="small" variant="flat" theme="dark" color="primary-button" class="btn-action action-open-folder ma-1" :href="folderUrl(file)" target="_blank">
                                  <translate>File Browser</translate>
                                </v-btn>
                              </td>
                            </tr>
                            <tr>
                              <td title="Unique ID"> UID </td>
                              <td>
  <!--                              TODO: change filter-->
  <!--                              <span class="clickable" @click.stop.prevent="copyText(file.UID)">{{ file.UID | uppercase }}</span>-->
                                <span class="clickable" @click.stop.prevent="copyText(file.UID)">{{ file.UID }}</span>
                              </td>
                            </tr>
                            <tr v-if="file.InstanceID" title="XMP">
                              <td>
                                <translate>Instance ID</translate>
                              </td>
                              <td>
  <!--                              TODO: change filter-->
  <!--                              <span class="clickable" @click.stop.prevent="copyText(file.InstanceID)">{{ file.InstanceID | uppercase }}</span></td-->
                                <span class="clickable" @click.stop.prevent="copyText(file.InstanceID)">{{ file.InstanceID }}</span></td
                              >
                            </tr>
                            <tr>
                              <td title="SHA-1">
                                <translate>Hash</translate>
                              </td>
                              <td
                                ><span class="clickable" @click.stop.prevent="copyText(file.Hash)">{{ file.Hash }}</span></td
                              >
                            </tr>
                            <tr v-if="file.Name">
                              <td>
                                <translate>Filename</translate>
                              </td>
                              <td
                                ><span class="clickable" @click.stop.prevent="copyText(file.Name)">{{ file.Name }}</span></td
                              >
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
                              <td
                                ><span class="clickable" @click.stop.prevent="copyText(file.OriginalName)">{{ file.OriginalName }}</span></td
                              >
                            </tr>
                            <tr>
                              <td>
                                <translate>Size</translate>
                              </td>
                              <td>{{ file.sizeInfo() }}</td>
                            </tr>
                            <tr v-if="file.Software">
                              <td>
                                <translate>Software</translate>
                              </td>
                              <td>{{ file.Software }}</td>
                            </tr>
                            <tr v-if="file.FileType">
                              <td>
                                <translate>Type</translate>
                              </td>
                              <td>{{ file.typeInfo() }}</td>
                            </tr>
                            <tr v-if="file.isAnimated()">
                              <td>
                                <translate>Animated</translate>
                              </td>
                              <td>
                                <translate>Yes</translate>
                              </td>
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
  <!--                            TODO: change filter-->
  <!--                            <td>{{ file.Projection | capitalize }}</td>-->
                              <td>{{ file.Projection }}</td>
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
                                <v-select
                                  v-model="file.Orientation"
                                  flat
                                  variant="solo"
                                  autocomplete="off"
                                  hide-details
                                  color="surface-variant"
                                  :items="options.Orientations()"
                                  item-title="text"
                                  item-value="value"
                                  :readonly="readonly || !features.edit || file.Error || (file.Frames && file.Frames > 1) || (file.Duration && file.Duration > 1) || (file.FileType !== 'jpg' && file.FileType !== 'png')"
                                  :disabled="busy"
                                  class="input-orientation"
                                  @update:model-value="changeOrientation(file)"
                                >
                                  <template #selection="{ item }">
                                    <span :title="item.text"><v-icon :class="orientationClass(item)">mdi-account-box-outline</v-icon></span>
                                  </template>
                                  <template #item="{ item }">
                                    <span :title="item.text"><v-icon :class="orientationClass(item)">mdi-account-box-outline</v-icon></span>
                                  </template>
                                </v-select>
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
  <!--                            TODO: change filter-->
  <!--                            <td>{{ file.MainColor | capitalize }}</td>-->
                              <td>{{ file.MainColor }}</td>
                            </tr>
                            <tr v-if="file.Chroma">
                              <td>
                                <translate>Chroma</translate>
                              </td>
                              <td><v-progress-linear :model-value="file.Chroma" style="max-width: 300px" :title="`${file.Chroma}%`"></v-progress-linear></td>
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
                              <td
                                >{{ formatTime(file.CreatedAt) }}
                                <translate>in</translate>
  <!--                              TODO: change filter-->
  <!--                              {{ Math.round(file.CreatedIn / 1000000) | number("0,0") }} ms-->
                                {{ Math.round(file.CreatedIn / 1000000) }} ms
                              </td>
                            </tr>
                            <tr v-if="file.UpdatedIn">
                              <td>
                                <translate>Updated</translate>
                              </td>
                              <td
                                >{{ formatTime(file.UpdatedAt) }}
                                <translate>in</translate>
  <!--                              TODO: change filter-->
  <!--                              {{ Math.round(file.UpdatedIn / 1000000) | number("0,0") }} ms-->
                                {{ Math.round(file.UpdatedIn / 1000000) }} ms
                              </td>
                            </tr>
                          </tbody>
                        </v-table>
                      </div>
                    </v-col>
                  </v-row>
                </v-container>
              </v-card-text>
            </v-card>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </template>
    </v-expansion-panels>
    <p-file-delete-dialog :show="deleteFile.dialog" @cancel="closeDeleteDialog" @confirm="confirmDeleteFile"></p-file-delete-dialog>
  </div>
</template>

<script>
import Thumb from "model/thumb";
import { DateTime } from "luxon";
import Notify from "common/notify";
import Util from "common/util";
import * as options from "options/options";

export default {
  name: "PTabPhotoFiles",
  props: {
    model: {
      type: Object,
      default: () => {},
    },
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      state: [0],
      deleteFile: {
        dialog: false,
        file: null,
      },
      features: this.$config.settings().features,
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      experimental: this.$config.get("experimental"),
      canAccessPrivate: this.$config.allow("photos", "access_library") && this.$config.allow("photos", "access_private"),
      options: options,
      busy: false,
      rtl: this.$rtl,
      listColumns: [
        {
          text: this.$gettext("Primary"),
          value: "Primary",
          sortable: false,
          align: "center",
          class: "p-col-primary",
        },
        { text: this.$gettext("Name"), value: "Name", sortable: false, align: "left" },
        {
          text: this.$gettext("Dimensions"),
          value: "",
          sortable: false,
          class: "hidden-sm-and-down",
        },
        { text: this.$gettext("Size"), value: "Size", sortable: false, class: "hidden-xs" },
        { text: this.$gettext("Type"), value: "", sortable: false, align: "left" },
        { text: this.$gettext("Status"), value: "", sortable: false, align: "left" },
      ],
    };
  },
  computed: {},
  methods: {
    async copyText(text) {
      if (!text) {
        return;
      }

      try {
        await Util.copyToMachineClipboard(text);
        this.$notify.success(this.$gettext("Copied to clipboard"));
      } catch (error) {
        this.$notify.error(this.$gettext("Failed copying to clipboard"));
      }
    },
    orientationClass(file) {
      if (!file) {
        return [];
      }
      return [`orientation-${file.value}`];
    },
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
    folderUrl(m) {
      if (!m) {
        return "#";
      }

      const name = m.Name;

      // "#" chars in path names must be explicitly escaped,
      // see https://github.com/photoprism/photoprism/issues/3695
      const path = name.substring(0, name.lastIndexOf("/")).replaceAll(":", "%3A").replaceAll("#", "%23");
      return this.$router.resolve({ path: "/index/files/" + path }).href;
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
    changeOrientation(file) {
      if (!file) {
        return;
      }

      this.busy = true;

      this.model
        .changeFileOrientation(file)
        .then(() => {
          this.$notify.success(this.$gettext("Changes successfully saved"));
          this.busy = false;
        })
        .catch(() => {
          this.busy = false;
        });
    },
    formatTime(s) {
      return DateTime.fromISO(s).toLocaleString(DateTime.DATETIME_MED);
    },
    refresh() {},
  },
};
</script>
