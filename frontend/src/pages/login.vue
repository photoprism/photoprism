<template>
  <div class="p-page p-page-login">
    <v-toolbar flat color="secondary" dense class="mb-3" :height="42">
      <v-toolbar-title class="subheading">
        {{ siteDescription }}
      </v-toolbar-title>
    </v-toolbar>
    <v-form ref="form" dense autocomplete="off" class="p-form-login" accept-charset="UTF-8" @submit.prevent="login">
      <v-card flat tile class="ma-2 application">
        <v-card-actions>
          <v-layout wrap align-top>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="username"
                  required hide-details
                  type="text"
                  :disabled="loading"
                  :label="$gettext('Name')"
                  browser-autocomplete="off"
                  color="secondary-dark"
                  class="input-name"
                  placeholder="username"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="pa-2">
              <v-text-field
                  v-model="password"
                  required hide-details
                  :type="showPassword ? 'text' : 'password'"
                  :disabled="loading"
                  :label="$gettext('Password')"
                  browser-autocomplete="off"
                  color="secondary-dark"
                  placeholder="••••••••"
                  class="input-password"
                  :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter.native="login"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 class="px-2 py-3">
              <v-btn color="primary-button"
                     class="white--text ml-0 action-confirm"
                     depressed
                     :disabled="loading || !password || !username"
                     @click.stop="login">
                <translate>Sign in</translate>
                <v-icon :right="!rtl" :left="rtl" dark>login</v-icon>
              </v-btn>
              <v-btn color="primary-button"
                     class="white--text ml-0 action-confirm"
                     depressed
                     :disabled="loading"
                     @click.stop="loginExternal"
                     v-if="!!authProvider" >
                <translate>Sign in with {{ authProvider }}</translate>
                <v-icon :right="!rtl" :left="rtl" dark>login</v-icon>
              </v-btn>
            <template v-if="authProvider">
              <p class="mt-3">⚠️ Sign in with local user accounts except admin is disabled when using OpenID Connect.</p>
            </template>
            </v-flex>
          </v-layout>
        </v-card-actions>
      </v-card>
    </v-form>

    <p-about-footer></p-about-footer>
  </div>
</template>

<script>
import Notify from "../common/notify";
import axios from "axios";

export default {
  name: 'Login',
  data() {
    const c = this.$config.values;
    return {
      loading: false,
      showPassword: false,
      username: "",
      password: "",
      siteDescription: c.siteDescription ? c.siteDescription : c.siteCaption,
      authProvider: c.oidc ? "OpenID Connect" : null,
      nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
      rtl: this.$rtl,
    };
  },
  created() {
    const c = window.__CONFIG__;
    const preventAutoLogin = sessionStorage.getItem("preventAutoLogin");
    sessionStorage.removeItem("preventAutoLogin");
    if (!c.oidc || this.$route.query.preventAutoLogin || preventAutoLogin) {
      return;
    }
    const cleanup = () => {
      window.localStorage.removeItem('config');
      window.localStorage.removeItem('auth_error');
    };
    const redirect = () => {
      // check if oidc provider is available
      axios.get(c.oidc,{ timeout: 1000}).then(response => {
        // redirect to oidc provider
        window.location.href = '/api/v1/auth/external';
      }).catch(error => {
        if (c.debug) console.log(error);
      });
    };
    const externalLogin = this.onExternalLogin(cleanup, redirect);
    externalLogin();
  },
  methods: {
    login() {
      if (!this.username || !this.password) {
        return;
      }

      this.loading = true;
      this.$session.login(this.username, this.password).then(
        () => {
          this.loading = false;
          this.$router.push(this.nextUrl);
        }
      ).catch(() => this.loading = false);
    },
    loginExternal() {
      let popup = window.open('api/v1/auth/external', "external-login");
      const onstorage = window.onstorage;
      const cleanup = () => {
        window.localStorage.removeItem('config');
        window.localStorage.removeItem('auth_error');
        window.onstorage = onstorage;
        popup.close();
      };

      window.onstorage = this.onExternalLogin(cleanup);
    },
    onExternalLogin(cleanup, redirect) {
      return () => {
        const sid = window.localStorage.getItem('session_id');
        const data = window.localStorage.getItem('data');
        const config = window.localStorage.getItem('config');
        const error = window.localStorage.getItem('auth_error');

        if (error !== null) {
          console.log(error);
          cleanup();
          Notify.error(`${error}`);
          return;
        }
        if (sid === null || data === null || config === null) {
          if (typeof redirect === 'function') redirect();
          return;
        }
        this.$session.setId(sid);
        this.$session.setData(JSON.parse(data));
        this.$session.setConfig(JSON.parse(config));
        //this.$session.sendClientInfo();
        this.$router.push(this.nextUrl);
        cleanup();
      };
    },
  },
};
</script>
