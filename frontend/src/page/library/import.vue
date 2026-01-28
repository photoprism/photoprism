<template>
  <div class="p-tab p-tab-import">
    <v-form ref="form" class="p-form p-photo-import" validate-on="invalid-input" @submit.prevent="submit">
      <div class="form-header">
        <span v-if="fileName" class="text-break">{{ $gettext(`Importing %{s}…`, { s: fileName }) }}</span>
        <span v-else-if="busy">{{ $gettext(`Importing files to originals…`) }}</span>
        <span v-else-if="completed">{{ $gettext(`Done.`) }}</span>
        <span v-else-if="$config.filesQuotaReached()"
          >{{ $gettext(`Insufficient storage.`) }} {{ $gettext(`Increase storage size or delete files to continue.`) }}</span
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
          <v-combobox
            v-model="selectedAlbums"
            v-model:menu="albumsMenu"
            :disabled="busy || !ready"
            hide-details
            chips
            closable-chips
            return-object
            multiple
            class="input-albums"
            :items="albums"
            item-title="Title"
            item-value="UID"
            :placeholder="$gettext('Select or create albums')"
            @update:menu="onAlbumsMenuUpdate"
            @keydown.enter.stop="onAlbumsEnter"
          >
            <template #no-data>
              <v-list-item>
                <v-list-item-title>
                  {{ $gettext(`Press enter to create a new album.`) }}
                </v-list-item-title>
              </v-list-item>
            </template>
            <template #chip="chip">
              <v-chip
                :model-value="chip.selected"
                :disabled="chip.disabled"
                prepend-icon="mdi-bookmark"
                class="text-truncate"
                @click:close="removeSelection(chip.index)"
              >
                {{ chip.item.title ? chip.item.title : chip.item }}
              </v-chip>
            </template>
          </v-combobox>
        </div>
        <div class="form-options">
          <v-checkbox
            v-model="settings.import.move"
            :disabled="busy || !ready"
            color="surface-variant"
            density="compact"
            :label="$gettext('Move Files')"
            :hint="$gettext('Remove imported files to save storage. Unsupported file types will never be deleted, they remain in their current location.')"
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
          <v-btn :disabled="!busy || !ready" variant="flat" color="button" class="action-cancel" @click.stop="cancelImport()">
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
          <v-btn :disabled="busy || !ready || $config.filesQuotaReached()" variant="flat" color="highlight" class="action-import" @click.stop="startImport()">
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
import Album from "model/album";
import { createAlbumSelectionWatcher } from "common/albums";
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
      albums: [],
      selectedAlbums: [],
      albumsMenu: false,
      suppressAlbumsMenuOpen: false,
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.reset();

        // Set currently selected albums.
        if (this.data && Array.isArray(this.data.albums)) {
          this.selectedAlbums = this.data.albums;
        } else {
          this.selectedAlbums = [];
        }

        // Fetch albums from backend.
        this.load("");
      } else {
        this.reset();
      }
    },
    selectedAlbums: createAlbumSelectionWatcher("albums"),
  },
  reset() {
    this.albumsMenu = false;
    this.suppressAlbumsMenuOpen = false;
  },
  created() {
    this.subscriptionId = this.$event.subscribe("import", this.handleEvent);
    this.load();
  },
  beforeUnmount() {
    this.$event.unsubscribe(this.subscriptionId);
  },
  methods: {
    load(q) {
      if (this.loading) {
        return;
      }

      this.onLoad();

      const params = {
        q: q,
        count: 2000,
        offset: 0,
        type: "album",
      };

      Album.search(params)
        .then((response) => {
          this.albums = response.models;
        })
        .finally(() => {
          this.onLoaded();
        });

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
    onLoad() {
      this.loading = true;
    },
    onLoaded() {
      this.loading = false;
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

      this.onLoad();

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
        .finally(() => this.onLoaded());
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

      let addToAlbums = [];

      if (this.selectedAlbums && this.selectedAlbums.length > 0) {
        this.selectedAlbums.forEach((a) => {
          if (typeof a === "string") {
            addToAlbums.push(a);
          } else if (a instanceof Album && a.UID) {
            addToAlbums.push(a.UID);
          } else if (typeof a === "object" && a?.UID) {
            addToAlbums.push(a.UID);
          }
        });
      }

      // Deduplicate album UIDs
      addToAlbums = [...new Set(addToAlbums)];

      this.settings.import.albums = addToAlbums;

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
    onAlbumsEnter() {
      this.suppressAlbumsMenuOpen = true;
      this.albumsMenu = false;
      window.setTimeout(() => {
        this.suppressAlbumsMenuOpen = false;
      }, 250);
    },
    onAlbumsMenuUpdate(val) {
      if (val && this.suppressAlbumsMenuOpen) {
        this.albumsMenu = false;
        return;
      }
      this.albumsMenu = val;
    },
    removeSelection(index) {
      this.selectedAlbums.splice(index, 1);
    },
  },
};
</script>
