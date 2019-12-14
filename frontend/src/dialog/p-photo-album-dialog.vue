<template>
    <v-dialog v-model="show" persistent max-width="350" class="p-photo-album-dialog" @keydown.esc="cancel">
        <v-card raised elevation="24">
            <v-container fluid class="pb-2 pr-2 pl-2">
                <v-layout row wrap>
                    <v-flex xs3 text-xs-center>
                        <v-icon size="54" color="grey lighten-1">folder</v-icon>
                    </v-flex>
                    <v-flex xs9 text-xs-left align-self-center>
                        <v-autocomplete
                                v-model="album"
                                :items="albums"
                                :search-input.sync="search"
                                :loading="loading"
                                hide-details
                                item-text="AlbumName"
                                item-value="AlbumUUID"
                                :label="labels.select"
                                color="secondary-dark"
                                flat solo
                        >
                        </v-autocomplete>
                    </v-flex>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="grey lighten-3" class="p-photo-dialog-cancel">
                            <translate>Cancel</translate>
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="p-photo-dialog-confirm"><translate>Add to album</translate>
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    import Album from "../model/album";

    export default {
        name: 'p-photo-album-dialog',
        props: {
            show: Boolean,
        },
        data() {
            return {
                loading: true,
                search: null,
                album: "",
                albums: [],
                labels: {
                    select: this.$gettext("Select album"),
                }
            }
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
                this.$emit('confirm', this.album);
            },
        },
        updated() {
            if(this.albums.length > 0) {
                this.loading = false;
                return;
            }

            const params = {
                q: "",
                count: 1000,
                offset: 0,
            };

            Album.search(params).then(response => {
                if(response.models.length > 0) {
                    this.album = response.models[0].AlbumUUID;
                }

                this.albums = response.models;
                this.loading = false;
            });
        },
    }
</script>
