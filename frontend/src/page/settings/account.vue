<template>
  <div class="p-tab p-settings-account">
    <div class="width-lg pa-3">
      <v-form
        ref="form"
        v-model="valid"
        class="p-form-account ma-0 pa-0"
        accept-charset="UTF-8"
        @submit.prevent="onChange"
      >
        <input
          ref="upload"
          type="file"
          class="d-none input-upload"
          accept="image/png, image/jpeg"
          @change.stop="onUploadAvatar()"
        />
        <v-card flat tile class="bg-background ma-0 pa-0">
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="8" sm="9" md="10" align-self="stretch" class="pa-0 d-flex">
                <v-row align="start" dense>
                  <v-col md="2" class="hidden-sm-and-down">
                    <v-text-field
                      v-model="user.Details.NameTitle"
                      density="comfortable"
                      :disabled="busy"
                      maxlength="32"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$pgettext('Account', 'Title')"
                      class="input-name-title"
                      :rules="rules.text(false, 0, 32, $pgettext('Account', 'Title'))"
                      @change="onChangeName"
                    ></v-text-field>
                  </v-col>
                  <v-col md="6" class="hidden-sm-and-down">
                    <v-text-field
                      v-model="user.Details.GivenName"
                      density="comfortable"
                      :disabled="busy"
                      maxlength="64"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Given Name')"
                      class="input-given-name"
                      :rules="rules.text(false, 1, 64, $gettext('Given Name'))"
                      @change="onChangeName"
                    ></v-text-field>
                  </v-col>
                  <v-col md="4" class="hidden-sm-and-down">
                    <v-text-field
                      v-model="user.Details.FamilyName"
                      density="comfortable"
                      :disabled="busy"
                      maxlength="64"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Family Name')"
                      class="input-family-name"
                      :rules="rules.text(false, 1, 64, $gettext('Family Name'))"
                      @change="onChangeName"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" md="5">
                    <v-text-field
                      v-model="user.DisplayName"
                      :disabled="busy"
                      maxlength="200"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Display Name')"
                      class="input-display-name"
                      :rules="rules.text(true, 1, 200, $gettext('Display Name'))"
                      @change="onChange"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" md="7">
                    <v-text-field
                      v-model="user.Email"
                      type="email"
                      maxlength="255"
                      :disabled="busy"
                      autocomplete="off"
                      autocorrect="off"
                      autocapitalize="none"
                      :label="$gettext('Email')"
                      class="input-email"
                      :rules="rules.email()"
                      @change="onChange"
                    ></v-text-field>
                  </v-col>
                </v-row>
              </v-col>
              <v-col class="text-center" cols="4" sm="3" md="2" align-self="center">
                <v-avatar
                  :size="$vuetify.display.md ? 100 : 112"
                  :class="{ clickable: !busy }"
                  @click.stop.prevent="onChangeAvatar()"
                >
                  <v-img
                    :alt="accountInfo"
                    :title="$gettext('Change Avatar')"
                    :src="$vuetify.display.md ? user.getAvatarURL('tile_100') : user.getAvatarURL('tile_224')"
                  ></v-img>
                </v-avatar>
              </v-col>
              <v-col v-if="user.Details.Bio" cols="12">
                <v-textarea
                  v-model="user.Details.Bio"
                  auto-grow
                  rows="2"
                  class="input-bio"
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  :disabled="busy"
                  maxlength="2000"
                  :rules="rules.text(false, 1, 2000, $gettext('Bio'))"
                  :label="$gettext('Bio')"
                  @change="onChange"
                ></v-textarea>
              </v-col>
              <v-col cols="12">
                <v-textarea
                  v-model="user.Details.About"
                  auto-grow
                  rows="2"
                  class="input-about"
                  autocorrect="off"
                  autocapitalize="none"
                  autocomplete="off"
                  :disabled="busy"
                  maxlength="500"
                  :rules="rules.text(false, 1, 500, $gettext('About'))"
                  :label="$gettext('About')"
                  @change="onChange"
                ></v-textarea>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
        <v-card flat tile class="my-3 pa-0 bg-background">
          <v-card-title class="ma-0 pa-2 text-subtitle-2">
            {{ $gettext(`Security and Access`) }}
          </v-card-title>
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="12" sm="6">
                <v-btn
                  block
                  variant="flat"
                  color="button"
                  class="action-change-password"
                  :disabled="isPublic || isDemo || user.Name === '' || getProvider() !== 'local'"
                  @click.stop="showDialog('password')"
                >
                  {{ $gettext(`Change Password`) }}
                  <v-icon end>mdi-lock</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="12" sm="6">
                <v-btn
                  block
                  variant="flat"
                  color="button"
                  class="action-passcode-dialog"
                  :disabled="isPublic || isDemo || user.disablePasscodeSetup(session.hasPassword())"
                  @click.stop="showDialog('passcode')"
                >
                  {{ $gettext(`2-Factor Authentication`) }}
                  <v-icon v-if="user.AuthMethod === '2fa'" end>mdi-shield-alert</v-icon>
                  <v-icon v-else-if="user.disablePasscodeSetup(session.hasPassword())" end>mdi-shield-check</v-icon>
                  <v-icon v-else end>mdi-shield-alert</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="12" sm="6">
                <v-btn
                  block
                  variant="flat"
                  color="button"
                  class="action-apps-dialog"
                  :disabled="isPublic || isDemo || user.Name === ''"
                  @click.stop="showDialog('apps')"
                >
                  {{ $gettext(`Apps and Devices`) }}
                  <v-icon end>mdi-cellphone-link</v-icon>
                </v-btn>
              </v-col>
              <v-col cols="12" sm="6">
                <v-btn
                  block
                  variant="flat"
                  color="button"
                  class="action-webdav-dialog"
                  :disabled="isPublic || isDemo || !user.hasWebDAV()"
                  @click.stop="showDialog('webdav')"
                >
                  {{ $gettext(`Connect via WebDAV`) }}
                  <v-icon end>mdi-swap-horizontal</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
        <v-card flat tile class="my-3 pa-0 bg-background">
          <v-card-title class="ma-0 pa-2 text-subtitle-2">
            {{ $gettext(`Birth Date`) }}
          </v-card-title>
          <v-card-actions class="ma-0 pa-0">
            <v-row align="start" dense>
              <v-col cols="6" sm="3">
                <v-combobox
                  :model-value="user?.Details?.BirthDay > 0 ? user.Details.BirthDay : null"
                  :disabled="busy"
                  :label="$gettext('Day')"
                  :placeholder="$gettext('Unknown')"
                  density="comfortable"
                  autocomplete="off"
                  hide-no-data
                  hide-details
                  item-title="text"
                  item-value="value"
                  :items="options.Days()"
                  validate-on="input"
                  :rules="rules.day(false)"
                  class="input-birth-day"
                  @update:model-value="setBirthDay"
                >
                </v-combobox>
              </v-col>
              <v-col cols="6" sm="3">
                <v-combobox
                  :model-value="user?.Details?.BirthMonth > 0 ? user.Details.BirthMonth : null"
                  :disabled="busy"
                  :label="$gettext('Month')"
                  :placeholder="$gettext('Unknown')"
                  density="comfortable"
                  autocomplete="off"
                  hide-no-data
                  hide-details
                  item-title="text"
                  item-value="value"
                  :items="options.MonthsShort()"
                  validate-on="input"
                  :rules="rules.month(false)"
                  class="input-birth-month"
                  @update:model-value="setBirthMonth"
                >
                </v-combobox>
              </v-col>
              <v-col cols="12" sm="6">
                <v-combobox
                  :model-value="user?.Details?.BirthYear > 0 ? user.Details.BirthYear : null"
                  :disabled="busy"
                  :label="$gettext('Year')"
                  :placeholder="$gettext('Unknown')"
                  density="comfortable"
                  autocomplete="off"
                  hide-no-data
                  hide-details
                  :items="options.Years(1900)"
                  validate-on="input"
                  :rules="rules.year(false, 1000)"
                  class="input-birth-year"
                  @update:model-value="setBirthYear"
                >
                </v-combobox>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
        <v-card flat tile class="my-3 pa-0 bg-background">
          <v-card-title class="ma-0 pa-2 text-subtitle-2">
            {{ $gettext(`Contact Details`) }}
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
                  :rules="rules.text(false, 1, 500, $gettext('Location'))"
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
                  :rules="rules.country()"
                  @update:model-value="onChange"
                >
                </v-autocomplete>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="user.Details.SiteURL"
                  density="comfortable"
                  :disabled="busy"
                  type="url"
                  maxlength="500"
                  autocomplete="off"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Website')"
                  class="input-site-url"
                  :rules="rules.url()"
                  @change="onChange"
                ></v-text-field>
              </v-col>
            </v-row>
          </v-card-actions>
        </v-card>
      </v-form>
    </div>
    <p-settings-apps :visible="dialog.apps" :model="user" @close="dialog.apps = false"></p-settings-apps>
    <p-settings-passcode
      :visible="dialog.passcode"
      :model="user"
      @close="dialog.passcode = false"
      @updateUser="updateUser()"
    ></p-settings-passcode>
    <p-settings-password
      :visible="dialog.password"
      :model="user"
      @close="dialog.password = false"
    ></p-settings-password>
    <p-settings-webdav :visible="dialog.webdav" @close="dialog.webdav = false"></p-settings-webdav>
  </div>
