<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="500"
    class="p-dialog p-service-edit"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" accept-charset="UTF-8" tabindex="1" @submit.prevent>
      <v-card>
        <v-card-title v-if="scope === 'sharing'" class="d-flex justify-space-between align-center ga-3">
          <h6 class="text-h6">
            {{ $gettext("Manual Upload") }}
          </h6>
          <v-switch v-model="model.AccShare" :disabled="model.AccType !== 'webdav'"></v-switch>
        </v-card-title>
        <v-card-title v-else-if="scope === 'sync'" class="d-flex justify-space-between align-center ga-3">
          <h6 class="text-h6">
            {{ $gettext("Remote Sync") }}
          </h6>
          <v-switch v-model="model.AccSync" :disabled="model.AccType !== 'webdav'"></v-switch>
        </v-card-title>
        <v-card-title v-else class="d-flex justify-space-between align-center ga-3">
          <h6 class="text-h6">
            {{ $gettext("Edit Account") }}
          </h6>
          <v-btn icon variant="text" class="action-remove" @click.stop.prevent="remove()">
            <v-icon color="surface-variant">mdi-delete</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text class="dense">
          <v-row v-if="scope === 'sharing'" dense>
            <v-col cols="12">
              <v-autocomplete
                v-model="model.SharePath"
                hide-details
                hide-no-data
                flat
                autocomplete="off"
                :hint="$gettext('Folder')"
                :search.sync="search"
                :items="pathItems"
                :loading="loading"
                item-title="abs"
                item-value="abs"
                :label="$gettext('Default Folder')"
                :disabled="!model.AccShare || loading"
              >
              </v-autocomplete>
            </v-col>
            <v-col cols="12" sm="6" class="input-share-size">
              <v-select
                v-model="model.ShareSize"
                :disabled="!model.AccShare"
                :label="$gettext('Size')"
                autocomplete="off"
                item-title="text"
                item-value="value"
                :items="options.ThumbSizes()"
              ></v-select>
            </v-col>
            <v-col cols="12" sm="6">
              <v-select
                v-model="model.ShareExpires"
                :disabled="!model.AccShare"
                :label="$gettext('Expires')"
                autocomplete="off"
                item-title="text"
                item-value="value"
                :items="options.Expires()"
              ></v-select>
            </v-col>
          </v-row>
          <v-row v-else-if="scope === 'sync'" dense>
            <v-col cols="12" sm="6">
              <v-autocomplete
                v-model="model.SyncPath"
                hide-details
                hide-no-data
                flat
                autocomplete="off"
                :hint="$gettext('Folder')"
                :search.sync="search"
                :items="pathItems"
                :loading="loading"
                item-title="abs"
                item-value="abs"
                :label="$gettext('Folder')"
                :disabled="!model.AccSync || loading"
              >
              </v-autocomplete>
            </v-col>
            <v-col cols="12" sm="6">
              <v-select
                v-model="model.SyncInterval"
                :disabled="!model.AccSync"
                :label="$gettext('Interval')"
                autocomplete="off"
                hide-details
                flat
                color="surface-variant"
                item-title="text"
                item-value="value"
                :items="options.Intervals()"
              ></v-select>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox
                v-model="model.SyncDownload"
                density="comfortable"
                :disabled="!model.AccSync || readonly"
                hide-details
                true-icon="mdi-radiobox-marked"
                false-icon="mdi-radiobox-blank"
                :label="$gettext('Download remote files')"
                @update:model-value="onChangeSync('download')"
              ></v-checkbox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox
                v-model="model.SyncUpload"
                density="comfortable"
                :disabled="!model.AccSync"
                true-icon="mdi-radiobox-marked"
                false-icon="mdi-radiobox-blank"
                :label="$gettext('Upload local files')"
                hide-details
                @update:model-value="onChangeSync('upload')"
              ></v-checkbox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox
                v-model="model.SyncFilenames"
                density="comfortable"
                :disabled="!model.AccSync"
                :label="$gettext('Preserve filenames')"
                hide-details
              ></v-checkbox>
            </v-col>
            <v-col cols="12" sm="6">
              <v-checkbox
                v-model="model.SyncRaw"
                density="comfortable"
                :disabled="!model.AccSync"
                :label="$gettext('Sync raw and video files')"
                hide-details
              ></v-checkbox>
            </v-col>
          </v-row>
          <v-row v-else dense>
            <v-col cols="12">
              <v-text-field
                v-model="model.AccName"
                autofocus
                autocomplete="off"
                :label="$gettext('Name')"
                placeholder=""
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-text-field
                v-model="model.AccURL"
                autocomplete="off"
                :label="$gettext('Service URL')"
                placeholder="https://www.example.com/"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field v-model="model.AccUser" autocomplete="off" :label="$gettext('Username')"></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="model.AccPass"
                hide-details
                autocomplete="new-password"
                :label="$gettext('Password')"
                placeholder="********"
                :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                :type="showPassword ? 'text' : 'password'"
                @click:append-inner="showPassword = !showPassword"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="model.AccKey"
                hide-details
                flat
                autocomplete="off"
                :label="$gettext('API Key')"
                placeholder="********"
                color="surface-variant"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6" class="input-account-type">
              <v-select
                v-model="model.AccType"
                :label="$gettext('Type')"
                autocomplete="off"
                hide-details
                flat
                color="surface-variant"
                item-title="text"
                item-value="value"
                :items="options.AccountTypes()"
              >
              </v-select>
            </v-col>
            <v-col cols="12" sm="6">
              <v-select
                v-model="model.AccTimeout"
                :label="$gettext('Timeout')"
                autocomplete="off"
                hide-details
                flat
                color="surface-variant"
                item-title="text"
                item-value="value"
                :items="options.Timeouts()"
              >
              </v-select>
            </v-col>
            <v-col cols="12" sm="6">
              <v-select
                v-model="model.RetryLimit"
                :label="$gettext('Retry Limit')"
                autocomplete="off"
                hide-details
                flat
                color="surface-variant"
                item-title="text"
                item-value="value"
                :items="options.RetryLimits()"
              >
              </v-select>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons mt-4">
          <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn variant="flat" color="highlight" class="action-save" @click.stop="save">
            {{ $gettext(`Save`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import * as options from "options/options";

export default {
  name: "PServiceEdit",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    scope: {
      type: String,
      default: "",
    },
    model: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
      options: options,
      showPassword: false,
      loading: false,
      search: null,
      path: "/",
      paths: [{ abs: "/" }],
      pathItems: [],
      newPath: "",
      readonly: this.$config.get("readonly"),
    };
  },
  computed: {},
  watch: {
    search(q) {
      if (this.loading) return;

      const exists = this.paths.findIndex((p) => p.value === q);

      if (exists !== -1 || !q) {
        this.pathItems = this.paths;
        this.newPath = "";
      } else {
        this.newPath = q;
        this.pathItems = this.paths.concat([{ abs: q }]);
      }
    },
    visible: function (show) {
      if (show) {
        this.loading = false;
        this.showPassword = false;
        this.onChange();
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    close() {
      this.$emit("close");
    },
    remove() {
      this.$emit("remove");
    },
    confirm() {
      this.model.AccShare = true;
      this.save();
    },
    disable(prop) {
      this.model[prop] = false;

      this.save();
    },
    enable(prop) {
      this.model[prop] = true;
    },
    save() {
      if (this.loading) {
        this.$notify.busy();
        return;
      }

      this.loading = true;

      this.model.update().then(() => {
        this.loading = false;
        this.$notify.success(this.$gettext("Changes successfully saved"));
        this.$emit("confirm");
      });
    },
    onChangeSync(dir) {
      switch (dir) {
        case "upload":
          this.model.SyncDownload = !this.model.SyncUpload;
          break;
        default:
          this.model.SyncUpload = !this.model.SyncDownload;
      }
    },
    onChange() {
      if (this.loading) {
        return;
      }

      this.onChangeSync();
      this.paths = [{ abs: "/" }];

      this.loading = true;
      this.model
        .Folders()
        .then((p) => {
          if (p && p.length > 0) {
            this.paths = p;
            this.pathItems = [...this.paths];
            this.path = this.model.SharePath;
          }
        })
        .finally(() => {
          this.loading = false;
        });
    },
  },
};
</script>
