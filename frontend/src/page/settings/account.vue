<template>
  <div class="p-tab p-settings-account">
    <v-form ref="form" lazy-validation
            dense class="p-form-account" accept-charset="UTF-8"
            @submit.prevent="onChange">
      <input ref="upload" type="file" class="d-none input-upload" @change.stop="onUploadAvatar()">
      <v-card flat tile class="mt-2 px-1 application">
        <v-card-actions>
          <v-layout row wrap align-top>
            <v-flex
                class="p-photo pa-3 text-xs-center"
                xs4 sm3 md2 xl1 fill-height
            >
              <div class="user-avatar" @click.exact="onChangeAvatar()">
                <v-img :src="user.getAvatarURL()"
                       :alt="displayName"  aspect-ratio="1"
                       class="grey lighten-1 elevation-0 clickable"
                ></v-img>
              </div>
            </v-flex>
            <v-flex xs8 sm9 md10 xl11 fill-height class="pa-0">
              <v-layout wrap align-top>
                <v-flex xs12 md3 class="pa-2">
                  <v-text-field
                      v-model="user.Name"
                      hide-details required box flat readonly
                      browser-autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Username')"
                      class="input-name"
                      color="secondary-dark"
                  ></v-text-field>
                </v-flex>
                <v-flex xs12 md9 class="pa-2">
                  <v-text-field
                      v-model="user.DisplayName"
                      hide-details required box flat
                      :disabled="busy"
                      browser-autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Display Name')"
                      class="input-display-name"
                      color="secondary-dark"
                      @change="onChange"
                  ></v-text-field>
                </v-flex>
                <v-flex xs12 class="pa-2">
                  <v-text-field
                      v-model="user.Email"
                      hide-details required box flat
                      type="email"
                      :disabled="busy"
                      browser-autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Email')"
                      class="input-email"
                      color="secondary-dark"
                      @change="onChange"
                  ></v-text-field>
                </v-flex>
              </v-layout>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-1">
          <h3 class="body-2 mb-0">
            <translate>Security and Access</translate>
          </h3>
        </v-card-title>
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="pa-2">
              <v-btn block depressed color="secondary-light" class="action-change-password compact" :disabled="isPublic || isDemo || user.Name === ''"
                     @click.stop="showDialog('password')">
                <translate>Change Password</translate>
                <v-icon :right="!rtl" :left="rtl" dark>lock</v-icon>
              </v-btn>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-btn block depressed color="secondary-light" class="action-webdav-dialog compact"
                     :disabled="isPublic || isDemo || !user.WebDAV" @click.stop="showDialog('webdav')">
                <translate>Connect via WebDAV</translate>
                <v-icon :right="!rtl" :left="rtl" dark>sync_alt</v-icon>
              </v-btn>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-1">
          <h3 class="body-2 mb-0">
            <translate>Work Details</translate>
          </h3>
        </v-card-title>

        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                  v-model="user.Details.OrgName"
                  hide-details required box flat
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Organization')"
                  class="input-org-name"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs6 class="pa-2">
              <v-text-field
                  v-model="user.Details.OrgURL"
                  hide-details required box flat
                  :disabled="busy"
                  type="url"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('URL')"
                  class="input-org-url"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs6 class="pa-2">
              <v-text-field
                  v-model="user.Details.OrgTitle"
                  hide-details required box flat
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Title')"
                  class="input-position"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                  v-model="user.Details.OrgEmail"
                  hide-details required box flat
                  :disabled="busy"
                  type="email"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Email')"
                  class="input-org-email"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
      <v-card flat tile class="mt-0 px-1 application">
        <v-card-title primary-title class="pb-1">
          <h3 class="body-2 mb-0">
            <translate>Contact Details</translate>
          </h3>
        </v-card-title>
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="user.Details.Location"
                  hide-details required box flat
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Location')"
                  class="input-location"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-autocomplete
                  v-model="user.Details.Country"
                  :disabled="busy"
                  :label="$gettext('Country')" hide-details box flat
                  hide-no-data
                  browser-autocomplete="off"
                  color="secondary-dark"
                  item-value="Code"
                  item-text="Name"
                  :items="countries"
                  class="input-country"
                  @change="onChange"
              >
              </v-autocomplete>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                  v-model="user.Details.Phone"
                  hide-details required box flat
                  :disabled="busy"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Phone')"
                  class="input-phone"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                  v-model="user.Details.ProfileURL"
                  hide-details required box flat
                  :disabled="busy"
                  type="url"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Profile')"
                  class="input-profile-url"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                  v-model="user.Details.FeedURL"
                  hide-details required box flat
                  :disabled="busy"
                  type="url"
                  browser-autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Feed')"
                  class="input-feed-url"
                  color="secondary-dark"
                  @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-textarea v-model="user.Details.Bio"  auto-grow flat box hide-details
                          rows="3" class="input-bio" color="secondary-dark"
                          autocorrect="off" autocapitalize="none" browser-autocomplete="off"
                          :disabled="busy"
                          :label="$gettext('Bio')"
                          @change="onChange"></v-textarea>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>
    <p-about-footer></p-about-footer>
    <p-account-password-dialog :show="dialog.password" @cancel="dialog.password = false" @confirm="dialog.password = false"></p-account-password-dialog>
    <p-webdav-dialog :show="dialog.webdav" @close="dialog.webdav = false"></p-webdav-dialog>
  </div>
</template>

<script>
import PAccountPasswordDialog from "dialog/account/password.vue";
import countries from "options/countries.json";
import Notify from "common/notify";
import Api from "common/api";
import User from "model/user";

export default {
  name: 'PSettingsAccount',
  components: {PAccountPasswordDialog},
  data() {
    const isDemo = this.$config.isDemo();
    const isPublic = this.$config.isPublic();
    return {
      busy: isDemo || isPublic,
      isDemo: isDemo,
      isPublic: isPublic,
      rtl: this.$rtl,
      user: new User(this.$session.getUser()),
      countries: countries,
      dialog: {
        password: false,
        webdav: false,
      },
    };
  },
  created() {
    if(this.isPublic && !this.isDemo) {
      this.$router.push({ name: "settings" });
    }
  },
  computed: {
    displayName() {
      const user = this.$session.getUser();
      if (user) {
        return user.getDisplayName();
      }

      return this.$gettext("Unregistered");
    },
  },
  methods: {
    showDialog(name) {
      if (!name) {
        return;
      }
      this.dialog[name] = true;
    },
    disabled() {
      return (this.isDemo || this.busy);
    },
    onChangeAvatar() {
      if (this.busy) {
        return;
      }
      this.$refs.upload.click();
    },
    onLogout() {
      this.$session.logout();
    },
    onChange() {
      if (this.busy) {
        return;
      }
      this.busy = true;
      this.user.update().then((u) => {
        this.user = new User(u);
        this.$session.setUser(u);
        this.$notify.success(this.$gettext("Settings saved"));
      }).finally(() => this.busy = false);
    },
    onUploadAvatar() {
      if (this.busy) {
        return;
      }

      this.busy = true;

      Notify.info(this.$gettext("Uploading…"));

      this.user.uploadAvatar(this.$refs.upload.files).then((u) => {
        this.user = new User(u);
        this.$session.setUser(u);
        this.$notify.success(this.$gettext("Settings saved"));
      }).finally(() => this.busy = false);
    }
  },
};
</script>