</template>

<script>
import PSettingsApps from "component/settings/apps.vue";
import PSettingsPasscode from "component/settings/passcode.vue";
import PSettingsPassword from "component/settings/password.vue";
import PSettingsWebdav from "component/settings/webdav.vue";
import countries from "options/countries.json";
import User from "model/user";
import * as options from "options/options";
import { rules } from "common/form";

export default {
  name: "PSettingsAccount",
  components: {
    PSettingsApps,
    PSettingsPasscode,
    PSettingsPassword,
    PSettingsWebdav,
  },
  data() {
    const isDemo = this.$config.isDemo();
    const isPublic = this.$config.isPublic();
    const user = this.$session.getUser();

    return {
      busy: isDemo || isPublic,
      options,
      rules,
      isDemo,
      isPublic,
      valid: true,
      rtl: this.$isRtl,
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
  mounted() {
    this.$refs.form.validate();
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
    onChangeAvatar() {
      if (this.busy) {
        return;
      }
      this.$refs.upload.click();
    },
    onChangeName() {
      this.user.Details.NameSrc = "manual";
      return this.onChange();
    },
    onChange() {
      if (this.busy || !this?.$refs?.form) {
        return;
      }

      this.busy = true;

      this.$refs.form
        .validate()
        .then((form) => {
          if (form.valid) {
            this.user
              .update()
              .then((u) => {
                this.user = new User(u);
                this.$session.setUser(u);
                this.$notify.success(this.$gettext("Settings saved"));
              })
              .finally(() => {
                this.busy = false;
              });
          } else {
            this.$notify.error(this.$gettext("Changes could not be saved"));
            this.busy = false;
          }
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Changes could not be saved"));
          this.busy = false;
        });
    },
    onUploadAvatar() {
      if (this.busy) {
        return;
      }

      this.busy = true;

      this.$notify.info(this.$gettext("Uploading…"));

      this.user
        .uploadAvatar(this.$refs.upload.files)
        .then((u) => {
          this.user = new User(u);
          this.$session.setUser(u);
          this.$notify.success(this.$gettext("Settings saved"));
        })
        .finally(() => (this.busy = false));
    },
    setBirthDay(v) {
      if (Number.isInteger(v?.value)) {
        this.user.Details.BirthDay = v?.value;
        this.onChange();
      } else if (!v) {
        this.user.Details.BirthDay = -1;
        this.onChange();
      } else if (this.rules.isNumberRange(v, 1, 31)) {
        this.user.Details.BirthDay = Number(v);
        this.onChange();
      }
    },
    setBirthMonth(v) {
      if (Number.isInteger(v?.value)) {
        this.user.Details.BirthMonth = v?.value;
        this.onChange();
      } else if (!v) {
        this.user.Details.BirthMonth = -1;
        this.onChange();
      } else if (this.rules.isNumberRange(v, 1, 12)) {
        this.user.Details.BirthMonth = Number(v);
        this.onChange();
      }
    },
    setBirthYear(v) {
      if (Number.isInteger(v?.value)) {
        this.user.Details.BirthYear = v?.value;
        this.onChange();
      } else if (!v) {
        this.user.Details.BirthYear = -1;
        this.onChange();
      } else if (this.rules.isNumberRange(v, 1000, Number(new Date().getUTCFullYear()))) {
        this.user.Details.BirthYear = Number(v);
        this.onChange();
      }
    },
  },
};
</script>
