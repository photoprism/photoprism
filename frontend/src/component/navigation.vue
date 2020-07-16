<template>
  <div id="p-navigation">
    <template v-if="$vuetify.breakpoint.smAndDown || !auth">
    <v-toolbar dark fixed flat color="navigation darken-1" class="nav-small"
               @click.stop="showNavigation()">
      <v-toolbar-side-icon class="nav-show" v-if="auth"></v-toolbar-side-icon>

      <v-toolbar-title class="nav-title">{{ page.title }}</v-toolbar-title>

      <v-spacer></v-spacer>

      <v-avatar
              tile
              :size="28"
              class="clickable"
              @click.stop="openUpload"
              v-show="!drawer"
      >
        <img src="/static/img/logo-white.svg" alt="Logo">
      </v-avatar>
    </v-toolbar>
    <v-toolbar dark flat color="navigation darken-1">
    </v-toolbar>
    </template>
    <v-navigation-drawer
            v-if="auth"
            v-model="drawer"
            :mini-variant="mini"
            :width="270"
            :mobile-break-point="960"
            class="nav-sidebar navigation"
            fixed dark app
    >
      <v-toolbar flat>
        <v-list class="navigation-home">
          <v-list-tile class="nav-logo">
            <v-list-tile-avatar class="clickable" @click.stop.prevent="goHome">
              <img src="/static/img/logo-avatar.svg" alt="Logo">
            </v-list-tile-avatar>
            <v-list-tile-content>
              <v-list-tile-title class="title">
                PhotoPrism
              </v-list-tile-title>
            </v-list-tile-content>
            <v-list-tile-action class="hidden-sm-and-down">
              <v-btn icon @click.stop="mini = !mini" class="nav-minimize">
                <v-icon>chevron_left</v-icon>
              </v-btn>
            </v-list-tile-action>
          </v-list-tile>
        </v-list>
      </v-toolbar>

      <v-list class="pt-3">
        <v-list-tile v-if="mini" @click.stop="mini = !mini" class="nav-expand">
          <v-list-tile-action>
            <v-icon>chevron_right</v-icon>
          </v-list-tile-action>
        </v-list-tile>

        <v-list-tile v-if="mini" to="/photos" @click="" class="nav-photos">
          <v-list-tile-action>
            <v-icon>photo</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate>Photos</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!mini" prepend-icon="photo" no-action>
          <v-list-tile slot="activator" to="/photos" @click.stop="" class="nav-photos">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Photos">Photos</translate>
                <span v-if="config.count.photos > 0" class="nav-count">{{ config.count.photos }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'mono:true quality:3 photo:true' }}" :exact="true" @click="">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate>Monochrome</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'stack:true' }}" :exact="true" @click="" class="nav-stacks">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate>Stacks</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'panorama:true' }}" :exact="true" @click="" class="nav-panoramas">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate>Panoramas</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'scan:true' }}" :exact="true" @click="" class="nav-scans">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate>Scans</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/review" @click="" v-if="$config.feature('review')"
                       class="nav-review">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate>Review</translate>
                <span v-show="config.count.review > 0" class="nav-count">{{ config.count.review }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/archive" @click="" class="nav-archive" v-show="$config.feature('archive')">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate>Archive</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile to="/favorites" @click="" class="nav-favorites">
          <v-list-tile-action>
            <v-icon>favorite</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Favorites">Favorites</translate>
              <span v-show="config.count.favorites > 0" class="nav-count">{{ config.count.favorites }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile to="/private" @click="" class="nav-private" v-show="$config.feature('private')">
          <v-list-tile-action>
            <v-icon>lock</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Private">Private</translate>
              <span v-show="config.count.private > 0" class="nav-count">{{ config.count.private }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile to="/videos" @click="" class="nav-video">
          <v-list-tile-action>
            <v-icon>movie_creation</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Videos">Videos</translate>
              <span v-show="config.count.videos > 0" class="nav-count">{{ config.count.videos }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="mini" to="/albums" @click="" class="nav-albums">
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
          <v-list-tile slot="activator" to="/albums" @click.stop="" class="nav-albums">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Albums">Albums</translate>
                <span v-if="config.count.albums > 0" class="nav-count">{{ config.count.albums }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/folders" class="nav-folders">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Folders">Folders</translate>
                <span v-show="config.count.folders > 0"
                      class="nav-count">{{ config.count.folders }}</span></v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/unsorted" class="nav-unsorted">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Unsorted">Unsorted</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile :to="{ name: 'calendar' }" @click="" class="nav-calendar">
          <v-list-tile-action>
            <v-icon>date_range</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Calendar">Calendar</translate>
              <span v-show="config.count.months > 0"
                    class="nav-count">{{ config.count.months }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile :to="{ name: 'moments' }" @click="" class="nav-moments"
                     v-show="$config.feature('moments')">
          <v-list-tile-action>
            <v-icon>star</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Moments">Moments</translate>
              <span v-show="config.count.moments > 0"
                    class="nav-count">{{ config.count.moments }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="mini" :to="{ name: 'places' }" @click="" class="nav-places"
                     v-show="$config.feature('places')">
          <v-list-tile-action>
            <v-icon>place</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Places">Places</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!mini" prepend-icon="place" no-action v-show="$config.feature('places')">
          <v-list-tile slot="activator" to="/places" @click.stop="" class="nav-places">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Places">Places</translate>
                <span v-show="config.count.places > 0"
                      class="nav-count">{{ config.count.places }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/states" @click="" class="nav-states">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="States">States</translate>
                <span v-show="config.count.states > 0" class="nav-count">{{ config.count.states }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile to="/labels" @click="" class="nav-labels" v-show="$config.feature('labels')">
          <v-list-tile-action>
            <v-icon>label</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Labels">Labels</translate>
              <span v-show="config.count.labels > 0"
                    class="nav-count">{{ config.count.labels }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="mini && $config.feature('library')" to="/library" @click="" class="nav-library">
          <v-list-tile-action>
            <v-icon>camera_roll</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Library">Library</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!mini && $config.feature('library')" prepend-icon="camera_roll" no-action>
          <v-list-tile slot="activator" to="/library" @click.stop="" class="nav-library">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Library">Library</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/library/files" @click="" class="nav-originals" v-show="$config.feature('files')">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Originals">Originals</translate>
                <span v-show="config.count.files > 0" class="nav-count">{{ config.count.files }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/library/hidden" @click="" class="nav-hidden">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Hidden">Hidden</translate>
                <span v-show="config.count.hidden > 0" class="nav-count">{{ config.count.hidden }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/library/errors" @click="" class="nav-errors">
            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Errors">Errors</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <template v-if="!config.disableSettings">
          <v-list-tile v-if="mini" to="/settings" @click="" class="nav-settings">
            <v-list-tile-action>
              <v-icon>settings</v-icon>
            </v-list-tile-action>

            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Settings">Settings</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-group v-else prepend-icon="settings" no-action>
            <v-list-tile slot="activator" to="/settings" @click.stop="" class="nav-settings">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="Settings">Settings</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile :to="{ name: 'about' }" :exact="true" @click="" class="nav-about">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="Help">Help</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile  :to="{ name: 'license' }" :exact="true" @click="" class="nav-license">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="License">License</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-group>
        </template>

        <v-list-tile @click="logout" class="nav-logout" v-show="!public && auth">
          <v-list-tile-action>
            <v-icon>power_settings_new</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Logout">Logout</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile to="/login" @click="" class="nav-login" v-show="!auth">
          <v-list-tile-action>
            <v-icon>lock</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Login">Login</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile to="/help/websockets" @click="" class="nav-connecting navigation" v-show="$config.disconnected" style="position:fixed; bottom: 0; left:0; right: 0;">
          <v-list-tile-action>
            <v-icon color="warning">wifi_off</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="text--warning">
              <translate key="Offline">Offline</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>
    <p-reload-dialog :show="reload.dialog" @close="reload.dialog = false"></p-reload-dialog>
    <p-upload-dialog :show="upload.dialog" @cancel="upload.dialog = false"
                     @confirm="upload.dialog = false"></p-upload-dialog>
    <p-photo-edit-dialog :show="edit.dialog" :selection="edit.selection" :index="edit.index" :album="edit.album"
                         @close="edit.dialog = false"></p-photo-edit-dialog>
  </div>
</template>

<script>
    import Album from "model/album";
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
                reload: {
                    subscription: null,
                    dialog: false,
                },
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
        },
        methods: {
            openUpload() {
                if (this.auth && !this.readonly && this.$config.feature('upload')) {
                    this.upload.dialog = true;
                } else {
                    this.goHome();
                }
            },
            feature(name) {
                return this.$config.values.settings.features[name];
            },
            goHome() {
                if(this.$route.name !== "home") {
                    this.$router.push({name: "home"});
                }
            },
            showNavigation() {
                if (this.auth) {
                    this.drawer = true;
                    this.mini = false;
                }
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
            this.reload.subscription = Event.subscribe("dialog.reload", () => this.reload.dialog = true);
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
            Event.unsubscribe(this.reload.subscription);
            Event.unsubscribe(this.upload.subscription);
            Event.unsubscribe(this.edit.subscription);
        }
    };
</script>
