<template>
  <div class="p-tab p-tab-index">
    <v-form ref="form" class="p-photo-index" lazy-validation dense @submit.prevent="submit">
      <v-container fluid>
        <p class="subheading">
          <span v-if="fileName" class="break-word">{{ action }} {{ fileName }}…</span>
          <span v-else-if="action">{{ action }}…</span>
          <span v-else-if="busy"><translate>Indexing media and sidecar files…</translate></span>
          <span v-else-if="completed"><translate>Done.</translate></span>
          <span v-else><translate>Press button to start indexing…</translate></span>
        </p>

        <v-autocomplete
          v-model="settings.index.path"
          color="secondary-dark"
          class="my-3 input-index-folder"
          hide-details
          hide-no-data flat solo browser-autocomplete="off"
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
          <v-progress-linear color="secondary-dark" height="1.5em" :value="completed"
                             :indeterminate="busy"></v-progress-linear>
        </p>

        <v-layout wrap align-top class="pb-3">
          <v-flex xs12 sm6 lg3 xl2 class="px-2 pb-2 pt-2">
            <v-checkbox
              v-model="settings.index.rescan"
              :disabled="busy || !ready"
              class="ma-0 pa-0"
              color="secondary-dark"
              :label="$gettext('Complete Rescan')"
              :hint="$gettext('Re-index all originals, including already indexed and unchanged files.')"
              prepend-icon="cached"
              persistent-hint
              @change="onChange"
            >
            </v-checkbox>
          </v-flex>
          <v-flex v-if="isAdmin" xs12 sm6 lg3 xl2 class="px-2 pb-2 pt-2">
            <v-checkbox
              v-model="cleanup"
              :disabled="busy || !ready"
              class="ma-0 pa-0"
              color="secondary-dark"
              :label="$gettext('Cleanup')"
              :hint="$gettext('Delete orphaned index entries, sidecar files and thumbnails.')"
              prepend-icon="delete_sweep"
              persistent-hint
            >
            </v-checkbox>
          </v-flex>
        </v-layout>

        <v-btn
          :disabled="!busy || !ready"
          color="primary-button"
          class="white--text ml-0 mt-2 action-cancel"
          depressed
          @click.stop="cancelIndexing()"
        >
          <translate>Cancel</translate>
        </v-btn>

        <v-btn
          :disabled="busy || !ready"
          color="primary-button"
          class="white--text ml-0 mt-2 action-index"
          depressed
          @click.stop="startIndexing()"
        >
          <translate>Start</translate>
          <v-icon :right="!rtl" :left="rtl" dark>update</v-icon>
        </v-btn>

        <v-alert
          v-if="ready && !busy && config.count.hidden > 1"
          :value="true"
          color="error"
          icon="priority_high"
          class="mt-3"
          outline
        >
          <translate :translate-params="{n: config.count.hidden}">The index currently contains %{n} hidden files.</translate>
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
import {Folder, RootOriginals} from "model/folder";

export default {
  name: 'PTabIndex',
  data() {
    const root = {"path": "/", "name": this.$gettext("All originals")};

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
    this.subscriptionId = Event.subscribe('index', this.handleEvent);
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

        if (this.settings.index.path !== this.root.path) {
          this.dirs.push({
            path: this.settings.index.path,
            name: "/" + Util.truncate(this.settings.index.path, 100, "…")
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

      Folder.findAll(RootOriginals).then((r) => {
        const folders = r.models ? r.models : [];
        const currentPath = this.settings.index.path;
        let found = currentPath === this.root.path;

        this.dirs = [this.root];

        for (let i = 0; i < folders.length; i++) {
          if (currentPath === folders[i].Path) {
            found = true;
          }

          this.dirs.push({path: folders[i].Path, name: "/" + Util.truncate(folders[i].Path, 100, "…")});
        }

        if (!found) {
          this.settings.index.path = this.root.path;
        }
      }).finally(() => this.loading = false);
    },
    submit() {
      // DO NOTHING
    },
    cancelIndexing() {
      Api.delete('index');
    },
    startIndexing() {
      this.source = Axios.CancelToken.source();
      this.started = Date.now();
      this.busy = true;
      this.completed = 0;
      this.fileName = '';

      const ctx = this;
      Notify.blockUI();

      // Request parameters.
      const params = {
        path: this.settings.index.path,
        rescan: this.settings.index.rescan,
        cleanup: this.cleanup,
      };

      // Submit POST request.
      Api.post('index', params, {cancelToken: this.source.token}).then(function () {
        Notify.unblockUI();
        ctx.busy = false;
        ctx.completed = 100;
        ctx.fileName = '';
      }).catch(function (e) {
        Notify.unblockUI();

        if (Axios.isCancel(e)) {
          // Run in background.
          return;
        }

        Notify.error(ctx.$gettext("Indexing failed"));

        ctx.busy = false;
        ctx.completed = 0;
        ctx.fileName = '';
      });
    },
    handleEvent(ev, data) {
      if (this.source) {
        this.source.cancel('run in background');
        this.source = null;
        Notify.unblockUI();
      }

      const type = ev.split('.')[1];

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
        case 'completed':
          this.action = "";
          this.busy = false;
          this.completed = 100;
          this.fileName = '';
          break;
        default:
          console.log(data);
      }
    },
  },
};
</script>
