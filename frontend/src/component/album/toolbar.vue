<template>
  <v-form
    ref="form"
    validate-on="invalid-input"
    autocomplete="off"
    class="p-photo-toolbar p-album-toolbar"
    :class="{ 'album-navigation--cover': !compact && !searchOpen }"
    accept-charset="UTF-8"
    @submit.prevent="submitSearch"
  >
    <v-toolbar flat :style="toolbarStyle" :class="{ 'album-toolbar--cover': !compact && !searchOpen }" density="compact" class="page-toolbar">
      <v-text-field
        v-if="canSearch && searchOpen"
        ref="search"
        v-model="searchText"
        :placeholder="$gettext('Search This Album')"
        :aria-label="$gettext('Search This Album')"
        prepend-inner-icon="mdi-magnify"
        clearable
        hide-details
        single-line
        rounded="pill"
        variant="solo-filled"
        color="surface-variant"
        autocorrect="off"
        autocapitalize="none"
        density="compact"
        class="input-search background-inherit elevation-0 album-search-field"
        @click:clear="clearSearch"
        @keydown.esc.stop.prevent="closeSearch"
      ></v-text-field>
      <v-toolbar-title v-else :title="album.Title" class="page__title">
        <router-link :to="{ name: collectionRoute }" class="album-back-link" :aria-label="T(collectionTitle)">
          <v-icon>{{ $isRtl ? 'mdi-chevron-right' : 'mdi-chevron-left' }}</v-icon>
          <bdi>{{ compact ? $util.balanceBidi(album.Title) : T(collectionTitle) }}</bdi>
        </router-link>
      </v-toolbar-title>

      <v-btn v-if="canSearch && !searchOpen" icon="mdi-magnify" :title="$gettext('Search This Album')" :aria-label="$gettext('Search This Album')" class="action-album-search" @click="openSearch"></v-btn>
      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn v-bind="menuProps" icon :title="$gettext('Sort Order')" :aria-label="$gettext('Sort Order')" class="action-album-sort">
            <v-icon>mdi-sort</v-icon><v-icon v-if="activeSortOption?.reversible" size="14">{{ sortDescending(activeSortOption) ? 'mdi-arrow-down' : 'mdi-arrow-up' }}</v-icon>
          </v-btn>
        </template>
        <v-list class="album-sort-menu" density="compact">
          <v-list-item v-for="item in sorting" :key="item.value" :active="isActiveSort(item)" :class="`action-album-sort-${item.value}`" @click="selectSort(item.value)">
            <v-list-item-title>{{ item.text }}</v-list-item-title>
            <template v-if="item.reversible" #append>
              <v-icon size="small">{{ sortDescending(item) ? 'mdi-arrow-down' : 'mdi-arrow-up' }}</v-icon>
            </template>
          </v-list-item>
        </v-list>
      </v-menu>

      <v-menu>
        <template #activator="{ props: menuProps }">
          <v-btn v-bind="menuProps" :title="$gettext('Toggle View')" :aria-label="$gettext('Toggle View')" class="action-album-view" icon>
            <v-icon>{{ settings.view === 'list' ? 'mdi-view-list' : settings.view === 'mosaic' ? 'mdi-view-comfy' : 'mdi-view-column' }}</v-icon>
            <v-icon class="album-view-arrow" size="12">mdi-menu-down</v-icon>
          </v-btn>
        </template>
        <v-list density="compact">
          <v-list-item :active="settings.view === 'cards'" prepend-icon="mdi-view-column" :title="$gettext('Cards')" class="action-view-cards" @click="setView('cards')"></v-list-item>
          <v-list-item v-if="listView" :active="settings.view === 'list'" prepend-icon="mdi-view-list" :title="$gettext('List')" class="action-view-list" @click="setView('list')"></v-list-item>
          <v-list-item :active="settings.view === 'mosaic'" prepend-icon="mdi-view-comfy" :title="$gettext('Mosaic')" class="action-view-mosaic" @click="setView('mosaic')"></v-list-item>
        </v-list>
      </v-menu>

      <p-action-menu :items="menuActions" button-class="ms-1"></p-action-menu>
    </v-toolbar>

    <p-share-dialog :visible="dialog.share" :model="album" @upload="webdavUpload" @close="dialog.share = false"></p-share-dialog>
    <p-service-upload
      :visible="dialog.upload"
      :items="{ albums: album.getId() }"
      :model="album"
      @close="dialog.upload = false"
      @confirm="dialog.upload = false"
    ></p-service-upload>
    <p-album-edit-dialog :visible="dialog.edit" :album="album" @close="dialog.edit = false"></p-album-edit-dialog>
    <p-confirm-dialog
      :visible="dialog.delete"
      :text="$gettext('Are you sure you want to delete this album?')"
      :action="$gettext('Delete')"
      icon="mdi-delete-outline"
      @close="dialog.delete = false"
      @confirm="onDeleteConfirm"
    ></p-confirm-dialog>
  </v-form>
