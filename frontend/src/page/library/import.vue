<template>
  <div class="p-tab p-tab-import">
    <v-form ref="form" class="p-photo-import" validate-on="blur" @submit.prevent="submit">
      <v-container fluid>
        <p class="text-subtitle-1">
          <span v-if="fileName" class="break-word"><translate :translate-params="{ name: fileName }">Importing %{name}…</translate></span>
          <span v-else-if="busy"><translate>Importing files to originals…</translate></span>
          <span v-else-if="completed"><translate>Done.</translate></span>
          <span v-else><translate>Press button to start importing…</translate></span>
        </p>

        <v-autocomplete
          v-model="settings.import.path"
          color="surface-variant"
          class="mt-6 input-import-folder"
          hide-details
          hide-no-data
          flat
          variant="solo-filled"
          autocomplete="off"
          :items="dirs"
          item-title="name"
          item-value="path"
          :loading="loading"
          :disabled="busy || !ready"
          @update:model-value="onChange"
          @focus="onFocus"
        >
        </v-autocomplete>
        <v-progress-linear :model-value="completed" :indeterminate="busy"></v-progress-linear>

        <div class="action-controls">
          <v-checkbox
            v-model="settings.import.move"
            :disabled="busy || !ready"
            class="ma-0 pa-0"
            color="surface-variant"
            density="compact"
            :label="$gettext('Move Files')"
            :hint="$gettext('Remove imported files to save storage. Unsupported file types will never be deleted, they remain in their current location.')"
            prepend-icon="mdi-delete"
            persistent-hint
            @update:model-value="onChange"
          >
          </v-checkbox>
          <p>
            <translate>Imported files will be sorted by date and given a unique name to avoid duplicates.</translate>
            <translate>JPEGs and thumbnails are automatically rendered as needed.</translate>
            <translate>Original file names will be stored and indexed.</translate>
            <translate>Note you may manually manage your originals folder and importing is optional.</translate>
          </p>
        </div>
        <!-- v-row align="start" class="my-3 mx-0" no-gutters>
          <v-col cols="12">
          </v-col>
          <v-col cols="12">
            <p class="text-body-2 py-3">
              <translate>Imported files will be sorted by date and given a unique name to avoid duplicates.</translate>
              <translate>JPEGs and thumbnails are automatically rendered as needed.</translate>
              <translate>Original file names will be stored and indexed.</translate>
              <translate>Note you may manually manage your originals folder and importing is optional.</translate>
            </p>
          </v-col>
        </v-row -->

        <div class="action-buttons">
          <v-btn :disabled="!busy || !ready" variant="flat" color="button" class="action-cancel" @click.stop="cancelImport()">
            <translate>Cancel</translate>
          </v-btn>

          <v-btn v-if="!$config.values.readonly && $config.feature('upload')" :disabled="busy || !ready" variant="flat" color="primary-button" class="hidden-xs action-upload" @click.stop="showUpload()">
            <translate>Upload</translate>
            <v-icon :end="!rtl" :start="rtl">mdi-cloud-upload</v-icon>
          </v-btn>

          <v-btn :disabled="busy || !ready" variant="flat" color="primary-button" class="action-import" @click.stop="startImport()">
            <translate>Import</translate>
            <v-icon :end="!rtl" :start="rtl">mdi-sync</v-icon>
          </v-btn>
        </div>
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
  unmounted() {
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
