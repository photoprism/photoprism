<template>
  <v-dialog :value="show" persistent max-width="500" class="p-account-edit-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-card-title class="pa-2">
        <v-row v-if="scope === 'sharing'" class="py-2 pr-0 pl-2">
          <v-col cols="9">
            <h3 class="headline ma-0 pa-0">
              {{ $gettext("Manual Upload") }}
            </h3>
          </v-col>
          <v-col cols="3" class="text-xs-right">
            <v-switch v-model="model.AccShare" color="secondary-dark" :true-value="true" :false-value="false" :disabled="model.AccType !== 'webdav'" class="ma-0 hidden-xs-only float-right" hide-details></v-switch>
            <v-switch v-model="model.AccShare" color="secondary-dark" :true-value="true" :false-value="false" :disabled="model.AccType !== 'webdav'" class="ma-0 hidden-sm-and-up float-right" hide-details></v-switch>
          </v-col>
        </v-row>
        <v-row v-else-if="scope === 'sync'" class="pa-2">
          <v-col cols="9">
            <h3 class="headline ma-0 pa-0">
              {{ $gettext("Remote Sync") }}
            </h3>
          </v-col>
          <v-col cols="3" class="text-xs-right">
            <v-switch v-model="model.AccSync" color="secondary-dark" :true-value="true" :false-value="false" :disabled="model.AccType !== 'webdav'" class="mt-0 hidden-xs-only float-right" hide-details flat></v-switch>
            <v-switch v-model="model.AccSync" color="secondary-dark" :true-value="true" :false-value="false" :disabled="model.AccType !== 'webdav'" class="mt-0 hidden-sm-and-up float-right" hide-details flat></v-switch>
          </v-col>
        </v-row>
        <v-row v-else class="pt-2 pr-0 pl-2">
          <v-col cols="10">
            <h3 class="headline ma-0 pa-0">
              {{ $gettext("Edit Account") }}
            </h3>
          </v-col>
          <v-col cols="2" class="text-xs-right">
            <v-btn icon text :ripple="false" class="action-remove mt-0" @click.stop.prevent="remove()">
              <v-icon color="secondary-dark">delete</v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-card-text class="py-0 px-2">
        <v-row v-if="scope === 'sharing'">
          <v-col cols="12" class="pa-2">
            <v-autocomplete
              v-model="model.SharePath"
              color="secondary-dark"
              hide-details
              hide-no-data
              filled
              flat
              autocomplete="off"
              hint="Folder"
              :search-input.sync="search"
              :items="pathItems"
              :loading="loading"
              item-text="abs"
              item-value="abs"
              :label="$gettext('Default Folder')"
              :disabled="!model.AccShare || loading"
            >
            </v-autocomplete>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2 input-share-size">
            <v-select v-model="model.ShareSize" :disabled="!model.AccShare" :label="$gettext('Size')" autocomplete="off" hide-details filled flat color="secondary-dark" item-text="text" item-value="value" :items="items.sizes"></v-select>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-select v-model="model.ShareExpires" :disabled="!model.AccShare" :label="$gettext('Expires')" autocomplete="off" hide-details filled flat color="secondary-dark" item-text="text" item-value="value" :items="options.Expires()"></v-select>
          </v-col>
        </v-row>
        <v-row v-else-if="scope === 'sync'">
          <v-col cols="12" sm="6" class="pa-2">
            <v-autocomplete
              v-model="model.SyncPath"
              color="secondary-dark"
              hide-details
              hide-no-data
              filled
              flat
              autocomplete="off"
              :hint="$gettext('Folder')"
              :search-input.sync="search"
              :items="pathItems"
              :loading="loading"
              item-text="abs"
              item-value="abs"
              :label="$gettext('Folder')"
              :disabled="!model.AccSync || loading"
            >
            </v-autocomplete>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-select v-model="model.SyncInterval" :disabled="!model.AccSync" :label="$gettext('Interval')" autocomplete="off" hide-details filled flat color="secondary-dark" item-text="text" item-value="value" :items="options.Intervals()"></v-select>
          </v-col>
          <v-col cols="12" sm="6" class="px-2">
            <v-checkbox v-model="model.SyncDownload" :disabled="!model.AccSync || readonly" hide-details flat color="secondary-dark" on-icon="radio_button_checked" off-icon="radio_button_unchecked" :label="$gettext('Download remote files')" @change="onChangeSync('download')"></v-checkbox>
          </v-col>
          <v-col cols="12" sm="6" class="px-2">
            <v-checkbox v-model="model.SyncFilenames" :disabled="!model.AccSync" hide-details flat color="secondary-dark" :label="$gettext('Preserve filenames')"></v-checkbox>
          </v-col>
          <v-col cols="12" sm="6" class="px-2">
            <v-checkbox v-model="model.SyncUpload" :disabled="!model.AccSync" hide-details flat color="secondary-dark" on-icon="radio_button_checked" off-icon="radio_button_unchecked" :label="$gettext('Upload local files')" @change="onChangeSync('upload')"></v-checkbox>
          </v-col>
          <v-col cols="12" sm="6" class="px-2">
            <v-checkbox v-model="model.SyncRaw" :disabled="!model.AccSync" hide-details flat color="secondary-dark" :label="$gettext('Sync raw and video files')"></v-checkbox>
          </v-col>
        </v-row>
        <v-row v-else class="pt-0">
          <v-col cols="12" class="pa-2">
            <v-text-field v-model="model.AccName" hide-details autofocus filled flat autocomplete="off" :label="$gettext('Name')" placeholder="" color="secondary-dark" required></v-text-field>
          </v-col>
          <v-col cols="12" class="pa-2">
            <v-text-field v-model="model.AccURL" hide-details filled flat autocomplete="off" :label="$gettext('Service URL')" placeholder="https://www.example.com/" color="secondary-dark"></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-text-field v-model="model.AccUser" hide-details filled flat autocomplete="off" :label="$gettext('Username')" placeholder="optional" color="secondary-dark"></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-text-field
              v-model="model.AccPass"
              hide-details
              filled
              flat
              autocomplete="new-password"
              :label="$gettext('Password')"
              placeholder="optional"
              color="secondary-dark"
              :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showPassword ? 'text' : 'password'"
              @click:append="showPassword = !showPassword"
            ></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2">
            <v-text-field v-model="model.AccKey" hide-details filled flat autocomplete="off" :label="$gettext('API Key')" placeholder="optional" color="secondary-dark" required></v-text-field>
          </v-col>
          <v-col cols="12" sm="6" class="pa-2 input-account-type">
            <v-select v-model="model.AccType" :label="$gettext('Type')" autocomplete="off" hide-details filled flat color="secondary-dark" item-text="text" item-value="value" :items="items.types"> </v-select>
          </v-col>
          <v-col cols="12" sm="6" class="px-2">
            <v-select v-model="model.AccTimeout" :label="$gettext('Timeout')" autocomplete="off" hide-details filled flat color="secondary-dark" item-text="text" item-value="value" :items="options.Timeouts()"> </v-select>
          </v-col>
          <v-col cols="12" sm="6" class="px-2">
            <v-select v-model="model.RetryLimit" :label="$gettext('Retry Limit')" autocomplete="off" hide-details filled flat color="secondary-dark" item-text="text" item-value="value" :items="options.RetryLimits()"> </v-select>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions class="pt-0 pb-2 px-2">
        <v-row class="pa-2">
          <v-col cols="12" class="text-xs-right pt-6 pb-0">
            <v-btn depressed color="secondary-light" class="action-cancel ml-2" @click.stop="cancel">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn depressed dark color="primary-button" class="action-save compact" @click.stop="save">
              <translate>Save</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import * as options from "options/options";

