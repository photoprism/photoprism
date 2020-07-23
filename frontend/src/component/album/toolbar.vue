<template>
  <v-form lazy-validation dense
          ref="form" autocomplete="off" class="p-photo-toolbar p-album-toolbar" accept-charset="UTF-8"
          @submit.prevent="filterChange">
    <v-toolbar flat color="secondary">
      <v-toolbar-title :title="album.Title">
        {{ album.Title }}
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon @click.stop="refresh" class="hidden-xs-only action-reload">
        <v-icon>refresh</v-icon>
      </v-btn>

      <v-btn icon @click.stop="dialog.edit = true" class="action-edit">
        <v-icon>edit</v-icon>
      </v-btn>

      <v-btn icon @click.stop="dialog.share = true" v-if="$config.feature('share')" class="action-share">
        <v-icon>share</v-icon>
      </v-btn>

      <v-btn icon v-if="settings.view === 'cards'" @click.stop="setView('list')">
        <v-icon>view_list</v-icon>
      </v-btn>
      <v-btn icon v-else-if="settings.view === 'list'" @click.stop="setView('mosaic')">
        <v-icon>view_comfy</v-icon>
      </v-btn>
      <v-btn icon v-else @click.stop="setView('cards')">
        <v-icon>view_column</v-icon>
      </v-btn>

      <v-btn icon @click.stop="showUpload()" v-if="!$config.values.readonly && $config.feature('upload')"
             class="hidden-sm-and-down action-upload">
        <v-icon>cloud_upload</v-icon>
      </v-btn>
    </v-toolbar>

    <template v-if="album.Description">
      <v-card flat class="px-2 py-1 hidden-sm-and-down"
              color="secondary-light"
      >
        <v-card-text>
          {{ album.Description }}
        </v-card-text>
      </v-card>
      <v-card flat class="pa-0 hidden-md-and-up"
              color="secondary-light"
      >
        <v-card-text>
          {{ album.Description }}
        </v-card-text>
      </v-card>
    </template>

    <p-share-dialog :show="dialog.share" :model="album" @upload="webdavUpload"
                    @close="dialog.share = false"></p-share-dialog>
    <p-share-upload-dialog :show="dialog.upload" :selection="[album.getId()]" @cancel="dialog.upload = false"
                           @confirm="dialog.upload = false"></p-share-upload-dialog>
    <p-album-edit-dialog :show="dialog.edit" :album="album" @close="dialog.edit = false"></p-album-edit-dialog>
  </v-form>
</template>
<script>
    import Event from "pubsub-js";

    export default {
        name: 'p-album-toolbar',
        props: {
            album: Object,
            filter: Object,
            settings: Object,
            refresh: Function,
            filterChange: Function,
        },
        data() {
            const cameras = [{
                ID: 0,
                Name: this.$gettext('All Cameras')
            }].concat(this.$config.get('cameras'));
            const countries = [{
                ID: '',
                Name: this.$gettext('All Countries')
            }].concat(this.$config.get('countries'));

            return {
                experimental: this.$config.get("experimental"),
                isFullScreen: !!document.fullscreenElement,
                categories: this.$config.albumCategories(),
                searchExpanded: false,
                options: {
                    'views': [
                        {value: 'mosaic', text: this.$gettext('Mosaic')},
                        {value: 'cards', text: this.$gettext('Cards')},
                        {value: 'list', text: this.$gettext('List')},
                    ],
                    'countries': countries,
                    'cameras': cameras,
                    'sorting': [
                        {value: 'added', text: this.$gettext('Recently added')},
                        {value: 'edited', text: this.$gettext('Recently edited')},
                        {value: 'newest', text: this.$gettext('Newest first')},
                        {value: 'oldest', text: this.$gettext('Oldest first')},
                        {value: 'name', text: this.$gettext('Sort by file name')},
                        {value: 'similar', text: this.$gettext('Group by similarity')},
                        {value: 'relevance', text: this.$gettext('Most relevant')},
                    ],
                },
                dialog: {
                    share: false,
                    upload: false,
                    edit: false,
                },
                labels: {
                    title: this.$gettext("Album Name"),
                    description: this.$gettext("Description"),
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    sort: this.$gettext("Sort Order"),
                    category: this.$gettext("Category"),
                },
                titleRule: v => v.length <= this.$config.get('clip') || this.$gettext("Name too long"),
                growDesc: false,
            };
        },
        methods: {
            webdavUpload() {
                this.dialog.share = false;
                this.dialog.upload = true;
            },
            showUpload() {
                Event.publish("dialog.upload");
            },
            expand() {
                this.searchExpanded = !this.searchExpanded;
                this.growDesc = !this.growDesc;
            },
            updateAlbum() {
                if (this.album.wasChanged()) {
                    this.album.update();
                }
            },
            dropdownChange() {
                this.filterChange();

                if (window.innerWidth < 600) {
                    this.searchExpanded = false;
                }

                if (this.filter.order !== this.album.Order) {
                    this.album.Order = this.filter.order;
                    this.updateAlbum()
                }
            },
            setView(name) {
                this.settings.view = name;
                this.filterChange();
            },
            clearQuery() {
                this.filter.q = '';
                this.filterChange();
            },
        }
    };
</script>
