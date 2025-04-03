<template>
  <div class="p-tab p-tab-index">
    <v-form ref="form" class="p-form p-photo-index" validate-on="invalid-input" @submit.prevent="submit">
      <div class="form-header">
        <span v-if="fileName" class="text-break">{{ action }} {{ fileName }}…</span>
        <span v-else-if="action">{{ action }}…</span>
        <span v-else-if="busy">{{ $gettext(`Indexing media and sidecar files…`) }}</span>
        <span v-else-if="completed">{{ $gettext(`Done.`) }}</span>
        <span v-else>{{ $gettext(`Select the folder to be indexed…`) }}</span>
      </div>
      <div class="form-body">
        <div class="form-controls">
          <v-autocomplete
            v-model="settings.index.path"
            color="surface-variant"
            class="input-index-folder"
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
          <v-progress-linear :model-value="completed" :indeterminate="busy" :height="16" color="selected">
            <span v-if="eta" class="eta text-caption opacity-80">{{ eta }}</span>
          </v-progress-linear>
        </div>
        <div class="form-options">
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
      </div>
      <div class="form-actions">
        <div class="action-buttons">
          <v-btn
            :disabled="!busy || !ready"
            variant="flat"
            color="button"
            class="action-cancel"
            @click.stop="cancelIndexing()"
          >
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn
            :disabled="busy || !ready"
            variant="flat"
            color="highlight"
            class="action-index"
            @click.stop="startIndexing()"
          >
            {{ $gettext(`Start`) }}
            <v-icon end>mdi-update</v-icon>
          </v-btn>
        </div>
      </div>
      <div v-if="ready && !busy && config.count.hidden > 1" class="form-footer">
        <v-alert color="primary" icon="mdi-alert-circle-outline" class="v-alert--default" variant="outlined">
          <div>
            {{ $gettext(`The index currently contains %{n} hidden files.`, { n: config.count.hidden }) }}
            {{
              $gettext(
                `Their format may not be supported, they haven't been converted to JPEG yet or there are duplicates.`
              )
            }}
          </div>
        </v-alert>
      </div>
    </v-form>
  </div>
</template>

<script>
import $api from "common/api";
import Axios from "axios";
import $notify from "common/notify";
import Settings from "model/settings";
import { Folder, RootOriginals } from "model/folder";

export default {
  name: "PTabIndex",
  data() {
    const root = { path: "/", name: this.$gettext("All originals") };

    return {
      ready: !this.$config.loading(),
      settings: new Settings(this.$config.getSettings()),
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
      eta: "",
      cleanup: false,
      source: null,
      root: root,
      dirs: [root],
      rtl: this.$isRtl,
    };
  },
  created() {
    this.subscriptionId = this.$event.subscribe("index", this.handleEvent);
    this.load();
  },
  beforeUnmount() {
    this.$event.unsubscribe(this.subscriptionId);
  },
  methods: {
    load() {
      this.$config.load().then(() => {
        this.settings.setValues(this.$config.getSettings());
        this.dirs = [this.root];

        if (this.settings.index.path !== this.root.path) {
          this.dirs.push({
            path: this.settings.index.path,
            name: "/" + this.$util.truncate(this.settings.index.path, 100, "…"),
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

            this.dirs.push({ path: folders[i].Path, name: "/" + this.$util.truncate(folders[i].Path, 100, "…") });
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
      $api.delete("index");
    },
    startIndexing() {
      this.source = Axios.CancelToken.source();
      this.started = Date.now();
      this.busy = true;
      this.completed = 0;
      this.fileName = "";

      const ctx = this;
      $notify.blockUI("busy");

      // Request parameters.
      const params = {
        path: this.settings.index.path,
        rescan: this.settings.index.rescan,
        cleanup: this.cleanup,
      };

      // Submit POST request.
      $api
        .post("index", params, { cancelToken: this.source.token })
        .then(function () {
          $notify.unblockUI();
          ctx.busy = false;
          ctx.completed = 100;
          ctx.fileName = "";
        })
        .catch(function (e) {
          $notify.unblockUI();

          if (Axios.isCancel(e)) {
            // Run in background.
            return;
          }

          $notify.error(ctx.$gettext("Indexing failed"));

          ctx.busy = false;
          ctx.completed = 0;
          ctx.fileName = "";
        });
    },
    handleEvent(ev, data) {
      if (this.source) {
        this.source.cancel("run in background");
        this.source = null;
        $notify.unblockUI();
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
