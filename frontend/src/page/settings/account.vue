<template>
  <div class="p-tab p-settings-account">
    <v-form ref="form" v-model="valid" lazy-validation dense class="p-form-account pb-4 width-lg" accept-charset="UTF-8" @submit.prevent="onChange">
      <input ref="upload" type="file" class="d-none input-upload" accept="image/png, image/jpeg" @change.stop="onUploadAvatar()" />
      <v-card flat tile class="mt-2 px-1 application">
        <v-card-actions>
          <v-layout row wrap align-top>
            <v-flex xs8 sm9 md10 fill-height class="pa-0">
              <v-layout wrap align-top>
                <v-flex md2 class="pa-2 hidden-sm-and-down">
                  <v-select
                    v-model="user.Details.Gender"
                    :label="$gettext('Gender')"
                    hide-details
                    box
                    flat
                    :disabled="busy"
                    item-text="text"
                    item-value="value"
                    color="secondary-dark"
                    :items="options.Gender()"
                    class="input-gender"
                    :rules="[(v) => validLength(v, 0, 16) || $gettext('Invalid')]"
                    @change="onChange"
                  >
                  </v-select>
                </v-flex>
                <v-flex md2 class="pa-2 hidden-sm-and-down">
                  <v-text-field
                    v-model="user.Details.NameTitle"
                    hide-details
                    required
                    box
                    flat
                    :disabled="busy"
                    maxlength="32"
                    browser-autocomplete="off"
                    autocorrect="off"
                    autocapitalize="none"
                    :label="$pgettext('Account', 'Title')"
                    class="input-name-title"
                    color="secondary-dark"
                    :rules="[(v) => validLength(v, 0, 32) || $gettext('Invalid')]"
                    @change="onChangeName"
                  ></v-text-field>
                </v-flex>
                <v-flex md4 class="pa-2 hidden-sm-and-down">
                  <v-text-field
                    v-model="user.Details.GivenName"
                    hide-details
                    required
                    box
                    flat
                    :disabled="busy"
                    maxlength="64"
                    browser-autocomplete="off"
                    autocorrect="off"
                    autocapitalize="none"
                    :label="$gettext('Given Name')"
                    class="input-given-name"
                    color="secondary-dark"
                    :rules="[(v) => validLength(v, 0, 64) || $gettext('Invalid')]"
                    @change="onChangeName"
                  ></v-text-field>
                </v-flex>
                <v-flex md4 class="pa-2 hidden-sm-and-down">
                  <v-text-field
                    v-model="user.Details.FamilyName"
                    hide-details
                    required
                    box
                    flat
                    :disabled="busy"
                    maxlength="64"
                    browser-autocomplete="off"
                    autocorrect="off"
                    autocapitalize="none"
                    :label="$gettext('Family Name')"
                    class="input-family-name"
                    color="secondary-dark"
                    :rules="[(v) => validLength(v, 0, 64) || $gettext('Invalid')]"
                    @change="onChangeName"
                  ></v-text-field>
                </v-flex>
                <v-flex xs12 md4 class="pa-2">
                  <v-text-field
                    v-model="user.DisplayName"
                    hide-details
                    required
                    box
                    flat
                    :disabled="busy"
                    maxlength="200"
                    browser-autocomplete="off"
                    autocorrect="off"
                    autocapitalize="none"
                    :label="$gettext('Display Name')"
                    class="input-display-name"
                    color="secondary-dark"
                    :rules="[(v) => validLength(v, 1, 200) || $gettext('Required')]"
                    @change="onChange"
                  ></v-text-field>
                </v-flex>
                <v-flex xs12 md8 class="pa-2">
                  <v-text-field
                    v-model="user.Email"
                    hide-details
                    required
                    box
                    flat
                    validate-on-blur
                    type="email"
                    maxlength="250"
                    :disabled="busy"
                    browser-autocomplete="off"
                    autocorrect="off"
                    autocapitalize="none"
                    :label="$gettext('Email')"
                    class="input-email"
                    color="secondary-dark"
                    :rules="[(v) => (!!v && validEmail(v)) || $gettext('Invalid')]"
                    @change="onChange"
                  ></v-text-field>
                </v-flex>
              </v-layout>
            </v-flex>

            <v-flex class="pa-2 text-xs-center" xs4 sm3 md2 align-self-center>
              <v-avatar :size="$vuetify.breakpoint.xsOnly ? 100 : 128" :class="{ clickable: !busy }" @click.stop.prevent="onChangeAvatar()">
                <img :src="$vuetify.breakpoint.xsOnly ? user.getAvatarURL('tile_100') : user.getAvatarURL('tile_224')" :alt="accountInfo" :title="$gettext('Change Avatar')" />
              </v-avatar>
            </v-flex>

            <v-flex v-if="user.Details.Bio" xs12 class="pa-2">
              <v-textarea
                v-model="user.Details.Bio"
                auto-grow
                flat
                box
                hide-details
                rows="2"
                class="input-bio"
                color="secondary-dark"
                autocorrect="off"
                autocapitalize="none"
                browser-autocomplete="off"
                :disabled="busy"
                maxlength="2000"
                :rules="[(v) => validLength(v, 0, 2000) || $gettext('Invalid')]"
                :label="$gettext('Bio')"
                @change="onChange"
              ></v-textarea>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-textarea
                v-model="user.Details.About"
                auto-grow
                flat
                box
                hide-details
                rows="2"
                class="input-about"
                color="secondary-dark"
                autocorrect="off"
                autocapitalize="none"
                browser-autocomplete="off"
                :disabled="busy"
                maxlength="500"
                :rules="[(v) => validLength(v, 0, 500) || $gettext('Invalid')]"
                :label="$gettext('About')"
                @change="onChange"
              ></v-textarea>
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
              <v-btn block depressed color="secondary-light" class="action-change-password compact" :disabled="isPublic || isDemo || user.Name === '' || getProvider() !== 'local'" @click.stop="showDialog('password')">
                <translate>Change Password</translate>
                <v-icon :right="!rtl" :left="rtl" dark>lock</v-icon>
              </v-btn>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-btn block depressed color="secondary-light" class="action-passcode-dialog compact" :disabled="isPublic || isDemo || user.disablePasscodeSetup(session.hasPassword())" @click.stop="showDialog('passcode')">
                <translate>2-Factor Authentication</translate>
                <v-icon v-if="user.AuthMethod === '2fa'" :right="!rtl" :left="rtl" dark>gpp_good</v-icon>
                <v-icon v-else-if="user.disablePasscodeSetup(session.hasPassword())" :right="!rtl" :left="rtl" dark>shield</v-icon>
                <v-icon v-else :right="!rtl" :left="rtl" dark>gpp_maybe</v-icon>
              </v-btn>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-btn block depressed color="secondary-light" class="action-apps-dialog compact" :disabled="isPublic || isDemo || user.Name === ''" @click.stop="showDialog('apps')">
                <translate>Apps and Devices</translate>
                <v-icon :right="!rtl" :left="rtl" dark>devices</v-icon>
              </v-btn>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-btn block depressed color="secondary-light" class="action-webdav-dialog compact" :disabled="isPublic || isDemo || !user.hasWebDAV()" @click.stop="showDialog('webdav')">
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
            <translate>Birth Date</translate>
          </h3>
        </v-card-title>
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs3 class="pa-2">
              <v-autocomplete v-model="user.Details.BirthDay" :disabled="busy" :label="$gettext('Day')" browser-autocomplete="off" hide-no-data hide-details box flat color="secondary-dark" :items="options.Days()" class="input-birth-day" @change="onChange"> </v-autocomplete>
            </v-flex>
            <v-flex xs3 class="pa-2">
              <v-autocomplete v-model="user.Details.BirthMonth" :disabled="busy" :label="$gettext('Month')" browser-autocomplete="off" hide-no-data hide-details box flat color="secondary-dark" :items="options.MonthsShort()" class="input-birth-month" @change="onChange"> </v-autocomplete>
            </v-flex>
            <v-flex xs6 class="pa-2">
              <v-autocomplete v-model="user.Details.BirthYear" :disabled="busy" :label="$gettext('Year')" browser-autocomplete="off" hide-no-data hide-details box flat color="secondary-dark" :items="options.Years()" class="input-birth-year" @change="onChange"> </v-autocomplete>
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
                hide-details
                required
                box
                flat
                :disabled="busy"
                maxlength="500"
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Location')"
                class="input-location"
                color="secondary-dark"
                :rules="[(v) => validLength(v, 0, 500) || $gettext('Invalid')]"
                @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm4 class="pa-2">
              <v-autocomplete
                v-model="user.Details.Country"
                :disabled="busy"
                :label="$gettext('Country')"
                hide-no-data
                hide-details
                box
                flat
                browser-autocomplete="off"
                color="secondary-dark"
                item-value="Code"
                item-text="Name"
                :items="countries"
                class="input-country"
                :rules="[(v) => validLength(v, 0, 2) || $gettext('Invalid')]"
                @change="onChange"
              >
              </v-autocomplete>
            </v-flex>
            <v-flex xs12 sm8 class="pa-2">
              <v-text-field
                v-model="user.Details.Phone"
                hide-details
                required
                box
                flat
                :disabled="busy"
                maxlength="32"
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Phone')"
                class="input-phone"
                color="secondary-dark"
                :rules="[(v) => validLength(v, 0, 32) || $gettext('Invalid')]"
                @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                v-model="user.Details.SiteURL"
                hide-details
                required
                box
                flat
                :disabled="busy"
                type="url"
                maxlength="500"
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Website')"
                class="input-site-url"
                color="secondary-dark"
                :rules="[(v) => validUrl(v) || $gettext('Invalid')]"
                @change="onChange"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pa-2">
              <v-text-field
                v-model="user.Details.FeedURL"
                hide-details
                required
                box
                flat
                :disabled="busy"
                type="url"
                maxlength="500"
                browser-autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Feed')"
                class="input-feed-url"
                color="secondary-dark"
                :rules="[(v) => validUrl(v) || $gettext('Invalid')]"
                @change="onChange"
              ></v-text-field>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>
    <p-account-apps-dialog :show="dialog.apps" :model="user" @close="dialog.apps = false"></p-account-apps-dialog>
    <p-account-passcode-dialog :show="dialog.passcode" :model="user" @close="dialog.passcode = false" @updateUser="updateUser()"></p-account-passcode-dialog>
    <p-account-password-dialog :show="dialog.password" :model="user" @close="dialog.password = false"></p-account-password-dialog>
    <p-webdav-dialog :show="dialog.webdav" @close="dialog.webdav = false"></p-webdav-dialog>
  </div>
