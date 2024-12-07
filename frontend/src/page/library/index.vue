<template>
  <div class="p-tab p-tab-index">
    <v-form ref="form" class="p-photo-index" validate-on="blur" @submit.prevent="submit">
      <v-container fluid>
        <p class="text-subtitle-1">
          <span v-if="fileName" class="break-word">{{ action }} {{ fileName }}…</span>
          <span v-else-if="action">{{ action }}…</span>
          <span v-else-if="busy"><translate>Indexing media and sidecar files…</translate></span>
          <span v-else-if="completed"><translate>Done.</translate></span>
          <span v-else><translate>Press button to start indexing…</translate></span>
        </p>

        <v-autocomplete
          v-model="settings.index.path"
          color="surface-variant"
          class="mt-6 input-index-folder"
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
            v-model="settings.index.rescan"
            :disabled="busy || !ready"
            :label="$gettext('Complete Rescan')"
            :hint="$gettext('Re-index all originals, including already indexed and unchanged files.')"
            prepend-icon="mdi-cached"
            persistent-hint
            @update:model-value="onChange"
          >
          </v-checkbox>
          <v-checkbox
            v-if="isAdmin"
            v-model="cleanup"
            :disabled="busy || !ready"
            :label="$gettext('Cleanup')"
            :hint="$gettext('Delete orphaned index entries, sidecar files and thumbnails.')"
            prepend-icon="mdi-delete-sweep"
            persistent-hint
          >
          </v-checkbox>
        </div>

        <div class="action-buttons">
          <v-btn :disabled="!busy || !ready" variant="flat" color="button" class="action-cancel" @click.stop="cancelIndexing()">
            <translate>Cancel</translate>
          </v-btn>

          <v-btn :disabled="busy || !ready" variant="flat" color="primary-button" class="action-index" @click.stop="startIndexing()">
            <translate>Start</translate>
            <v-icon :end="!rtl" :start="rtl">mdi-update</v-icon>
          </v-btn>
        </div>

        <v-alert v-if="ready && !busy && config.count.hidden > 1" color="error" icon="mdi-exclamation" class="mt-6" variant="outlined">
          <translate :translate-params="{ n: config.count.hidden }">The index currently contains %{n} hidden files.</translate>
          <translate>Their format may not be supported, they haven't been converted to JPEG yet or there are duplicates.</translate>
        </v-alert>
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
import { Folder, RootOriginals } from "model/folder";

export default {
  name: "PTabIndex",
  data() {
    const root = { path: "/", name: this.$gettext("All originals") };

    return {
      ready: !this.$config.loading(),
      settings: new Settings(this.$config.settings()),
      readonly: this.$config.get("readonly"),
      config: this.$config.values,
      isAdmin: this.$session.isAdmin(),
      started: false,
      busy: false,
      loading: false,
      completed: 0,
      subscriptionId: "",
      action: "",
      fileName: "",
      cleanup: false,
      source: null,
      root: root,
      dirs: [root],
      rtl: this.$rtl,
    };
  },
  created() {
    this.subscriptionId = Event.subscribe("index", this.handleEvent);
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

        if (this.settings.index.path !== this.root.path) {
          this.dirs.push({
            path: this.settings.index.path,
            name: "/" + Util.truncate(this.settings.index.path, 100, "…"),
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

      Folder.findAll(RootOriginals)
        .then((r) => {
          const folders = r.models ? r.models : [];
          const currentPath = this.settings.index.path;
          let found = currentPath === this.root.path;

          this.dirs = [this.root];

          for (let i = 0; i < folders.length; i++) {
            if (currentPath === folders[i].Path) {
              found = true;
            }

            this.dirs.push({ path: folders[i].Path, name: "/" + Util.truncate(folders[i].Path, 100, "…") });
          }

          if (!found) {
            this.settings.index.path = this.root.path;
          }
        })
        .finally(() => (this.loading = false));
    },
    submit() {
      // DO NOTHING
    },
    cancelIndexing() {
      Api.delete("index");
    },
    startIndexing() {
      this.source = Axios.CancelToken.source();
      this.started = Date.now();
      this.busy = true;
      this.completed = 0;
      this.fileName = "";

      const ctx = this;
      Notify.blockUI();

      // Request parameters.
      const params = {
        path: this.settings.index.path,
        rescan: this.settings.index.rescan,
        cleanup: this.cleanup,
      };

      // Submit POST request.
      Api.post("index", params, { cancelToken: this.source.token })
        .then(function () {
          Notify.unblockUI();
          ctx.busy = false;
          ctx.completed = 100;
          ctx.fileName = "";
        })
        .catch(function (e) {
          Notify.unblockUI();

          if (Axios.isCancel(e)) {
            // Run in background.
            return;
          }

          Notify.error(ctx.$gettext("Indexing failed"));

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
        case "folder":
          this.action = this.$gettext("Indexing");
          this.busy = true;
          this.completed = 0;
          this.fileName = data.filePath;
          break;
        case "indexing":
          this.action = this.$gettext("Indexing");
          this.busy = true;
          this.completed = 0;
          this.fileName = data.fileName;
          break;
        case "updating":
          if (data.step === "stacks") {
            this.action = this.$gettext("Updating stacks");
          } else if (data.step === "moments") {
            this.action = this.$gettext("Updating moments");
          } else if (data.step === "faces") {
            this.action = this.$gettext("Updating faces");
          } else if (data.step === "previews") {
            this.action = this.$gettext("Updating previews");
          } else if (data.step === "cleanup") {
            this.action = this.$gettext("Cleaning index and cache");
          } else {
            this.action = this.$gettext("Updating index");
          }

          this.busy = true;
          this.completed = 0;
          this.fileName = "";
          break;
        case "converting":
          this.action = this.$gettext("Converting");
          this.busy = true;
          this.completed = 0;
          this.fileName = data.fileName;
          break;
        case "thumbnails":
          this.action = this.$gettext("Creating thumbnails for");
          this.busy = true;
          this.completed = 0;
          this.fileName = data.fileName;
          break;
        case "completed":
          this.action = "";
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
