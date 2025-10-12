<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="500"
    class="p-dialog p-photo-album-dialog"
    @keydown.esc.exact="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" accept-charset="UTF-8" tabindex="1" @submit.prevent="confirm">
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon icon="mdi-bookmark" size="28" color="primary"></v-icon>
          <h6 class="text-h6">{{ $gettext(`Add to album`) }}</h6>
        </v-card-title>
        <v-card-text>
          <v-combobox
            ref="input"
            v-model="selectedAlbums"
            :disabled="loading"
            :loading="loading"
            hide-details
            chips
            closable-chips
            multiple
            class="input-albums"
            :items="items"
            item-title="Title"
            item-value="UID"
            :placeholder="$gettext('Select or create albums')"
            return-object
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
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn
            :disabled="selectedAlbums.length === 0"
            variant="flat"
            color="highlight"
            class="action-confirm text-white"
            @click.stop="confirm"
          >
            {{ $gettext(`Confirm`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import Album from "model/album";
import { createAlbumSelectionWatcher } from "common/albums";

// TODO: Handle cases where users have more than 10000 albums.
const MaxResults = 10000;

export default {
  name: "PPhotoAlbumDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["close", "confirm"],
  data() {
    return {
      loading: false,
      albums: [],
      items: [],
      selectedAlbums: [],
      labels: {
        addToAlbum: this.$gettext("Add to album"),
        createAlbum: this.$gettext("Create album"),
      },
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.reset();
        this.load("");
      }
    },
    selectedAlbums: createAlbumSelectionWatcher('items'),
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
    confirm() {
      if (this.loading) {
        return;
      }

      const existingUids = [];
      const namesToCreate = [];

      (this.selectedAlbums || []).forEach((a) => {
        if (typeof a === "object" && a?.UID) {
          existingUids.push(a.UID);
        } else if (typeof a === "string" && a.length > 0) {
          namesToCreate.push(a);
        }
      });

      // Deduplicate existing UIDs
      const uniqueExistingUids = [...new Set(existingUids)];

      this.loading = true;

      if (namesToCreate.length === 0) {
        this.$emit("confirm", uniqueExistingUids);
        this.loading = false;
        return;
      }

      // Create albums in parallel and handle partial failures without closing the dialog
      const creations = namesToCreate.map((title) => ({
        title,
        promise: new Album({ Title: title, UID: "", Favorite: false }).save(),
      }));

      Promise.allSettled(creations.map((c) => c.promise))
        .then((results) => {
          const createdAlbums = [];
          const failedTitles = [];

          results.forEach((res, idx) => {
            const originalTitle = creations[idx].title;
            if (res.status === "fulfilled" && res.value && res.value.UID) {
              createdAlbums.push(res.value);
            } else {
              failedTitles.push(originalTitle);
            }
          });

          if (failedTitles.length > 0) {
            // Replace successfully created string tokens with album objects so they are not retried
            const byTitle = new Map(createdAlbums.map((a) => [a.Title || a.title || "", a]));
            this.selectedAlbums = (this.selectedAlbums || []).map((it) => {
              if (typeof it === "string") {
                const t = it.trim();
                const created = byTitle.get(t);
                return created ? created : it;
              }
              return it;
            });

            // Add created albums to the combobox items so they can be selected by object
            const known = new Set((this.items || []).map((a) => a.UID));
            createdAlbums.forEach((a) => {
              if (a && a.UID && !known.has(a.UID)) {
                this.items.push(a);
                known.add(a.UID);
              }
            });

            // Notify user and keep dialog open for corrections
            this.$notify.error(
              this.$gettext("Some albums could not be created. Please edit the names and try again.")
            );
            return; // Do not emit confirm; keep dialog open
          }

          // All created successfully → emit and let parent close the dialog
          const createdUids = createdAlbums
            .map((a) => a && a.UID)
            .filter((u) => typeof u === "string" && u.length > 0);
          this.$emit("confirm", [...uniqueExistingUids, ...createdUids]);
        })
        .finally(() => {
          this.loading = false;
        });
    },
    onLoad() {
      this.loading = true;
      this.$nextTick(() => {
        if (document.activeElement !== this.$refs.input) {
          this.$refs.input.focus();
        }
      });
    },
    onLoaded() {
      this.loading = false;
      this.$nextTick(() => {
        if (document.activeElement !== this.$refs.input) {
          this.$refs.input.focus();
        }
      });
    },
    reset() {
      this.loading = false;
      this.selectedAlbums = [];
      this.albums = [];
      this.items = [];
    },
    removeSelection(index) {
      this.selectedAlbums.splice(index, 1);
    },
    load(q) {
      if (this.loading) {
        return;
      }

      this.onLoad();

      const params = {
        q: q,
        count: MaxResults,
        offset: 0,
        type: "album",
      };

      Album.search(params)
        .then((response) => {
          this.albums = response.models;
          this.items = [...this.albums];
        })
        .finally(() => {
          this.onLoaded();
        });
    },
  },
};
</script>
