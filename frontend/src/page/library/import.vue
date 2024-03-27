<template>
  <div class="p-tab p-tab-import">
    <v-form ref="form" class="p-photo-import" lazy-validation dense @submit.prevent="submit">
      <v-container fluid>
        <p class="subheading">
          <span v-if="fileName" class="break-word"><translate :translate-params="{ name: fileName }">Importing %{name}…</translate></span>
          <span v-else-if="busy"><translate>Importing files to originals…</translate></span>
          <span v-else-if="completed"><translate>Done.</translate></span>
          <span v-else><translate>Press button to start importing…</translate></span>
        </p>

        <v-autocomplete
          v-model="settings.import.path"
          color="secondary-dark"
          class="my-3 input-import-folder"
          hide-details
          hide-no-data
          flat
          solo
          browser-autocomplete="off"
          :items="dirs"
          :loading="loading"
          :disabled="busy || !ready"
          item-text="name"
          item-value="path"
          @change="onChange"
          @focus="onFocus"
        >
        </v-autocomplete>

        <p class="options">
          <v-progress-linear color="secondary-dark" height="1.5em" :value="completed" :indeterminate="busy"></v-progress-linear>
        </p>

        <v-layout wrap align-top class="pb-2">
          <v-flex xs12 class="px-2 pb-2 pt-2">
            <v-checkbox
              v-model="settings.import.move"
              :disabled="busy || !ready"
              class="ma-0 pa-0"
              color="secondary-dark"
              :label="$gettext('Move Files')"
              :hint="$gettext('Remove imported files to save storage. Unsupported file types will never be deleted, they remain in their current location.')"
              prepend-icon="delete"
              persistent-hint
              @change="onChange"
            >
            </v-checkbox>
          </v-flex>
          <v-flex xs12 class="px-2 pb-2 pt-2">
            <p class="body-1 pt-2">
              <translate>Imported files will be sorted by date and given a unique name to avoid duplicates.</translate>
              <translate>JPEGs and thumbnails are automatically rendered as needed.</translate>
              <translate>Original file names will be stored and indexed.</translate>
              <translate>Note you may manually manage your originals folder and importing is optional.</translate>
            </p>
          </v-flex>
        </v-layout>

        <v-btn :disabled="!busy || !ready" color="primary-button" class="white--text ml-0 action-cancel" depressed @click.stop="cancelImport()">
          <translate>Cancel</translate>
        </v-btn>

        <v-btn v-if="!$config.values.readonly && $config.feature('upload')" :disabled="busy || !ready" color="primary-button" class="white--text ml-0 hidden-xs-only action-upload" depressed @click.stop="showUpload()">
          <translate>Upload</translate>
          <v-icon :right="!rtl" :left="rtl" dark>cloud_upload</v-icon>
        </v-btn>

        <v-btn :disabled="busy || !ready" color="primary-button" class="white--text ml-0 mt-2 action-import" depressed @click.stop="startImport()">
          <translate>Import</translate>
          <v-icon :right="!rtl" :left="rtl" dark>sync</v-icon>
        </v-btn>
      </v-container>
    </v-form>
  </div>
</template>

<script>
import Api from "common/api";
import Axios from "axios";
import Notify from "common/notify";
import Event from "pubsub-js";
import Settings from "model/settings";
import Util from "common/util";
import { Folder, RootImport } from "model/folder";

export default {
  name: "PTabImport",
  data() {
    const root = { path: "/", name: this.$gettext("All files from import folder") };

    return {
      ready: !this.$config.loading(),
      settings: new Settings(this.$config.settings()),
      started: false,
      busy: false,
      loading: false,
      completed: 0,
      subscriptionId: "",
      fileName: "",
      source: null,
      root: root,
      dirs: [root],
      rtl: this.$rtl,
    };
  },
  created() {
    this.subscriptionId = Event.subscribe("import", this.handleEvent);
    this.load();
  },
  destroyed() {
    Event.unsubscribe(this.subscriptionId);
  },
  methods: {
    load() {
      this.$config.load().then(() => {
        this.settings.setValues(this.$config.settings());
        this.dirs = [this.root];

        if (this.settings.import.path !== this.root.path) {
          this.dirs.push({
            path: this.settings.import.path,
            name: "/" + Util.truncate(this.settings.import.path, 100, "…"),
          });
        }

        this.ready = true;
      });
    },
    onChange() {
      if (!this.$config.values.disable.settings) {
        this.settings.save();
      }
    },
    onFocus() {
      if (this.dirs.length > 2 || this.loading) {
        return;
      }

      this.loading = true;

      Folder.findAllUncached(RootImport)
        .then((r) => {
          const folders = r.models ? r.models : [];
          const currentPath = this.settings.import.path;
          let found = currentPath === this.root.path;

          this.dirs = [this.root];

          for (let i = 0; i < folders.length; i++) {
            if (currentPath === folders[i].Path) {
              found = true;
            }

            this.dirs.push({ path: folders[i].Path, name: "/" + Util.truncate(folders[i].Path, 100, "…") });
          }

          if (!found) {
            this.settings.import.path = this.root.path;
          }
        })
        .finally(() => (this.loading = false));
    },
    showUpload() {
      Event.publish("dialog.upload");
    },
    submit() {
      // DO NOTHING
    },
    cancelImport() {
      Api.delete("import");
    },
    startImport() {
      this.source = Axios.CancelToken.source();
      this.started = Date.now();
      this.busy = true;
      this.completed = 0;
      this.fileName = "";

      const ctx = this;
      Notify.blockUI();

      Api.post("import", this.settings.import, { cancelToken: this.source.token })
        .then(function () {
          Notify.unblockUI();
          ctx.busy = false;
          ctx.completed = 100;
          ctx.fileName = "";
        })
        .catch(function (e) {
          Notify.unblockUI();

          if (Axios.isCancel(e)) {
            // run in background
            return;
          }

          Notify.error(this.$gettext("Import failed"));

          ctx.busy = false;
          ctx.completed = 0;
          ctx.fileName = "";
        });
    },
    handleEvent(ev, data) {
      if (this.source) {
        this.source.cancel("run in background");
        this.source = null;
        Notify.unblockUI();
      }

      const type = ev.split(".")[1];

      switch (type) {
        case "file":
          this.busy = true;
          this.completed = 0;
          this.fileName = data.baseName;
          break;
        case "completed":
          this.busy = false;
          this.completed = 100;
          this.fileName = "";
          break;
        default:
          console.log(data);
      }
    },
  },
};
</script>