</template>
<script>
import $notify from "common/notify";
import download from "common/download";
import { T } from "common/gettext";
import $api from "common/api";
import { $session } from "app/session";
import { AlbumSortOptions, albumOrder, nextAlbumOrder } from "common/album-view";

import PActionMenu from "component/action/menu.vue";
import PConfirmDialog from "component/confirm/dialog.vue";

export default {
  name: "PAlbumToolbar",
  components: {
    PActionMenu,
    PConfirmDialog,
  },
  props: {
    compact: { type: Boolean, default: true },
    collapseProgress: { type: Number, default: 1 },
    coverStyle: { type: Object, default: () => ({}) },
    album: {
      type: Object,
      default: () => {},
    },
    filter: {
      type: Object,
      default: () => {},
    },
    updateFilter: {
      type: Function,
      default: () => {},
    },
    updateQuery: {
      type: Function,
      default: () => {},
    },
    settings: {
      type: Object,
      default: () => {},
    },
    refresh: {
      type: Function,
      default: () => {},
    },
  },
  data() {
    const settings = this.$config.getSettings();
    const features = settings.features;
    return {
      searchOpen: !!this.filter.q,
      searchText: this.filter.q || "",
      expanded: false,
      canUpload: this.$config.allow("files", "upload") && features.upload,
      canDownload: this.$config.allow("albums", "download") && features.download && !settings?.albums?.download?.disabled,
      canShare: this.$config.allow("albums", "share") && features.share,
      canManage: this.$config.allow("albums", "manage"),
      canDelete: this.$config.allow("albums", "delete"),
      experimental: this.$config.get("experimental"),
      isFullScreen: !!document.fullscreenElement,
      categories: this.$config.albumCategories(),
      collectionTitle: this.$route.meta?.collectionTitle ? this.$route.meta.collectionTitle : this.$gettext("Albums"),
      collectionRoute: this.$route.meta?.collectionRoute ? this.$route.meta.collectionRoute : "albums",
      navIcon: this.$isRtl ? "mdi-chevron-left" : "mdi-chevron-right",
      listView: this.$config.getSettings()?.search?.listView,
      dialog: {
        share: false,
        upload: false,
        edit: false,
        delete: false,
      },
    };
  },
  computed: {
    // toolbarStyle blends the cover controls into the normal PhotoPrism toolbar on scroll.
    toolbarStyle() {
      const progress = this.searchOpen ? 1 : this.collapseProgress;
      return {
        ...this.coverStyle,
        "--album-toolbar-opacity": progress,
        "--album-toolbar-color": `color-mix(in srgb, var(--album-toolbar-ink, var(--album-cover-ink, #fff)) ${100 - progress * 100}%, rgb(var(--v-theme-on-secondary)) ${progress * 100}%)`,
      };
    },
    // canSearch distinguishes signed-in album readers from share-link visitors.
    canSearch() {
      return $session.getUser()?.Role !== "visitor" && this.$config.allow("photos", "search");
    },
    // sorting translates the temporary order menu labels.
    sorting() {
      const labels = {
        newest: this.$gettext("Date Taken"), title: this.$gettext("Title"), name: this.$gettext("File Name"),
        size: this.$gettext("File Size"), duration: this.$gettext("Video Duration"), added: this.$gettext("Recently Added"),
        edited: this.$gettext("Recently Edited"), relevance: this.$gettext("Most Relevant"),
      };
      return AlbumSortOptions.map((item) => ({ ...item, text: labels[item.value] }));
    },
    // activeSortOption supplies the direction shown beside the sort button.
    activeSortOption() {
      return this.sorting.find((item) => this.isActiveSort(item));
    },
    // currentSort reads the URL override without modifying the album.
    currentSort() {
      return albumOrder(this.$route.query, this.album.Order);
    },
  },
  watch: {
    // filter.q restores the submitted text after history navigation.
    "filter.q"(value) {
      this.searchText = value || "";
      if (value) {
        this.searchOpen = true;
      }
    },
    // album.UID resets the search presentation when another album is opened.
    "album.UID"() {
      this.searchOpen = !!this.filter.q;
      this.searchText = this.filter.q || "";
    },
  },
  methods: {
    // openSearch exposes the album query field and gives it keyboard focus.
    openSearch() {
      this.searchOpen = true;
      this.$nextTick(() => this.$refs.search?.focus());
    },
    // submitSearch commits the draft without disturbing the selected order.
    submitSearch() {
      this.updateQuery({ q: this.searchText || "" });
    },
    // clearSearch removes the submitted query and keeps the field available.
    clearSearch() {
      this.searchText = "";
      this.updateQuery({ q: "" });
    },
    // closeSearch closes an empty field or restores the submitted query.
    closeSearch() {
      this.searchText = this.filter.q || "";
      this.searchOpen = !!this.filter.q;
    },
    // isActiveSort groups the two date directions under a single menu item.
    isActiveSort(item) {
      return item.value === this.currentSort.order || (item.value === "newest" && this.currentSort.order === "oldest");
    },
    // sortDescending reports the direction represented by each menu arrow.
    sortDescending(item) {
      if (!this.isActiveSort(item)) {
        return !!item.descending;
      }
      return item.value === "newest" ? this.currentSort.order === "newest" : !!item.descending !== this.currentSort.reverse;
    },
    // selectSort stores a temporary order in the URL while preserving submitted search.
    selectSort(value) {
      const next = nextAlbumOrder(value, this.currentSort);
      this.$router.replace({ query: { ...this.$route.query, order: next.order, reverse: next.reverse ? "true" : undefined } });
    },
    showExpansionPanel() {
      if (!this.expanded) {
        this.expanded = true;
      }
    },
    hideExpansionPanel() {
      if (this.expanded) {
        this.expanded = false;
      }
    },
    menuActions() {
      return [
        {
          name: "refresh",
          icon: "mdi-refresh",
          text: this.$gettext("Refresh"),
          shortcut: "Ctrl-R",
          visible: true,
          click: () => {
            this.refresh();
          },
        },
        {
          name: "edit",
          icon: "mdi-pencil",
          text: this.$gettext("Edit"),
          visible: this.canManage,
          click: () => {
            this.dialog.edit = true;
          },
        },
        {
          name: "share",
          icon: "mdi-share-variant",
          text: this.$gettext("Share"),
          class: "action-share",
          visible: this.canShare,
          click: () => {
            this.dialog.share = true;
          },
        },
        {
          name: "upload",
          icon: "mdi-cloud-upload",
          text: this.$gettext("Upload") + "…",
          shortcut: "Ctrl-U",
          visible: this.canUpload,
          click: () => {
            this.showUpload();
          },
        },
        {
          name: "download",
          icon: "mdi-download",
          text: this.$gettext("Download"),
          visible: this.canDownload,
          click: () => {
            this.download();
          },
        },
        {
          name: "delete",
          color: "danger",
          icon: "mdi-delete-outline",
          text: this.$gettext("Delete Album"),
          visible: this.canDelete && ["album", "moment", "state"].includes(this.album.Type),
          click: () => {
            this.dialog.delete = true;
          },
        },
      ];
    },
    T() {
      return T.apply(this, arguments);
    },
    webdavUpload() {
      this.dialog.share = false;
      this.dialog.upload = true;
    },
    showUpload() {
      // Pre-select manually managed albums in upload dialog.
      if (this.album.Type === "album") {
        this.$event.publish("dialog.upload", { albums: [this.album] });
      } else {
        this.$event.publish("dialog.upload", { albums: [] });
      }
    },
    onUpdate(v) {
      this.updateQuery(v);
    },
    // setView updates the presentation and URL without losing search or ordering.
    setView(name) {
      if (name) {
        if (name === "list" && !this.listView) {
          name = "mosaic";
        }

        this.refresh({ view: name });
        this.$router.replace({ query: { ...this.$route.query, view: name } });
      }
    },
    download() {
      this.onDownload(`${this.$config.apiUri}/albums/${this.album.UID}/dl?t=${this.$config.downloadToken}`);
    },
    onDownload(path) {
      $notify.success(this.$gettext("Downloading…"));

      download(path, "album.zip");
    },
    onDeleteConfirm() {
      $api.delete(`albums/${this.album.UID}`).catch(() => {
        $notify.error(this.$gettext("Unable to delete"));
      });
      this.dialog.delete = false;
    },
  },
};
</script>
