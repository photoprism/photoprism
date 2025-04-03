<template>
  <div class="p-tab p-tab-import">
    <v-form ref="form" class="p-form p-photo-import" validate-on="invalid-input" @submit.prevent="submit">
      <div class="form-header">
        <span v-if="fileName" class="text-break">{{ $gettext(`Importing %{s}…`, { s: fileName }) }}</span>
        <span v-else-if="busy">{{ $gettext(`Importing files to originals…`) }}</span>
        <span v-else-if="completed">{{ $gettext(`Done.`) }}</span>
        <span v-else-if="$config.filesQuotaReached()"
          >{{ $gettext(`Insufficient storage.`) }}
          {{ $gettext(`Increase storage size or delete files to continue.`) }}</span
        >
        <span v-else>{{ $gettext(`Select a source folder to import files…`) }}</span>
      </div>
      <div class="form-body">
        <div class="form-controls">
          <v-autocomplete
            v-model="settings.import.path"
            :items="dirs"
            :loading="loading"
            :disabled="busy || !ready || $config.filesQuotaReached()"
            color="surface-variant"
            class="input-import-folder"
            variant="solo-filled"
            autocomplete="off"
            item-title="name"
            item-value="path"
            hide-details
            hide-no-data
            flat
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
            v-model="settings.import.move"
            :disabled="busy || !ready"
            color="surface-variant"
            density="compact"
            :label="$gettext('Move Files')"
            :hint="
              $gettext(
                'Remove imported files to save storage. Unsupported file types will never be deleted, they remain in their current location.'
              )
            "
            prepend-icon="mdi-delete"
            persistent-hint
            @update:model-value="onChange"
          >
          </v-checkbox>
        </div>
        <div class="form-text">
          {{ $gettext(`Imported files will be sorted by date and given a unique name to avoid duplicates.`) }}
          {{ $gettext(`JPEGs and thumbnails are automatically rendered as needed.`) }}
          {{ $gettext(`Original file names will be stored and indexed.`) }}
          {{ $gettext(`Note you may manually manage your originals folder and importing is optional.`) }}
        </div>
      </div>
      <div class="form-actions">
        <div class="action-buttons">
          <v-btn
            :disabled="!busy || !ready"
            variant="flat"
            color="button"
            class="action-cancel"
            @click.stop="cancelImport()"
          >
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn
            v-if="!$config.values.readonly && $config.feature('upload')"
            :disabled="busy || !ready || $config.filesQuotaReached()"
            variant="flat"
            color="highlight"
            class="hidden-xs action-upload"
            @click.stop="showUpload()"
          >
            {{ $gettext(`Upload`) }}
            <v-icon end>mdi-cloud-upload</v-icon>
          </v-btn>
          <v-btn
            :disabled="busy || !ready || $config.filesQuotaReached()"
            variant="flat"
            color="highlight"
            class="action-import"
            @click.stop="startImport()"
          >
            {{ $gettext(`Import`) }}
            <v-icon end>mdi-plus</v-icon>
          </v-btn>
        </div>
      </div>
    </v-form>
  </div>
</template>

<script>
import $api from "common/api";
import Axios from "axios";
import $notify from "common/notify";
import Settings from "model/settings";
import { Folder, RootImport } from "model/folder";

export default {
  name: "PTabImport",
  data() {
    const root = { path: "/", name: this.$gettext("All files from import folder") };

    return {
      ready: !this.$config.loading(),
      settings: new Settings(this.$config.getSettings()),
      started: false,
      busy: false,
      loading: false,
      completed: 0,
      subscriptionId: "",
      fileName: "",
      eta: "",
      source: null,
      root: root,
      dirs: [root],
      rtl: this.$isRtl,
    };
  },
  created() {
    this.subscriptionId = this.$event.subscribe("import", this.handleEvent);
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

        if (this.settings.import.path !== this.root.path) {
          this.dirs.push({
            path: this.settings.import.path,
            name: "/" + this.$util.truncate(this.settings.import.path, 100, "…"),
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

            this.dirs.push({ path: folders[i].Path, name: "/" + this.$util.truncate(folders[i].Path, 100, "…") });
          }

          if (!found) {
            this.settings.import.path = this.root.path;
          }
        })
        .finally(() => (this.loading = false));
    },
    showUpload() {
      this.$event.publish("dialog.upload");
    },
    submit() {
      // DO NOTHING
    },
    cancelImport() {
      $api.delete("import");
    },
    startImport() {
      this.source = Axios.CancelToken.source();
      this.started = Date.now();
      this.busy = true;
      this.completed = 0;
      this.fileName = "";

      const ctx = this;
      $notify.blockUI("busy");

      $api
        .post("import", this.settings.import, { cancelToken: this.source.token })
        .then(function () {
          $notify.unblockUI();
          ctx.busy = false;
          ctx.completed = 100;
          ctx.fileName = "";
        })
        .catch(function (e) {
          $notify.unblockUI();

          if (Axios.isCancel(e)) {
            // run in background
            return;
          }

          $notify.error(ctx.$gettext("Import failed"));

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
