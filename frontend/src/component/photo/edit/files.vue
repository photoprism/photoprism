<template>
  <div class="p-tab p-tab-photo-files">
    <v-expansion-panels v-model="expanded" class="pa-0 elevation-0" variant="accordion" multiple>
      <v-expansion-panel
        v-for="file in view.model.fileModels().filter((f) => !f.Missing)"
        :key="file.UID"
        tabindex="1"
        style="margin-top: 1px"
        class="pa-0 elevation-0"
      >
        <v-expansion-panel-title>
          <div class="text-caption font-weight-bold filename">
            {{ file.baseName(70) }}
          </div>
        </v-expansion-panel-title>
        <v-expansion-panel-text>
          <v-card tile>
            <v-card-text class="pa-0">
              <div class="pa-0">
                <v-alert v-if="file.Error" type="error" class="my-0 text-capitalize">
                  {{ file.Error }}
                </v-alert>
                <v-row class="d-flex align-stretch" align="center" justify="center">
                  <v-col cols="12" class="pa-0 flex-grow-1">
                    <div class="v-table__overflow">
                      <v-table tile hover density="compact" class="photo-files d-flex bg-table">
                        <tbody>
                          <tr v-if="file.FileType === 'jpg' || file.FileType === 'png'">
                            <td>
                              {{ $gettext(`Preview`) }}
                            </td>
                            <td>
                              <v-img
                                :src="file.thumbnailUrl('tile_224')"
                                aspect-ratio="1"
                                max-width="150"
                                max-height="150"
                                rounded="4"
                                class="card elevation-0 clickable my-1"
                                @click.exact="openFile(file)"
                              ></v-img>
                            </td>
                          </tr>
                          <tr>
                            <td>
                              {{ $gettext(`Actions`) }}
                            </td>
                            <td>
                              <div class="action-buttons">
                                <v-btn
                                  v-if="features.download"
                                  density="comfortable"
                                  variant="flat"
                                  color="highlight"
                                  class="btn-action action-download"
                                  :disabled="busy"
                                  @click.stop.prevent="downloadFile(file)"
                                >
                                  {{ $gettext(`Download`) }}
                                  <v-icon icon="mdi-download" size="18" end></v-icon>
                                </v-btn>
                                <v-btn
                                  v-if="
                                    features.edit &&
                                    (file.FileType === 'jpg' || file.FileType === 'png') &&
                                    !file.Error &&
                                    !file.Primary
                                  "
                                  density="comfortable"
                                  variant="flat"
                                  color="highlight"
                                  class="btn-action action-primary"
                                  :disabled="busy"
                                  @click.stop.prevent="setPrimaryFile(file)"
                                >
                                  {{ $gettext(`Primary`) }}
                                  <v-icon icon="mdi-image" size="18" end></v-icon>
                                </v-btn>
                                <v-btn
                                  v-if="
                                    features.edit && !file.Sidecar && !file.Error && !file.Primary && file.Root === '/'
                                  "
                                  density="comfortable"
                                  variant="flat"
                                  color="highlight"
                                  class="btn-action action-unstack"
                                  :disabled="busy"
                                  @click.stop.prevent="unstackFile(file)"
                                >
                                  {{ $gettext(`Unstack`) }}
                                  <v-icon icon="mdi-undo-variant" size="18" end></v-icon>
                                </v-btn>
                                <v-btn
                                  v-if="features.delete && !file.Primary"
                                  density="comfortable"
                                  variant="flat"
                                  color="highlight"
                                  class="btn-action action-delete"
                                  :disabled="busy"
                                  @click.stop.prevent="showDeleteDialog(file)"
                                >
                                  {{ $gettext(`Delete`) }}
                                  <v-icon icon="mdi-delete" size="18" end></v-icon>
                                </v-btn>
                                <v-btn
                                  v-if="experimental && canAccessPrivate && file.Primary"
                                  density="comfortable"
                                  variant="flat"
                                  color="highlight"
                                  class="btn-action action-browse action-folder action-open-folder"
                                  @click.stop.prevent="openFolder(file)"
                                >
                                  <v-icon icon="mdi-folder" size="18" start></v-icon>
                                  {{ $gettext(`Browse`) }}
                                </v-btn>
                              </div>
                            </td>
                          </tr>
                          <tr>
                            <td title="Unique ID">UID</td>
                            <td class="text-break">
                              <span class="cursor-copy text-uppercase" @click.stop.prevent="$util.copyText(file.UID)">{{
                                file.UID
                              }}</span>
                            </td>
                          </tr>
                          <tr v-if="file.InstanceID" title="XMP">
                            <td>
                              {{ $gettext(`Instance ID`) }}
                            </td>
                            <td class="text-break">
                              <span
                                class="clickable text-uppercase"
                                @click.stop.prevent="$util.copyText(file.InstanceID)"
                                >{{ file.InstanceID }}</span
                              >
                            </td>
                          </tr>
                          <tr>
                            <td title="SHA-1">
                              {{ $gettext(`Hash`) }}
                            </td>
                            <td class="text-break">
                              <span class="cursor-copy text-break" @click.stop.prevent="$util.copyText(file.Hash)">{{
                                file.Hash
                              }}</span>
                            </td>
                          </tr>
                          <tr v-if="file.Name">
                            <td>
                              {{ $gettext(`Filename`) }}
                            </td>
                            <td class="text-break">
                              <span class="cursor-copy" @click.stop.prevent="$util.copyText(file.Name)">{{
                                file.Name
                              }}</span>
                            </td>
                          </tr>
                          <tr v-if="file.Root">
                            <td>
                              {{ $gettext(`Storage`) }}
                            </td>
                            <td>{{ file.storageInfo() }}</td>
                          </tr>
                          <tr v-if="file.OriginalName">
                            <td>
                              {{ $gettext(`Original Name`) }}
                            </td>
                            <td class="text-break">
                              <span class="cursor-copy" @click.stop.prevent="$util.copyText(file.OriginalName)">{{
                                file.OriginalName
                              }}</span>
                            </td>
                          </tr>
                          <tr v-if="file.FileType">
                            <td>
                              {{ $gettext(`Type`) }}
                            </td>
                            <td class="text-break">
                              <span v-tooltip="file?.Mime">{{ file.typeInfo() }}</span>
                            </td>
                          </tr>
                          <tr>
                            <td>
                              {{ $gettext(`Size`) }}
                            </td>
                            <td>
                              <span v-tooltip="Math.ceil(file?.Size / 1024).toLocaleString() + ' KB'">{{
                                file.sizeInfo()
                              }}</span>
                            </td>
                          </tr>
                          <tr v-if="file.Pages">
                            <td>
                              {{ $gettext(`Pages`) }}
                            </td>
                            <td>{{ file.Pages }}</td>
                          </tr>
                          <tr v-if="file.Software">
                            <td>
                              {{ $gettext(`Software`) }}
                            </td>
                            <td class="text-break">{{ file.Software }}</td>
                          </tr>
                          <tr v-if="file.isAnimated()">
                            <td>
                              {{ $gettext(`Animated`) }}
                            </td>
                            <td>
                              {{ $gettext(`Yes`) }}
                            </td>
                          </tr>
                          <tr v-if="file.Codec && file.Codec !== file.FileType">
                            <td>
                              {{ $gettext(`Codec`) }}
                            </td>
                            <td class="text-break">{{ codecName(file) }}</td>
                          </tr>
                          <tr v-if="file.Duration && file.Duration > 0">
                            <td>
                              {{ $gettext(`Duration`) }}
                            </td>
                            <td>{{ formatDuration(file) }}</td>
                          </tr>
                          <tr v-if="file.Frames">
                            <td>
                              {{ $gettext(`Frames`) }}
                            </td>
                            <td>{{ file.Frames }}</td>
                          </tr>
                          <tr v-if="file.FPS">
                            <td>
                              {{ $gettext(`FPS`) }}
                            </td>
                            <td>{{ file.FPS.toFixed(1) }}</td>
                          </tr>
                          <tr v-if="file.Primary">
                            <td>
                              {{ $gettext(`Primary`) }}
                            </td>
                            <td>
                              {{ $gettext(`Yes`) }}
                            </td>
                          </tr>
                          <tr v-if="file.HDR">
                            <td>
                              {{ $gettext(`High Dynamic Range (HDR)`) }}
                            </td>
                            <td>
                              {{ $gettext(`Yes`) }}
                            </td>
                          </tr>
                          <tr v-if="file.Portrait">
                            <td>
                              {{ $gettext(`Portrait`) }}
                            </td>
                            <td>
                              {{ $gettext(`Yes`) }}
                            </td>
                          </tr>
                          <tr v-if="file.Projection">
                            <td>
                              {{ $gettext(`Projection`) }}
                            </td>
                            <td class="text-capitalize">{{ file.Projection }}</td>
                          </tr>
                          <tr v-if="file.AspectRatio">
                            <td>
                              {{ $gettext(`Aspect Ratio`) }}
                            </td>
                            <td>{{ file.AspectRatio }} : 1</td>
                          </tr>
                          <tr v-if="file.Orientation">
                            <td>
                              {{ $gettext(`Orientation`) }}
                            </td>
                            <td>
                              <v-select
                                v-model="file.Orientation"
                                autocomplete="off"
                                hide-details
                                variant="solo"
                                max-width="120"
                                bg-color="transparent"
                                density="compact"
                                :items="options.Orientations()"
                                item-title="text"
                                item-value="value"
                                :list-props="{ density: 'compact' }"
                                :readonly="
                                  readonly ||
                                  !features.edit ||
                                  file.Error ||
                                  (file.Frames && file.Frames > 1) ||
                                  (file.Duration && file.Duration > 1) ||
                                  (file.FileType !== 'jpg' && file.FileType !== 'png')
                                "
                                :disabled="busy"
                                class="input-orientation"
                                @update:model-value="changeOrientation(file)"
                              >
                                <template #selection="{ item }">
                                  <v-icon :class="orientationClass(item)">mdi-account-box-outline</v-icon>
                                  <span>{{ item.title }}</span>
                                </template>
                                <template #item="{ props, item }">
                                  <v-list-item v-bind="props">
                                    <template #prepend>
                                      <v-icon :class="orientationClass(item)">mdi-account-box-outline</v-icon>
                                    </template>
                                  </v-list-item>
                                </template>
                              </v-select>
                            </td>
                          </tr>
                          <tr v-if="file.ColorProfile">
                            <td>
                              {{ $gettext(`Color Profile`) }}
                            </td>
                            <td class="text-break">{{ file.ColorProfile }}</td>
                          </tr>
                          <tr v-if="file.MainColor">
                            <td>
                              {{ $gettext(`Main Color`) }}
                            </td>
                            <td class="text-capitalize">{{ file.MainColor }}</td>
                          </tr>
                          <tr v-if="file?.Chroma > 0">
                            <td>
                              {{ $gettext(`Chroma`) }}
                            </td>
                            <td>
                              <v-progress-linear
                                v-tooltip="`${file.Chroma}%`"
                                :model-value="file.Chroma"
                                color="surface-variant"
                                style="max-width: 300px"
                              ></v-progress-linear>
                            </td>
                          </tr>
                          <tr v-if="file.Missing">
                            <td>
                              {{ $gettext(`Missing`) }}
                            </td>
                            <td>
                              {{ $gettext(`Yes`) }}
                            </td>
                          </tr>
                          <tr>
                            <td>
                              {{ $gettext(`Added`) }}
                            </td>
                            <td class="text-break">
                              {{ formatTime(file.CreatedAt) }}
                              {{ $gettext(`in`) }}
                              {{ $util.formatNs(file.CreatedIn) }}
                            </td>
                          </tr>
                          <tr v-if="file.UpdatedIn">
                            <td>
                              {{ $gettext(`Updated`) }}
                            </td>
                            <td class="text-break">
                              {{ formatTime(file.UpdatedAt) }}
                              {{ $gettext(`in`) }}
                              {{ $util.formatNs(file.UpdatedIn) }}
                            </td>
                          </tr>
                        </tbody>
                      </v-table>
                    </div>
                  </v-col>
                </v-row>
              </div>
            </v-card-text>
          </v-card>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
    <p-file-delete-dialog
      :visible="deleteFile.dialog"
      @close="closeDeleteDialog"
      @confirm="confirmDeleteFile"
    ></p-file-delete-dialog>
  </div>
