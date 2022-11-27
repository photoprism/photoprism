<template>
  <v-container fluid fill-height class="auth-login wallpaper background-welcome pa-4" :style="wallpaper()">
    <v-layout align-center justify-center>
      <v-flex xs12 sm8 md4 xl3 xxl2>
        <v-form ref="form" dense class="auth-login-form" accept-charset="UTF-8" @submit.prevent="login">
          <v-card class="elevation-12 auth-login-box blur-7">
            <v-card-text class="pa-4">
              <div class="logo text-xs-center">
                <img :src="$config.getIcon()" :alt="config.name">
              </div>
              <v-spacer></v-spacer>
              <v-text-field
                  v-model="username"
                  required hide-details solo flat light autofocus
                  type="text"
                  :disabled="loading"
                  name="username"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Name')"
                  background-color="grey lighten-5"
                  class="input-username text-selectable"
                  :color="colors.accent"
                  :placeholder="$gettext('Name')"
                  prepend-icon="person"
                  @keyup.enter.native="login"
              ></v-text-field>
              <v-text-field
                  v-model="password"
                  required hide-details solo flat light
                  :type="showPassword ? 'text' : 'password'"
                  :disabled="loading"
                  name="password"
                  autocorrect="off"
                  autocapitalize="none"
                  :label="$gettext('Password')"
                  background-color="grey lighten-5"
                  :placeholder="$gettext('Password')"
                  class="input-password mt-1 text-selectable"
                  :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                  prepend-icon="lock"
                  :color="colors.accent"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter.native="login"
              ></v-text-field>
              <v-spacer></v-spacer>
              <div class="action-buttons text-xs-center">
                <v-btn v-if="registerUri" :color="colors.secondary" outline :block="$vuetify.breakpoint.xsOnly"
                       :style="`color: ${colors.link}!important`" class="action-register ra-6 px-3 py-2 opacity-80" @click.stop="register">
                  <translate>Create Account</translate>
                </v-btn>
                <v-btn :color="colors.primary" depressed :disabled="loginDisabled" :block="$vuetify.breakpoint.xsOnly"
                       class="white--text action-confirm ra-6 py-2 px-3" @click.stop="login">
                  <translate>Sign in</translate>
                  <v-icon :right="!rtl" :left="rtl" dark>arrow_forward</v-icon>
                </v-btn>
              </div>
              <div v-if="passwordResetUri" class="text-xs-center opacity-80">
                <a :href="passwordResetUri" class="text-link" :style="`color: ${colors.link}!important`"><translate>Forgot password?</translate></a>
              </div>
            </v-card-text>
          </v-card>
        </v-form>
      </v-flex>
    </v-layout>
    <footer v-if="sponsor">
      <v-layout wrap align-top pa-0 ma-0>
        <v-flex xs12 class="pa-0 body-2 text-selectable text-xs-center white--tex text-sm-left sm6">
          {{ $config.getAbout() }}
        </v-flex>

        <v-flex v-if="config.legalInfo" xs12 sm6 class="pa-0 body-2 text-xs-center text-sm-right white--text">
          <a v-if="config.legalUrl" :href="config.legalUrl" target="_blank" class="text-link"
             :style="`color: ${colors.link}!important`">{{ config.legalInfo }}</a>
          <span v-else>{{ config.legalInfo }}</span>
        </v-flex>
        <v-flex v-else xs12 class="pa-0 body-2 text-selectable text-xs-center white--text text-sm-right sm6">
          <strong>{{ config.siteCaption ? config.siteCaption : config.siteTitle }}</strong>
        </v-flex>
      </v-layout>
    </footer>
    <footer v-else>
      <v-layout wrap align-top pa-0 ma-0>
        <v-flex xs12 sm6 class="pa-0 body-2 text-xs-center text-sm-left white--text text-selectable">
          <strong>{{ config.siteTitle }}</strong> – {{ config.siteCaption }}
        </v-flex>
        <v-flex xs12 sm6 class="pa-0 body-2 text-xs-center text-sm-right white--text">
          <v-btn
              href="https://photoprism.app/"
              target="_blank"
              color="transparent"
              class="white--text px-3 py-2 ma-0 action-about"
              round depressed small
          >
            <translate>Learn more</translate>
            <v-icon :left="rtl" :right="!rtl" size="16" class="ml-2" dark>diamond</v-icon>
          </v-btn>
        </v-flex>
      </v-layout>
    </footer>
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
      showPassword: false,
      username: "",
      password: "",
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
    }
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

      setTimeout(() => { window.location = route.href; }, 100);
    },
    register() {
      window.location = this.registerUri;
    },
    login() {
      const username = this.username.trim();
      const password = this.password.trim();

      if (username === "" || password === "") {
        return;
      }

      this.loading = true;
      this.$session.login(username, password).then(
        () => {
          this.load();
        }
      ).catch(() => this.loading = false);
    },
  }
};
</script>