export default {
  name: "PAccountEditDialog",
  props: {
    show: Boolean,
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
    const thumbs = this.$config.values.thumbs;

    return {
      options: options,
      showPassword: false,
      loading: false,
      search: null,
      path: "/",
      paths: [{ abs: "/" }],
      pathItems: [],
      newPath: "",
      items: {
        thumbs: thumbs,
        sizes: this.sizes(thumbs),
        types: [
          { value: "web", text: "Web" },
          { value: "webdav", text: "WebDAV / Nextcloud" },
          { value: "facebook", text: "Facebook" },
          { value: "twitter", text: "Twitter" },
          { value: "flickr", text: "Flickr" },
          { value: "instagram", text: "Instagram" },
          { value: "eyeem", text: "EyeEm" },
          { value: "telegram", text: "Telegram" },
          { value: "whatsapp", text: "WhatsApp" },
          { value: "gphotos", text: "Google Photos" },
          { value: "gdrive", text: "Google Drive" },
          { value: "onedrive", text: "Microsoft OneDrive" },
        ],
      },
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
    show: function (show) {
      if (show) {
        this.onChange();
      }
    },
  },
  methods: {
    cancel() {
      this.$emit("cancel");
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
    sizes(thumbs) {
      const result = [{ text: this.$gettext("Originals"), value: "" }];

      for (let i = 0; i < thumbs.length; i++) {
        let t = thumbs[i];

        result.push({ text: t.w + " × " + t.h, value: t.size });
      }

      return result;
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
      this.onChangeSync();
      this.paths = [{ abs: "/" }];

      this.loading = true;
      this.model
        .Folders()
        .then((p) => {
          for (let i = 0; i < p.length; i++) {
            this.paths.push(p[i]);
          }

          this.pathItems = [...this.paths];
          this.path = this.model.SharePath;
        })
        .finally(() => (this.loading = false));
    },
  },
};
</script>
