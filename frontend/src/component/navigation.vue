<template>
  <div id="p-navigation">
    <template v-if="$vuetify.breakpoint.smAndDown || !auth">
      <v-toolbar dark fixed flat scroll-off-screen :dense="$vuetify.breakpoint.smAndDown"  color="navigation darken-1" class="nav-small"
                  @click.stop="showNavigation()">

        <v-toolbar-side-icon v-if="auth" class="nav-show"></v-toolbar-side-icon>

        <v-toolbar-title class="nav-title">{{ page.title }}</v-toolbar-title>

        <v-spacer></v-spacer>

        <v-btn v-if="auth && !config.readonly && $config.feature('upload')" icon class="action-upload"
               :title="$gettext('Upload')" @click.stop="openUpload()">
          <v-icon>cloud_upload</v-icon>
        </v-btn>

      </v-toolbar>
      <v-toolbar dark flat :dense="$vuetify.breakpoint.smAndDown" color="#fafafa">
      </v-toolbar>
    </template>
    <v-navigation-drawer
        v-if="auth"
        v-model="drawer"
        :mini-variant="isMini"
        :width="270"
        :mobile-break-point="960"
        :mini-variant-width="80"
        class="nav-sidebar navigation p-flex-nav"
        fixed dark app
        :right="rtl"
    >
      <v-toolbar flat :dense="$vuetify.breakpoint.smAndDown">
        <v-list class="navigation-home">
          <v-list-tile class="nav-logo">
            <v-list-tile-avatar class="clickable" @click.stop.prevent="goHome">
              <img :src="$config.staticUri + '/img/logo-avatar.svg'" alt="Logo">
            </v-list-tile-avatar>
            <v-list-tile-content>
              <v-list-tile-title class="title">
                PhotoPrism
              </v-list-tile-title>
            </v-list-tile-content>
            <v-list-tile-action class="hidden-sm-and-down" :title="$gettext('Minimize')">
              <v-btn icon class="nav-minimize" @click.stop="toggleIsMini()">
                <v-icon v-if="!rtl">chevron_left</v-icon>
                <v-icon v-else>chevron_right</v-icon>
              </v-btn>
            </v-list-tile-action>
          </v-list-tile>
        </v-list>
      </v-toolbar>

      <v-list class="pt-3 p-flex-menu">
        <v-list-tile v-if="isMini" class="nav-expand" @click.stop="toggleIsMini()">
          <v-list-tile-action :title="$gettext('Expand')">
            <v-icon v-if="!rtl">chevron_right</v-icon>
            <v-icon v-else>chevron_left</v-icon>
          </v-list-tile-action>
        </v-list-tile>

        <v-list-tile v-if="isMini" to="/browse" class="nav-browse" @click.stop="">
          <v-list-tile-action :title="$gettext('Search')">
            <v-icon>search</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate>Search</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini" prepend-icon="search" no-action>
          <template #activator>
            <v-list-tile to="/browse" class="nav-browse" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="Search">Search</translate>
                  <span v-if="config.count.all > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.all }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile :to="{name: 'browse', query: { q: 'mono:true quality:3 photo:true' }}" :exact="true" class="nav-monochrome" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Monochrome</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'browse', query: { q: 'panorama:true' }}" :exact="true" class="nav-panoramas"
                       @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Panoramas</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'stack:true' }}" :exact="true" class="nav-stacks" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Stacks</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile :to="{name: 'photos', query: { q: 'scan:true' }}" :exact="true" class="nav-scans" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Scans</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile v-if="$config.feature('review') && hasPermission(aclResources.ResourceReview, aclActions.ActionRead)" to="/review" class="nav-review"
                       @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Review</translate>
                <span v-show="config.count.review > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.review }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile v-show="$config.feature('archive') && hasPermission(aclResources.ResourceArchive, aclActions.ActionRead)" to="/archive" class="nav-archive" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate>Archive</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile v-if="isMini && $config.feature('albums')" to="/albums" class="nav-albums" @click.stop="">
          <v-list-tile-action :title="$gettext('Albums')">
            <v-icon>bookmark</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
              <translate key="Albums">Albums</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini && $config.feature('albums')" prepend-icon="bookmark" no-action>
          <template #activator>
            <v-list-tile to="/albums" class="nav-albums" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="Albums">Albums</translate>
                  <span v-if="config.count.albums > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.albums }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile to="/unsorted" class="nav-unsorted">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Unsorted">Unsorted</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile v-if="$config.feature('videos')" to="/videos" class="nav-video" @click.stop="">
          <v-list-tile-action :title="$gettext('Videos')">
            <v-icon>play_circle_fill</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Videos">Videos</translate>
              <span v-show="config.count.videos > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.videos }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('people')" :to="{ name: 'people' }" class="nav-people" @click.stop="">
          <v-list-tile-action :title="$gettext('People')">
            <v-icon>person</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="People">People</translate>
              <span v-show="config.count.people > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.people }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile to="/favorites" class="nav-favorites" @click.stop="">
          <v-list-tile-action :title="$gettext('Favorites')">
            <v-icon>favorite</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Favorites">Favorites</translate>
              <span v-show="config.count.favorites > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.favorites }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('moments')" :to="{ name: 'moments' }" class="nav-moments"
                     @click.stop="">
          <v-list-tile-action :title="$gettext('Moments')">
            <v-icon>star</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Moments">Moments</translate>
              <span v-show="config.count.moments > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.moments }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile :to="{ name: 'calendar' }" class="nav-calendar" @click.stop="">
          <v-list-tile-action :title="$gettext('Calendar')">
            <v-icon>date_range</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Calendar">Calendar</translate>
              <span v-show="config.count.months > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.months }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="isMini" v-show="$config.feature('places')" :to="{ name: 'places' }" class="nav-places"
                     @click.stop="">
          <v-list-tile-action :title="$gettext('Places')">
            <v-icon>place</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Places">Places</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini" v-show="$config.feature('places')" prepend-icon="place" no-action>
          <template #activator>
            <v-list-tile to="/places" class="nav-places" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="Places">Places</translate>
                  <span v-show="config.count.places > 0"
                        :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.places }}</span>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile to="/states" class="nav-states" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="States">States</translate>
                <span v-show="config.count.states > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.states }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <v-list-tile v-show="$config.feature('labels')" to="/labels" class="nav-labels" @click.stop="">
          <v-list-tile-action :title="$gettext('Labels')">
            <v-icon>label</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Labels">Labels</translate>
              <span v-show="config.count.labels > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.labels }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('folders')" :to="{ name: 'folders' }" class="nav-folders" @click.stop="">
          <v-list-tile-action :title="$gettext('Folders')">
            <v-icon>folder</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Folders">Folders</translate>
              <span v-show="config.count.folders > 0"
                    :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.folders }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="$config.feature('private') && hasPermission(aclResources.ResourcePrivate, aclActions.ActionRead)" to="/private" class="nav-private" @click.stop="">
          <v-list-tile-action :title="$gettext('Private')">
            <v-icon>lock</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Private">Private</translate>
              <span v-show="config.count.private > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.private }}</span>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-if="isMini && $config.feature('library') && hasPermission(aclResources.ResourceLibrary, aclActions.ActionRead)" to="/library" class="nav-library" @click.stop="">
          <v-list-tile-action :title="$gettext('Library')">
            <v-icon>camera_roll</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Library">Library</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-group v-if="!isMini && $config.feature('library') && hasPermission(aclResources.ResourceLibrary, aclActions.ActionRead)" prepend-icon="camera_roll" no-action>
          <template #activator>
            <v-list-tile to="/library" class="nav-library" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title>
                  <translate key="Library">Library</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </template>

          <v-list-tile v-show="$config.feature('files')" to="/library/files" class="nav-originals" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Originals">Originals</translate>
                <span v-show="config.count.files > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.files }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/library/hidden" class="nav-hidden" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Hidden">Hidden</translate>
                <span v-show="config.count.hidden > 0" :class="`nav-count ${rtl ? '--rtl' : ''}`">{{ config.count.hidden }}</span>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-tile to="/library/errors" class="nav-errors" @click.stop="">
            <v-list-tile-content>
              <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                <translate key="Errors">Errors</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
        </v-list-group>

        <template v-if="!config.disable.settings && hasPermission(aclResources.ResourceSettings, aclActions.ActionRead)">
          <v-list-tile v-if="isMini" to="/settings" class="nav-settings" @click.stop="">
            <v-list-tile-action :title="$gettext('Settings')">
              <v-icon>settings</v-icon>
            </v-list-tile-action>

            <v-list-tile-content>
              <v-list-tile-title>
                <translate key="Settings">Settings</translate>
              </v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>

          <v-list-group v-else prepend-icon="settings" no-action>
            <template #activator>
              <v-list-tile to="/settings" class="nav-settings" @click.stop="">
                <v-list-tile-content>
                  <v-list-tile-title>
                    <translate key="Settings">Settings</translate>
                  </v-list-tile-title>
                </v-list-tile-content>
              </v-list-tile>
            </template>

            <v-list-tile :to="{ name: 'about' }" :exact="true" class="nav-about" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>About</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile v-show="!isPublic && auth" :to="{ name: 'feedback' }" :exact="true" class="nav-feedback"
                         @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate>Feedback</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>

            <v-list-tile :to="{ name: 'license' }" :exact="true" class="nav-license" @click.stop="">
              <v-list-tile-content>
                <v-list-tile-title :class="`menu-item ${rtl ? '--rtl' : ''}`">
                  <translate key="License">License</translate>
                </v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-group>
        </template>

        <v-list-tile v-show="!auth" to="/login" class="nav-login" @click.stop="">
          <v-list-tile-action :title="$gettext('Login')">
            <v-icon>lock</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Login">Login</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

      </v-list>

      <v-list class="p-user-box">

        <v-list-tile v-show="$config.disconnected" to="/help/websockets" class="nav-connecting navigation" @click.stop="">
          <v-list-tile-action :title="$gettext('Offline')">
            <v-icon color="warning">wifi_off</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title class="text--warning">
              <translate key="Offline">Offline</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile v-show="!isPublic && auth"  to="/account" class="p-profile">
          <v-list-tile-avatar color="grey" size="36">
            <span class="white--text headline">{{ displayName.length >= 1 ? displayName[0].toUpperCase() : "E" }}</span>
          </v-list-tile-avatar>

          <v-list-tile-content>
            <v-list-tile-title>
              {{ displayName }}
            </v-list-tile-title>
            <v-list-tile-sub-title style="font-size: smaller">{{ accountInfo }}</v-list-tile-sub-title>
          </v-list-tile-content>

          <v-list-tile-action :title="$gettext('Logout')">
            <v-btn icon @click="logout">
              <v-icon>power_settings_new</v-icon>
            </v-btn>
          </v-list-tile-action>
        </v-list-tile>

        <v-list-tile v-show="!isPublic && auth && isMini" class="nav-logout" @click="logout">
          <v-list-tile-action :title="$gettext('Logout')">
            <v-icon>power_settings_new</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>
              <translate key="Logout">Logout</translate>
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>

    </v-navigation-drawer>
    <div v-if="isTest" id="photoprism-info"><a href="https://photoprism.app/" target="_blank">Browse Your Life in Pictures</a></div>
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
  name: "PNavigation",
  data() {
    return {
      drawer: null,
      isMini: localStorage.getItem('last_navigation_mode') !== 'false',
      isPublic: this.$config.get("public"),
      isTest: this.$config.test,
      isReadOnly: this.$config.get("readonly"),
      session: this.$session,
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
      rtl: this.$rtl,
    };
  },
  computed: {
    auth() {
      return this.session.auth || this.isPublic;
    },
    displayName() {
      const user = this.$session.getUser();
      return user.FullName ? user.FullName : user.UserName;
    },
    accountInfo() {
      const user = this.$session.getUser();
      return user.PrimaryEmail ? user.PrimaryEmail : this.$gettext("Account");
    }
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
  },
  methods: {
    openUpload() {
      if (this.auth && !this.isReadOnly && this.$config.feature('upload')) {
        this.upload.dialog = true;
      } else {
        this.goHome();
      }
    },
    goHome() {
      if (this.$route.name !== "home") {
        this.$router.push({name: "home"});
      }
    },
    showNavigation() {
      if (this.auth) {
        this.drawer = true;
        this.isMini = false;
      }
    },
    createAlbum() {
      let name = "New Album";
      const album = new Album({Title: name, Favorite: false});
      album.save();
    },
    toggleIsMini() {
      this.isMini = !this.isMini;
      localStorage.setItem('last_navigation_mode', `${this.isMini}`);
    },
    logout() {
      this.$session.logout();
    },
  }
};
</script>
