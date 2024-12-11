<template>
  <div class="p-tab p-settings-account">
    <v-container class="width-lg pa-3">
      <v-form ref="form" v-model="valid" validate-on="blur" class="p-form-account ma-0 pa-0" accept-charset="UTF-8" @submit.prevent="onChange">
        <input ref="upload" type="file" class="d-none input-upload" accept="image/png, image/jpeg" @change.stop="onUploadAvatar()" />
        <v-card flat tile class="bg-background ma-0 pa-0">
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="8" sm="9" md="10" align-self="stretch" class="pa-0 d-flex">
                <v-row align="start" dense>
                  <v-col md="2" class="hidden-sm-and-down">
                    <v-text-field
                      v-model="user.Details.NameTitle"
                      required
                      density="comfortable"
                      :disabled="busy"
                      maxlength="32"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Title')"
                      class="input-name-title"
                      :rules="[(v) => validLength(v, 0, 32) || $gettext('Invalid')]"
                      @change="onChangeName"
                    ></v-text-field>
                  </v-col>
                  <v-col md="6" class="hidden-sm-and-down">
                    <v-text-field
                      v-model="user.Details.GivenName"
                      hide-details
                      required
                      density="comfortable"
                      :disabled="busy"
                      maxlength="64"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Given Name')"
                      class="input-given-name"
                      :rules="[(v) => validLength(v, 0, 64) || $gettext('Invalid')]"
                      @change="onChangeName"
                    ></v-text-field>
                  </v-col>
                  <v-col md="4" class="hidden-sm-and-down">
                    <v-text-field
                      v-model="user.Details.FamilyName"
                      hide-details
                      required
                      density="comfortable"
                      flat
                      :disabled="busy"
                      maxlength="64"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Family Name')"
                      class="input-family-name"
                      :rules="[(v) => validLength(v, 0, 64) || $gettext('Invalid')]"
                      @change="onChangeName"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" md="5">
                    <v-text-field
                      v-model="user.DisplayName"
                      hide-details
                      required
                      flat
                      :disabled="busy"
                      maxlength="200"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Display Name')"
                      class="input-display-name"
                      :rules="[(v) => validLength(v, 1, 200) || $gettext('Required')]"
                      @change="onChange"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" md="7">
                    <v-text-field
                      v-model="user.Email"
                      hide-details
                      required
                      flat
                      validate-on="blur"
                      type="email"
                      maxlength="255"
                      :disabled="busy"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Email')"
                      class="input-email"
                      :rules="[(v) => validEmail(v) || $gettext('Invalid')]"
                      @change="onChange"
                    ></v-text-field>
                  </v-col>
                </v-row>
              </v-col>
              <v-col class="text-center" cols="4" sm="3" md="2" align-self="center">
                <v-avatar :size="$vuetify.display.xs ? 100 : 112" :class="{ clickable: !busy }" @click.stop.prevent="onChangeAvatar()">
                  <v-img :alt="accountInfo" :title="$gettext('Change Avatar')" :src="$vuetify.display.xs ? user.getAvatarURL('tile_100') : user.getAvatarURL('tile_224')"></v-img>
                </v-avatar>
              </v-col>
              <v-col v-if="user.Details.Bio" cols="12">
                <v-textarea
                  v-model="user.Details.Bio"
                  auto-grow
                  hide-details
                  rows="2"
                  class="input-bio"
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  :disabled="busy"
                  maxlength="2000"
                  :rules="[(v) => validLength(v, 0, 2000) || $gettext('Invalid')]"
                  :label="$gettext('Bio')"
                  @change="onChange"
                ></v-textarea>
              </v-col>
              <v-col cols="12">
                <v-textarea
                  v-model="user.Details.About"
                  auto-grow
                  hide-details
                  rows="2"
                  class="input-about"
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  :disabled="busy"
                  maxlength="500"
                  :rules="[(v) => validLength(v, 0, 500) || $gettext('Invalid')]"
                  :label="$gettext('About')"
                  @change="onChange"
                ></v-textarea>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
        <v-card flat tile class="my-3 pa-0 bg-background">
          <v-card-title class="ma-0 pa-2 text-subtitle-2">
            <translate>Security and Access</translate>
          </v-card-title>
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="12" sm="6">
                <v-btn block variant="flat" color="button" class="action-change-password" :disabled="isPublic || isDemo || user.Name === '' || getProvider() !== 'local'" @click.stop="showDialog('password')">
                  <translate>Change Password</translate>
                  <v-icon :end="!rtl" :start="rtl">mdi-lock</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="12" sm="6">
                <v-btn block variant="flat" color="button" class="action-passcode-dialog" :disabled="isPublic || isDemo || user.disablePasscodeSetup(session.hasPassword())" @click.stop="showDialog('passcode')">
                  <translate>2-Factor Authentication</translate>
                  <v-icon v-if="user.AuthMethod === '2fa'" :end="!rtl" :start="rtl">mdi-shield-alert</v-icon>
                  <v-icon v-else-if="user.disablePasscodeSetup(session.hasPassword())" :end="!rtl" :start="rtl">mdi-shield-check</v-icon>
                  <v-icon v-else :end="!rtl" :start="rtl">mdi-shield-alert</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="12" sm="6">
                <v-btn block variant="flat" color="button" class="action-apps-dialog" :disabled="isPublic || isDemo || user.Name === ''" @click.stop="showDialog('apps')">
                  <translate>Apps and Devices</translate>
                  <v-icon :end="!rtl" :start="rtl">mdi-cellphone-link</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="12" sm="6">
                <v-btn block variant="flat" color="button" class="action-webdav-dialog" :disabled="isPublic || isDemo || !user.hasWebDAV()" @click.stop="showDialog('webdav')">
                  <translate>Connect via WebDAV</translate>
                  <v-icon :end="!rtl" :start="rtl">mdi-swap-horizontal</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
        <v-card flat tile class="my-3 pa-0 bg-background">
          <v-card-title class="ma-0 pa-2 text-subtitle-2">
            <translate>Birth Date</translate>
          </v-card-title>
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="6" sm="3" >
                <v-autocomplete
                  v-model="user.Details.BirthDay"
                  :disabled="busy"
                  :label="$gettext('Day')"
                  autocomplete="off"
                  hide-no-data
                  hide-details
                  item-title="text"
                  item-value="value"
                  :items="options.Days()"
                  :rules="[(v) => v === -1 || (v >= 1 && v <= 31) || $gettext('Invalid')]"
                  density="comfortable"
                  class="input-birth-day"
                  @update:modelValue="onChange"
                >
                </v-autocomplete>
              </v-col>
              <v-col cols="6" sm="3">
                <v-autocomplete
                  v-model="user.Details.BirthMonth"
                  :disabled="busy"
                  :label="$gettext('Month')"
                  autocomplete="off"
                  hide-no-data
                  hide-details
                  item-title="text"
                  item-value="value"
                  :items="options.MonthsShort()"
                  :rules="[(v) => v === -1 || (v >= 1 && v <= 12) || $gettext('Invalid')]"
                  density="comfortable"
                  class="input-birth-month"
                  @update:modelValue="onChange"
                >
                </v-autocomplete>
              </v-col>
              <v-col cols="12" sm="6">
                <v-autocomplete
                  v-model="user.Details.BirthYear"
                  :disabled="busy"
                  :label="$gettext('Year')"
                  autocomplete="off"
                  :items="options.Years()"
                  :rules="[(v) => v === -1 || (v >= 1000 && v <= 9999) || $gettext('Invalid')]"
                  density="comfortable"
                  class="input-birth-year"
                  @update:modelValue="onChange"
                >
                </v-autocomplete>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
        <v-card flat tile class="my-3 pa-0 bg-background">
          <v-card-title class="ma-0 pa-2 text-subtitle-2">
            <translate>Contact Details</translate>
          </v-card-title>
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="12" sm="7">
                <v-text-field
                  v-model="user.Details.Location"
                  density="comfortable"
                  :disabled="busy"
                  maxlength="500"
                  autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Location')"
                  class="input-location"
                  :rules="[(v) => validLength(v, 0, 500) || $gettext('Invalid')]"
                  @change="onChange"
                ></v-text-field>
              </v-col>
              <v-col cols="12" sm="5">
                <v-autocomplete
                  v-model="user.Details.Country"
                  :disabled="busy"
                  :label="$gettext('Country')"
                  density="comfortable"
                  autocomplete="off"
                  item-value="Code"
                  item-title="Name"
                  :items="countries"
                  class="input-country"
                  :rules="[(v) => validLength(v, 0, 2) || $gettext('Invalid')]"
                  @update:modelValue="onChange"
                >
                </v-autocomplete>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="user.Details.SiteURL"
                  hide-details
                  required
                  density="comfortable"
                  :disabled="busy"
                  type="url"
                  maxlength="500"
                  autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('URL')"
                  class="input-site-url"
                  :rules="[(v) => validUrl(v) || $gettext('Invalid')]"
                  @change="onChange"
                ></v-text-field>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
      </v-form>
    </v-container>
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
import User from "model/user";
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
        .then((u) => {
          this.user = new User(u);
          this.$session.setUser(u);
          this.$notify.success(this.$gettext("Settings saved"));
        })
        .finally(() => (this.busy = false));
    },
    onUploadAvatar() {
      if (this.busy) {
        return;
      }

      this.busy = true;

      Notify.info(this.$gettext("Uploading…"));

      this.user
        .uploadAvatar(this.$refs.upload.files)
        .then((u) => {
          this.user = new User(u);
          this.$session.setUser(u);
          this.$notify.success(this.$gettext("Settings saved"));
        })
        .finally(() => (this.busy = false));
    },
  },
};
</script>
