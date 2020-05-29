<template>
    <div id="p-navigation">
        <v-toolbar dark scroll-off-screen color="navigation darken-1" class="hidden-md-and-up p-navigation-small"
                   @click.stop="showNavigation()">
            <v-toolbar-side-icon class="p-navigation-show"></v-toolbar-side-icon>

            <v-toolbar-title class="p-navigation-title">{{ page.title }}</v-toolbar-title>

            <v-spacer></v-spacer>

            <v-toolbar-items>
                <v-btn icon class="action-upload" @click.stop="upload.dialog = true" v-if="!readonly && $config.feature('upload')">
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
                        <v-list-tile-avatar class="clickable" @click.stop.prevent="openDocs">
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
                        <v-list-tile-title><translate key="Photos">Photos</translate></v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="photo" no-action>
                    <v-list-tile slot="activator" to="/photos" @click.stop="" class="p-navigation-photos">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <translate key="Photos">Photos</translate>
                                <span v-if="config.count.photos > 0" class="p-navigation-count">{{ config.count.photos }}</span>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile :to="{name: 'photos', query: { q: 'mono:true quality:3 photo:true' }}" :exact="true" @click="">
                        <v-list-tile-content>
                            <v-list-tile-title><translate key="Monochrome">Monochrome</translate></v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile to="/review" @click="" v-if="$config.feature('review') && config.count.review > 0" class="p-navigation-review">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <translate key="Review">Review</translate>
                                <span v-show="config.count.review > 0" class="p-navigation-count">{{ config.count.review }}</span>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile to="/archive" @click="" class="p-navigation-archive" v-show="$config.feature('archive')">
                        <v-list-tile-content>
                            <v-list-tile-title><translate key="Archive">Archive</translate></v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile to="/favorites" @click="" class="p-navigation-favorites">
                    <v-list-tile-action>
                        <v-icon>favorite</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Favorites">Favorites</translate>
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
                            <translate key="Private">Private</translate>
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
                            <translate key="Videos">Videos</translate>
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
                            <translate key="Albums">Albums</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="folder" no-action>
                    <v-list-tile slot="activator" to="/albums" @click.stop="">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <translate key="Albums">Albums</translate>
                                <span v-if="config.count.albums > 0" class="p-navigation-count">{{ config.count.albums }}</span>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile>
                        <v-list-tile-content>
                            <v-list-tile-title><translate key="Folders">Folders</translate>
                                <span v-show="config.count.folders > 0"
                                      class="p-navigation-count">{{ config.count.folders }}</span></v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile>
                        <v-list-tile-content>
                            <v-list-tile-title><translate key="Months">Months</translate>
                                <span v-show="config.count.months > 0"
                                      class="p-navigation-count">{{ config.count.months }}</span></v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile v-for="(album, index) in config.albums"
                                 :key="index"
                                 :to="{ name: 'album', params: { uid: album.UID, slug: album.Slug } }">
                        <v-list-tile-content>
                            <v-list-tile-title v-if="album.Title">{{ album.Title }}</v-list-tile-title>
                            <v-list-tile-title v-else>Untitled</v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile :to="{ name: 'moments' }" @click="" class="p-navigation-moments"
                             v-show="config.experimental && $config.feature('moments')">
                    <v-list-tile-action>
                        <v-icon>star</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Moments">Moments</translate>
                            <span v-show="config.count.moments > 0"
                                  class="p-navigation-count">{{ config.count.moments }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <!-- v-list-tile to="/events" @click="" class="p-navigation-events">
                                    <v-list-tile-action>
                                        <v-icon>date_range</v-icon>
                                    </v-list-tile-action>

                                    <v-list-tile-content>
                                        <v-list-tile-title>Events</v-list-tile-title>
                                    </v-list-tile-content>
                                </v-list-tile -->

                <v-list-tile :to="{ name: 'places' }" @click="" class="p-navigation-places"
                             v-show="$config.feature('places')">
                    <v-list-tile-action>
                        <v-icon>place</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Places">Places</translate>
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


                <!-- v-list-tile to="/folders" @click="" class="p-navigation-folders">
                    <v-list-tile-action>
                        <v-icon>sd_storage</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Folders">Folders</translate>
                            <span v-show="config.count.folders > 0" class="p-navigation-count">{{ config.count.folders }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile -->

                <v-list-tile to="/labels" @click="" class="p-navigation-labels" v-show="$config.feature('labels')">
                    <v-list-tile-action>
                        <v-icon>label</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Labels">Labels</translate>
                            <span v-show="config.count.labels > 0"
                                  class="p-navigation-count">{{ config.count.labels }}</span>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile v-if="mini" to="/library" @click="" class="p-navigation-library">
                    <v-list-tile-action>
                        <v-icon>camera_roll</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Library">Library</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-group v-if="!mini" prepend-icon="camera_roll" no-action>
                    <v-list-tile slot="activator" to="/library" @click.stop="" class="p-navigation-library">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <translate key="Library">Library</translate>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>

                    <v-list-tile to="/files" @click="" class="p-navigation-files" v-show="$config.feature('files')">
                        <v-list-tile-content>
                            <v-list-tile-title>
                                <translate key="Files">Files</translate>
                                <span v-show="config.count.files > 0" class="p-navigation-count">{{ config.count.files }}</span>
                            </v-list-tile-title>
                        </v-list-tile-content>
                    </v-list-tile>
                </v-list-group>

                <v-list-tile to="/settings" @click="" class="p-navigation-settings" v-show="!config.disableSettings">
                    <v-list-tile-action>
                        <v-icon>settings</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Settings">Settings</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile @click="logout" class="p-navigation-logout" v-show="!public && auth">
                    <v-list-tile-action>
                        <v-icon>power_settings_new</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Logout">Logout</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

                <v-list-tile to="/login" @click="" class="p-navigation-login" v-show="!auth">
                    <v-list-tile-action>
                        <v-icon>lock</v-icon>
                    </v-list-tile-action>

                    <v-list-tile-content>
                        <v-list-tile-title>
                            <translate key="Login">Login</translate>
                        </v-list-tile-title>
                    </v-list-tile-content>
                </v-list-tile>

            </v-list>
        </v-navigation-drawer>
        <p-upload-dialog :show="upload.dialog" @cancel="upload.dialog = false"
                         @confirm="upload.dialog = false"></p-upload-dialog>
        <p-photo-edit-dialog :show="edit.dialog" :selection="edit.selection" :index="edit.index" :album="edit.album"
                             @close="edit.dialog = false"></p-photo-edit-dialog>
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
                const album = new Album({Title: name, Favorite: true});
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
        },
        destroyed() {
            Event.unsubscribe(this.upload.subscription);
            Event.unsubscribe(this.edit.subscription);
        }
    };
</script>
