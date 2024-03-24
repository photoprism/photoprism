<template>
  <v-container id="auth-login" fluid fill-height class="auth-login wallpaper background-welcome pa-4"
               :style="wallpaper()">
    <v-layout id="auth-layout" class="auth-layout">
      <v-flex xs12 sm9 md6 lg4 xl3 xxl2>
        <v-form ref="form" dense class="auth-login-form" accept-charset="UTF-8" @submit.prevent="onLogin">
          <v-card id="auth-login-box" class="elevation-12 auth-login-box blur-7">
            <v-card-text class="pa-4">
              <p-auth-header></p-auth-header>
              <v-spacer></v-spacer>
              <v-layout wrap align-top>
                <v-flex xs12 class="px-2 py-1">
                  <v-text-field
                    id="auth-username"
                    v-model="username"
                    :disabled="loading || enterPasscode"
                    name="username"
                    type="text"
                    :label="$gettext('Name')"
                    hide-details
                    required
                    solo
                    flat
                    light
                    autofocus
                    autocorrect="off"
                    autocapitalize="none"
                    autocomplete="username"
                    browser-autocomplete="username"
                    background-color="grey lighten-5"
                    class="input-username text-selectable"
                    color="primary"
                    prepend-inner-icon="person"
                    @keyup.enter.native="onLogin"
                  ></v-text-field>
                </v-flex>
                <v-flex v-if="enterPasscode" xs12 class="pa-2">
                  <v-text-field
                    id="auth-passcode"
                    v-model="passcode"
                    :disabled="loading"
                    name="passcode"
                    type="text"
                    :label="$gettext('Verification Code')"
                    mask="nnn nnn nnn nnn"
                    inputmode="text"
                    hide-details
                    required
                    solo
                    flat
                    light
                    autocorrect="off"
                    autocapitalize="none"
                    autocomplete="one-time-code"
                    browser-autocomplete="one-time-code"
                    background-color="grey lighten-5"
                    class="input-passcode text-selectable"
                    prepend-inner-icon="verified_user"
                    color="primary"
                    @keyup.enter.native="onLogin"
                  ></v-text-field>
                </v-flex>
                <v-flex v-else xs12 class="pa-2">
                  <v-text-field
                    id="auth-password"
                    v-model="password"
                    :disabled="loading"
                    name="password"
                    :type="showPassword ? 'text' : 'password'"
                    :label="$gettext('Password')"
                    hide-details
                    required
                    solo
                    flat
                    light
                    autocorrect="off"
                    autocapitalize="none"
                    autocomplete="current-password"
                    browser-autocomplete="current-password"
                    background-color="grey lighten-5"
                    class="input-password text-selectable"
                    :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                    prepend-inner-icon="lock"
                    color="primary"
                    @click:append="showPassword = !showPassword"
                    @keyup.enter.native="onLogin"
                  ></v-text-field>
                </v-flex>
                <v-flex xs12 class="px-2 py-1 auth-actions">
                  <div class="action-buttons auth-buttons text-xs-center">
                    <v-btn v-if="enterPasscode" :color="colors.secondary" outline :block="$vuetify.breakpoint.xsOnly"
                           :style="`color: ${colors.link}!important`" class="action-cancel ra-6 px-3 py-2 opacity-80"
                           @click.stop.prevent="onCancel">
                      <translate>Cancel</translate>
                    </v-btn>
                    <v-btn v-else-if="registerUri" :color="colors.secondary" outline :block="$vuetify.breakpoint.xsOnly"
                           :style="`color: ${colors.link}!important`" class="action-register ra-6 px-3 py-2 opacity-80"
                           @click.stop.prevent="onRegister">
                      <translate>Create Account</translate>
                    </v-btn>
                    <v-btn :color="colors.primary" depressed :disabled="loginDisabled"
                           :block="$vuetify.breakpoint.xsOnly" class="white--text action-confirm ra-6 py-2 px-3"
                           @click.stop.prevent="onLogin">
                      <translate>Sign in</translate>
                      <v-icon v-if="rtl" left dark>navigate_before</v-icon>
                      <v-icon v-else right dark>navigate_next</v-icon>
                    </v-btn>
                  </div>
                  <div v-if="passwordResetUri" class="auth-links text-xs-center opacity-80">
                    <a :href="passwordResetUri" class="text-link link--text">
                      <translate>Forgot password?</translate>
                    </a>
                  </div>
                </v-flex>
              </v-layout>
            </v-card-text>
          </v-card>
        </v-form>
      </v-flex>
    </v-layout>
    <p-auth-footer :colors="colors"></p-auth-footer>
  </v-container>
</template>

<script>
export default {
  name: "PPageLogin",
  data() {
    return {
      colors: {
        accent: "#05dde1",
        primary: "#00a6a9",
        secondary: "#505050",
        link: "#c8e3e7",
      },
      loading: false,
      username: "",
      password: "",
      showPassword: false,
      passcode: "",
      enterPasscode: false,
      sponsor: this.$config.isSponsor(),
      config: this.$config.values,
      siteDescription: this.$config.getSiteDescription(),
      nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
      wallpaperUri: this.$config.values.wallpaperUri,
      registerUri: this.$config.values.registerUri,
      passwordResetUri: this.$config.values.passwordResetUri,
      rtl: this.$rtl,
    };
  },
  computed: {
    loginDisabled() {
      return this.loading || this.username.trim() === "" || this.password.trim() === "";
    },
  },
  created() {
    this.$scrollbar.hide(this.$isMobile);
  },
  destroyed() {
    this.$scrollbar.show();
  },
  methods: {
    wallpaper() {
      if (this.wallpaperUri) {
        return `background-image: url(${this.wallpaperUri});`;
      }

      return "";
    },
    load() {
      this.$notify.blockUI();

      let route = this.$router.resolve({
        name: this.$session.getHome(),
      });

      setTimeout(() => {
        window.location = route.href;
      }, 100);
    },
    reset() {
      this.username = "";
      this.password = "";
      this.showPassword = false;
      this.passcode = "";
      this.enterPasscode = false;
    },
    onCancel() {
      if (this.loading) {
        return;
      }
      this.reset();
    },
    onRegister() {
      window.location = this.registerUri;
    },
    onLogin() {
      const username = this.username.trim();
      const password = this.password.trim();
      const passcode = this.passcode.trim();

      if (username === "" || password === "") {
        return;
      }

      this.loading = true;
      this.$session
        .login(username, password, passcode)
        .then(() => {
          this.load();
        })
        .catch((e) => {
          if (e.response?.data?.code === 32) {
            this.enterPasscode = true;
          }
          this.loading = false;
        });
    },
  },
};
</script>