</template>

<script>
import PAccountPasswordDialog from "dialog/account/password.vue";
import countries from "options/countries.json";
import Notify from "common/notify";
import * as options from "options/options";

export default {
  name: "PSettingsAccount",
  components: { PAccountPasswordDialog },
  data() {
    const isDemo = this.$config.isDemo();
    const isPublic = this.$config.isPublic();
    const user = this.$session.getUser();

    return {
      busy: isDemo || isPublic,
      options,
      isDemo,
      isPublic,
      valid: true,
      rtl: this.$rtl,
      user: user,
      countries: countries,
      session: this.$session,
      dialog: {
        apps: false,
        passcode: false,
        password: false,
        webdav: false,
      },
    };
  },
  computed: {
    accountInfo() {
      const user = this.$session.getUser();
      if (user) {
        return user.getAccountInfo();
      }

      return this.$gettext("Unregistered");
    },
    displayName() {
      const user = this.$session.getUser();
      if (user) {
        return user.getDisplayName();
      }

      return this.$gettext("Unregistered");
    },
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$router.push({ name: "settings" });
    }
  },
  methods: {
    getProvider() {
      return this.$session.provider ? this.$session.provider : this.user.AuthProvider;
    },
    showDialog(name) {
      if (!name) {
        return;
      }
      this.dialog[name] = true;
    },
    updateUser() {
      this.$notify.blockUI();
      this.$session
        .refresh()
        .then(() => {
          this.user = this.$session.getUser();
        })
        .finally(() => {
          this.$notify.unblockUI();
        });
    },
    validEmail(v) {
      if (typeof v !== "string" || v === "") {
        return true;
      } else if (!this.validLength(v, 0, 250)) {
        return false;
      }

      return /^\w+([.-]?\w+)*@\w+([.-]?\w+)*(\.\w{2,32})+$/.test(v);
    },
    validLength(v, min, max) {
      if (typeof v !== "string" && min <= 0) {
        return true;
      } else if (max > 0 && v.length > max) {
        return false;
      }

      return v.length >= min;
    },
    validUrl(v) {
      if (typeof v !== "string" || v === "") {
        return true;
      } else if (!this.validLength(v, 0, 500)) {
        return false;
      }

      try {
        new URL(v);
      } catch (e) {
        return false;
      }
      return true;
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
    onChangeName() {
      this.user.Details.NameSrc = "manual";
      return this.onChange();
    },
    onChange() {
      if (this.busy || !this.valid) {
        return;
      }
      this.busy = true;
      this.user
        .update()
        .then((user) => {
          this.user = user;
          this.$session.setUser(user);
          this.$notify.success(this.$gettext("Changes successfully saved"));
        })
        .finally(() => (this.busy = false));
    },
    onUploadAvatar() {
      if (this.busy) {
        return;
      }

      this.busy = true;

      Notify.info(this.$gettext("Updating picture…"));

      this.user
        .uploadAvatar(this.$refs.upload.files)
        .then((user) => {
          this.user = user;
          this.$session.setUser(user);
          this.$notify.success(this.$gettext("Changes successfully saved"));
        })
        .finally(() => (this.busy = false));
    },
  },
};
</script>
