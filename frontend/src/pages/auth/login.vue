<template>
  <v-container fluid fill-height class="auth-login wallpaper">
    <v-layout align-center justify-center>
      <v-flex xs12 sm8 md4>
          <v-card class="elevation-12 auth-login-box">
            <v-toolbar flat dark prominent :dense="false">
              <v-list class="navigation-home">
                <v-list-tile class="nav-logo">
                  <v-list-tile-avatar>
                    <img :src="$config.appIcon()" :alt="config.name">
                  </v-list-tile-avatar>
                  <v-list-tile-content>
                    <v-list-tile-title class="title">
                      PhotoPrism
                    </v-list-tile-title>
                  </v-list-tile-content>
                </v-list-tile>
              </v-list>
            </v-toolbar>
            <v-card-text class="pa-3">
              <v-form ref="form" dense autocomplete="off" class="p-form-login" accept-charset="UTF-8" @submit.prevent="login">
                <v-text-field
                    v-model="username"
                    required hide-details solo flat
                    type="text"
                    :disabled="loading"
                    :label="$gettext('Name')"
                    browser-autocomplete="off"
                    color="secondary-dark"
                    background-color="secondary-light"
                    class="input-name"
                    :placeholder="$gettext('Name')"
                    prepend-icon="person"
                ></v-text-field>
                <v-text-field
                    v-model="password"
                    required hide-details solo flat
                    :type="showPassword ? 'text' : 'password'"
                    :disabled="loading"
                    :label="$gettext('Password')"
                    browser-autocomplete="off"
                    color="secondary-dark"
                    background-color="secondary-light"
                    :placeholder="$gettext('Password')"
                    class="input-password mt-1"
                    :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                    prepend-icon="lock"
                    @click:append="showPassword = !showPassword"
                    @keyup.enter.native="login"
                ></v-text-field>
              </v-form>
            </v-card-text>
            <v-card-actions class="pa-3">
              <v-spacer></v-spacer>
              <v-btn color="primary-button" :disabled="loading || !password || !username"
                     class="white--text action-confirm" @click.stop="login">
                <translate>Sign in</translate>
                <v-icon :right="!rtl" :left="rtl" dark>login</v-icon>
              </v-btn>
            </v-card-actions>
          </v-card>
      </v-flex>
    </v-layout>

    <!-- p-about-footer v-if="!sponsor"></p-about-footer -->
  </v-container>
</template>

<script>

export default {
  name: "PPageAuthLogin",
  data() {
    const c = this.$config.values;

    return {
      loading: false,
      showPassword: false,
      username: "",
      password: "",
      sponsor: this.$config.isSponsor(),
      config: this.$config.values,
      siteDescription: c.siteDescription ? c.siteDescription : c.siteCaption,
      nextUrl: this.$route.params.nextUrl ? this.$route.params.nextUrl : "/",
      rtl: this.$rtl,
    };
  },
  created() {
    this.$scrollbar.hide();
  },
  destroyed() {
    this.$scrollbar.show();
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
  }
};
</script>
