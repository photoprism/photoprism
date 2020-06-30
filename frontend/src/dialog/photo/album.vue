<template>
  <v-dialog lazy v-model="show" persistent max-width="350" class="p-photo-album-dialog" @keydown.esc="cancel">
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <v-flex xs3 text-xs-center>
            <v-icon size="54" color="secondary-dark lighten-1" v-if="newAlbum">create_new_folder</v-icon>
            <v-icon size="54" color="secondary-dark lighten-1" v-else>folder</v-icon>
          </v-flex>
          <v-flex xs9 text-xs-left align-self-center>
            <v-autocomplete
                    v-model="album"
                    browser-autocomplete="off"
                    hint="Album Name"
                    :items="items"
                    :search-input.sync="search"
                    :loading="loading"
                    hide-details
                    hide-no-data
                    item-text="Title"
                    item-value="UID"
                    :label="labels.select"
                    color="secondary-dark"
                    flat solo
                    class="input-album"
            >
            </v-autocomplete>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-3">
            <v-btn @click.stop="cancel" depressed color="secondary-light" class="action-cancel">
              <translate>Cancel</translate>
            </v-btn>
            <v-btn color="secondary-dark" depressed dark @click.stop="confirm"
                   class="action-confirm">
              <span v-if="newAlbum">{{ labels.createAlbum }}</span>
              <span v-else>{{ labels.addToAlbum }}</span>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </v-dialog>
</template>
<script>
    import Album from "model/album";

    export default {
        name: 'p-photo-album-dialog',
        props: {
            show: Boolean,
        },
        data() {
            return {
                loading: false,
                search: null,
                newAlbum: null,
                album: "",
                albums: [],
                items: [],
                labels: {
                    select: this.$gettext("Album Name"),
                    addToAlbum: this.$gettext("Add to album"),
                    createAlbum: this.$gettext("Create album"),
                }
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
                if(this.album === "" && this.newAlbum) {
                    this.loading = true;

                    this.newAlbum.save().then((a) => {
                        this.loading = false;
                        this.$emit('confirm', a.UID);
                    });
                } else {
                    this.$emit('confirm', this.album);
                }
            },
            queryServer(q) {
                if(this.loading) {
                    return;
                }

                this.loading = true;

                const params = {
                    q: q,
                    count: 1000,
                    offset: 0,
                    type: "album"
                };

                Album.search(params).then(response => {
                    this.loading = false;

                    if(response.models.length > 0 && !this.album) {
                        this.album = response.models[0].UID;
                    }

                    this.albums = response.models;
                    this.items = [...this.albums];
                }).catch(() => this.loading = false);
            },
        },
        watch: {
            search (q) {
                const exists = this.albums.findIndex((album) => album.Title === q);

                if (exists !== -1 || !q) {
                    this.items = this.albums;
                    this.newAlbum = null;
                } else {
                    this.newAlbum = new Album({Title: q, UID: "", Favorite: true});
                    this.items = this.albums.concat([this.newAlbum]);
                }
            },
            show: function (show) {
                if (show) {
                    this.queryServer("");
                }
            }
        },
    }
</script>