</template>

<script>
import Thumb from "model/thumb";
import { DateTime } from "luxon";
import $notify from "common/notify";
import * as options from "options/options";

export default {
  name: "PTabPhotoFiles",
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      view: this.$view.data(),
      expanded: [0],
      deleteFile: {
        dialog: false,
        file: null,
      },
      features: this.$config.getSettings().features,
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      experimental: this.$config.get("experimental"),
      canAccessPrivate:
        this.$config.allow("photos", "access_library") && this.$config.allow("photos", "access_private"),
      options: options,
      busy: false,
      rtl: this.$isRtl,
      listColumns: [
        {
          title: this.$gettext("Primary"),
          key: "Primary",
          sortable: false,
          align: "center",
          class: "p-col-primary",
        },
        { title: this.$gettext("Name"), key: "Name", sortable: false, align: "left" },
        {
          title: this.$gettext("Dimensions"),
          headerProps: {
            class: "hidden-sm-and-down",
          },
          key: "",
          sortable: false,
        },
        {
          title: this.$gettext("Size"),
          headerProps: {
            class: "hidden-xs",
          },
          key: "Size",
          sortable: false,
        },
        { title: this.$gettext("Type"), key: "", sortable: false, align: "left" },
        { title: this.$gettext("Status"), key: "", sortable: false, align: "left" },
      ],
    };
  },
  methods: {
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

      return this.$util.formatDuration(file.Duration);
    },
    fileType(file) {
      if (!file || !file.FileType) {
        return "";
      }

      return this.$util.fileType(file.FileType);
    },
    codecName(file) {
      if (!file || !file.Codec) {
        return "";
      }

      return this.$util.codecName(file.Codec);
    },
    openFile(file) {
      this.$lightbox.openModels([Thumb.fromFile(this.view.model, file)], 0);
    },
    openFolder(file) {
      if (!file) {
        return "#";
      }

      const name = file.Name;

      // "#" chars in path names must be explicitly escaped,
      // see https://github.com/photoprism/photoprism/issues/3695
      const path = name.substring(0, name.lastIndexOf("/")).replaceAll(":", "%3A").replaceAll("#", "%23");
      const route = { path: "/index/files/" + path };

      if (this.$isMobile) {
        this.$emit("close");
        this.$router.push(route);
      } else {
        // Open in a new tab on desktop browsers.
        const routeUrl = this.$router.resolve(route).href;

        if (routeUrl) {
          window.open(routeUrl, "_blank");
        }
      }
    },
    downloadFile(file) {
      $notify.success(this.$gettext("Downloading…"));

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
        this.view.model.deleteFile(this.deleteFile.file.UID).finally(() => this.closeDeleteDialog());
      } else {
        this.closeDeleteDialog();
      }
    },
    unstackFile(file) {
      if (!file || !this.view?.model) {
        return;
      }

      this.view.model.unstackFile(file.UID);
    },
    setPrimaryFile(file) {
      if (!file || !this.view?.model) {
        return;
      }

      this.view.model.setPrimaryFile(file.UID);
    },
    changeOrientation(file) {
      if (!file || !this.view?.model) {
        return;
      }

      this.busy = true;

      this.view.model
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
