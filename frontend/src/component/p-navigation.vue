<template>
    <div id="p-navigation">
        <v-toolbar dark scroll-off-screen color="navigation darken-1" class="hidden-md-and-up p-navigation-small"
                   @click.stop="showNavigation()">
            <v-toolbar-side-icon class="p-navigation-show"></v-toolbar-side-icon>

            <v-toolbar-title class="p-navigation-title">{{ page.title }}</v-toolbar-title>

            <v-spacer></v-spacer>

            <v-toolbar-items>
                <v-btn icon @click.stop="upload.dialog = true" v-if="!readonly && $config.feature('upload')">
                    <v-icon>cloud_upload</v-icon>
                </v-btn>
            </v-toolbar-items>
        </v-toolbar>
        <v-navigation-drawer
                v-model="drawer"
                :mini-variant="mini || !auth"
                :width="270"
                :mobile-break-point="960"
                class="p-navigation-sidebar navigation"
                fixed dark app
        >
            <v-toolbar flat>
                <v-list class="navigation-home">
                    <v-list-tile class="p-navigation-logo">
                        <v-list-tile-avatar class="p-pointer" @click.stop.prevent="openDocs">
                            <img src="/static/img/logo.png" alt="Logo">
                        </v-list-tile-avatar>
                        <v-list-tile-content>
                            <v-list-tile-title class="title">
                                PhotoPrism
                            </v-list-tile-title>
                        </v-list-tile-content>
                        <v-list-tile-action class="hidden-sm-and-down">
                            <v-btn icon @click.stop="mini = !mini" class="p-navigation-minimize">
                                <v-icon>chevron_left</v-icon>
                            </v-btn>
                        </v-list-tile-action>
                    </v-list-tile>
                </v-list>
            </v-toolbar>

            <v-list class="pt-3" v-if="auth">
                <v-list-tile v-if="mini" @click.stop="mini = !mini" class="p-navigation-expand">
                    <v-list-tile-action>
                        <v-icon>chevron_right</v-icon>
                    </v-list-tile-action>
                </v-list-tile>

                <v-list-tile v-if="mini" to="/photos" @click="" class="p-navigation-photos">
                    <v-list-tile-action>
                        <v-icon>photo</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>{{ $gettext('Photos') }}</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="photo" no-action>
                    <v-list-tile slot="activator" to="/photos" @click.stop="" class="p-navigation-photos">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <span>{{ $gettext('Photos') }}</span>
                                <span v-if="config.count.photos > 0" class="p-navigation-count">{{ config.count.photos }}</span>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile :to="{name: 'photos', query: { q: 'mono:true quality:3 photo:true' }}" :exact="true" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title>{{ $gettext('Monochrome') }}</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile to="/review" @click="" v-if="$config.feature('review')">
                        <v-list-tile-content>
                            <v-list-tile-title>{{ $gettext('Review') }}</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile to="/archive" @click="" class="p-navigation-archive" v-show="$config.feature('archive')">
                        <v-list-tile-content>
                            <v-list-tile-title>{{ $gettext('Archive') }}</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile to="/favorites" @click="" class="p-navigation-favorites">
                    <v-list-tile-action>
                        <v-icon>favorite</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Favorites') }}</span>
                            <span v-show="config.count.favorites > 0" class="p-navigation-count">{{ config.count.favorites }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/private" @click="" class="p-navigation-private" v-show="$config.feature('private')" >
                    <v-list-tile-action>
                        <v-icon>lock</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Private') }}</span>
                            <span v-show="config.count.private > 0" class="p-navigation-count">{{ config.count.private }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/videos" @click="" class="p-navigation-video">
                    <v-list-tile-action>
                        <v-icon>movie_creation</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Videos') }}</span>
                            <span v-show="config.count.videos > 0" class="p-navigation-count">{{ config.count.videos }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile v-if="mini" to="/albums" @click="">
                    <v-list-tile-action>
                        <v-icon>folder</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Albums') }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="folder" no-action :append-icon="albumExpandIcon">
                    <v-list-tile slot="activator" to="/albums" @click.stop="">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <span>{{ $gettext('Albums') }}</span>
                                <span v-if="config.count.albums > 0" class="p-navigation-count">{{ config.count.albums }}</span>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile v-for="(album, index) in config.albums"
                                 :key="index"
                                 :to="{ name: 'album', params: { uuid: album.AlbumUUID, slug: album.AlbumSlug } }">
                        <v-list-tile-content>
                            <v-list-tile-title v-if="album.AlbumName">{{ album.AlbumName }}</v-list-tile-title>
                            <v-list-tile-title v-else>Untitled</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile to="/labels" @click="" class="p-navigation-labels" v-show="$config.feature('labels')">
                    <v-list-tile-action>
                        <v-icon>label</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Labels') }}</span>
                            <span v-show="config.count.labels > 0"
                                  class="p-navigation-count">{{ config.count.labels }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile :to="{ name: 'places' }" @click="" class="p-navigation-places"
                             v-show="$config.feature('places')">
                    <v-list-tile-action>
                        <v-icon>place</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Places') }}</span>
                            <span v-show="config.count.places > 0"
                                  class="p-navigation-count">{{ config.count.places }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <!-- v-list-tile to="/discover" @click="" class="p-navigation-discover" v-show="config.experimental">
                    <v-list-tile-action>
                        <v-icon>color_lens</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate>Discover</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile -->

                <!-- v-list-tile to="/events" @click="" class="p-navigation-events">
                    <v-list-tile-action>
                        <v-icon>date_range</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>Events</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile -->

                <!-- v-list-tile to="/people" @click="" class="p-navigation-people">
                    <v-list-tile-action>
                        <v-icon>people</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>People</v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile -->

                <v-list-tile to="/library" @click="" class="p-navigation-library">
                    <v-list-tile-action>
                        <v-icon>camera_roll</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Library') }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/settings" @click="" class="p-navigation-settings" v-show="!config.disableSettings">
                    <v-list-tile-action>
                        <v-icon>settings</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Settings') }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile @click="logout" class="p-navigation-logout" v-show="!public && auth">
                    <v-list-tile-action>
                        <v-icon>power_settings_new</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Logout') }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/login" @click="" class="p-navigation-login" v-show="!auth">
                    <v-list-tile-action>
                        <v-icon>lock</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <span>{{ $gettext('Login') }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

            </v-list>
        </v-navigation-drawer>
        <p-upload-dialog :show="upload.dialog" @cancel="upload.dialog = false"
                         @confirm="upload.dialog = false"></p-upload-dialog>
        <p-photo-edit-dialog :show="edit.dialog" :selection="edit.selection" :index="edit.index" :album="edit.album"
                             @close="edit.dialog = false"></p-photo-edit-dialog>
        <p-video-dialog ref="video" :play="video.play" :album="video.album"></p-video-dialog>
    </div>
</template>

<script>
    import Album from "../model/album";
    import Event from "pubsub-js";

    export default {
        name: "p-navigation",
        data() {
            return {
                drawer: null,
                mini: true,
                session: this.$session,
                public: this.$config.get("public"),
                readonly: this.$config.get("readonly"),
                config: this.$config.values,
                page: this.$config.page,
                upload: {
                    subscription: null,
                    dialog: false,
                },
                edit: {
                    subscription: null,
                    dialog: false,
                    album: null,
                    selection: [],
                    index: 0,
                },
                video: {
                    subscription: null,
                    album: null,
                    play: null,
                },
            };
        },
        computed: {
            auth() {
                return this.session.auth || this.public
            },
            albumExpandIcon() {
                if (this.config.count.albums > 0) {
                    return this.$vuetify.icons.expand
                }

                return ""
            },
        },
        methods: {
            feature(name) {
                return this.$config.values.settings.features[name];
            },
            openDocs() {
                window.open("https://docs.photoprism.org/", "_blank");
            },
            showNavigation() {
                this.drawer = true;
                this.mini = false;
            },
            createAlbum() {
                let name = "New Album";
                const album = new Album({AlbumName: name, AlbumFavorite: true});
                album.save();
            },
            logout() {
                this.$session.logout();
            },
        },
        created() {
            this.upload.subscription = Event.subscribe("dialog.upload", () => this.upload.dialog = true);

            this.edit.subscription = Event.subscribe("dialog.edit", (ev, data) => {
                if (!this.edit.dialog) {
                    this.edit.index = data.index;
                    this.edit.selection = data.selection;
                    this.edit.album = data.album;
                    this.edit.dialog = true;
                }
            });

            this.video.subscription = Event.subscribe("dialog.video", (ev, data) => {
                this.video.play = data.play;
                this.video.album = data.album;
                this.$refs.video.show = true;
            });
        },
        destroyed() {
            Event.unsubscribe(this.upload.subscription);
            Event.unsubscribe(this.edit.subscription);
            Event.unsubscribe(this.video.subscription);
        }
    };
</script>
